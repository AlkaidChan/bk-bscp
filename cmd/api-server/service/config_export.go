/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/TencentBlueKing/bk-bscp/internal/dal/repository"
	"github.com/TencentBlueKing/bk-bscp/internal/iam/auth"
	"github.com/TencentBlueKing/bk-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bscp/pkg/protocol/config-server"
	pbtr "github.com/TencentBlueKing/bk-bscp/pkg/protocol/core/template-revision"
	"github.com/TencentBlueKing/bk-bscp/pkg/rest"
)

func newConfigExportService(settings cc.Repository, authorizer auth.Authorizer,
	cfgClient pbcs.ConfigClient) (*configExport, error) {
	provider, err := repository.NewProvider(settings)
	if err != nil {
		return nil, err
	}
	config := &configExport{
		authorizer: authorizer,
		provider:   provider,
		cfgClient:  cfgClient,
	}
	return config, nil
}

type configExport struct {
	authorizer auth.Authorizer
	provider   repository.Provider
	cfgClient  pbcs.ConfigClient
}

type download struct {
	commitSpec     *table.CommitSpec
	configItemSpec *table.ConfigItemSpec
}

// ConfigFileExport 配置文件导出压缩包
func (c *configExport) ConfigFileExport(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())
	appIdStr := chi.URLParam(r, "app_id")
	appId, _ := strconv.Atoi(appIdStr)
	if appId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}
	kt.AppID = uint32(appId)
	releaseIDStr := chi.URLParam(r, "release_id")
	releaseID, _ := strconv.Atoi(releaseIDStr)

	// 验证非模板配置和模板配置是否存在冲突
	ci, err := c.cfgClient.ListConfigItems(kt.RpcCtx(), &pbcs.ListConfigItemsReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
		All:   true,
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	if ci.ConflictNumber > 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("create release failed there is a file conflict")))
		return
	}

	// 获取服务信息
	app, err := c.cfgClient.GetApp(kt.RpcCtx(), &pbcs.GetAppReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	var downloads []*download
	var fileName string
	if releaseID > 0 {
		// 获取发布信息
		release, errR := c.cfgClient.GetRelease(kt.RpcCtx(), &pbcs.GetReleaseReq{
			BizId:     kt.BizID,
			AppId:     kt.AppID,
			ReleaseId: uint32(releaseID),
		})
		if errR != nil {
			_ = render.Render(w, r, rest.BadRequest(errR))
			return
		}
		fileName = fmt.Sprintf("%s_%s.zip", app.GetSpec().Name, release.Spec.Name)
		downloads, err = c.getPublishedConfigItems(kt, uint32(releaseID))
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
	} else {
		fileName = fmt.Sprintf("%s.zip", app.GetSpec().Name)
		downloads, err = c.getUnPublishedConfigItems(kt)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
	}

	if len(downloads) == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("There are no files to download")))
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/zip")
	w.WriteHeader(http.StatusOK)

	// 创建 zip writer，将文件内容写入到 zip 文件中
	zipWriter := zip.NewWriter(w)
	defer func() { _ = zipWriter.Close() }()

	for _, file := range downloads {
		err := c.downloadFileToZip(kt, file, zipWriter)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(fmt.Errorf("failed to download files: %v", err)))
			return
		}
	}
}

// 下载文件且压缩成zip
func (c *configExport) downloadFileToZip(kt *kit.Kit, file *download, zipWriter *zip.Writer) error {
	body, contentLength, err := c.provider.Download(kt, file.commitSpec.Content.Signature)
	if err != nil {
		return err
	}

	defer body.Close()

	fileName := filepath.Join(file.configItemSpec.Path, file.configItemSpec.Name)
	trimmedPath := strings.TrimPrefix(fileName, "/")
	writer, err := zipWriter.Create(trimmedPath)
	if err != nil {
		return fmt.Errorf("Error creating ZIP file entry:%s", err.Error())
	}

	n, err := io.Copy(writer, body)
	if err != nil {
		return err
	}

	if n != contentLength {
		return errors.New("download failed file missing")
	}
	return nil
}

// 获取已发布的配置项
func (c *configExport) getPublishedConfigItems(kt *kit.Kit, releaseID uint32) ([]*download, error) {
	downloads := []*download{}
	// 获取非配置模板
	rci, err := c.cfgClient.ListReleasedConfigItems(kt.RpcCtx(), &pbcs.ListReleasedConfigItemsReq{
		BizId:     kt.BizID,
		AppId:     kt.AppID,
		ReleaseId: releaseID,
		All:       true,
	})
	if err != nil {
		return downloads, err
	}

	// 获取已发布的模板套餐
	rtci, err := c.cfgClient.ListReleasedAppBoundTmplRevisions(kt.RpcCtx(), &pbcs.ListReleasedAppBoundTmplRevisionsReq{
		BizId:     kt.BizID,
		AppId:     kt.AppID,
		ReleaseId: releaseID,
	})
	if err != nil {
		return downloads, err
	}
	for _, file := range rci.Details {
		downloads = append(downloads, &download{
			commitSpec: &table.CommitSpec{
				ContentID: file.CommitSpec.GetContentId(),
				Content: &table.ContentSpec{
					Signature: file.CommitSpec.Content.Signature,
					ByteSize:  file.CommitSpec.Content.ByteSize,
				},
			},
			configItemSpec: &table.ConfigItemSpec{
				Name: file.GetSpec().Name,
				Path: file.GetSpec().Path,
			},
		})
	}
	for _, v := range rtci.Details {
		for _, file := range v.TemplateRevisions {
			downloads = append(downloads, &download{
				commitSpec: &table.CommitSpec{
					Content: &table.ContentSpec{
						Signature: file.Signature,
						ByteSize:  file.ByteSize,
					},
				},
				configItemSpec: &table.ConfigItemSpec{
					Name: file.Name,
					Path: file.Path,
				},
			})
		}
	}

	return downloads, nil
}

// 获取未发布的配置项
func (c *configExport) getUnPublishedConfigItems(kt *kit.Kit) ([]*download, error) {
	downloads := []*download{}
	ci, err := c.cfgClient.ListConfigItems(kt.RpcCtx(), &pbcs.ListConfigItemsReq{
		BizId:      kt.BizID,
		AppId:      kt.AppID,
		All:        true,
		WithStatus: true,
		Status:     []string{constant.FileStateAdd, constant.FileStateRevise, constant.FileStateUnchange},
		Start:      0,
	})
	if err != nil {
		return downloads, err
	}
	for _, file := range ci.GetDetails() {
		downloads = append(downloads, &download{
			commitSpec: &table.CommitSpec{
				ContentID: file.CommitSpec.GetContentId(),
				Content: &table.ContentSpec{
					Signature: file.CommitSpec.Content.Signature,
					ByteSize:  file.CommitSpec.Content.ByteSize,
				},
			},
			configItemSpec: &table.ConfigItemSpec{
				Name: file.GetSpec().Name,
				Path: file.GetSpec().Path,
			},
		})
	}

	// 获取未发布的模板套餐
	tci, err := c.cfgClient.ListAppBoundTmplRevisions(kt.RpcCtx(), &pbcs.ListAppBoundTmplRevisionsReq{
		BizId:      kt.BizID,
		AppId:      kt.AppID,
		WithStatus: true,
		Status:     []string{constant.FileStateAdd, constant.FileStateRevise, constant.FileStateUnchange},
	})
	if err != nil {
		return downloads, err
	}
	for _, v := range tci.Details {
		for _, file := range v.TemplateRevisions {
			downloads = append(downloads, &download{
				commitSpec: &table.CommitSpec{
					Content: &table.ContentSpec{
						Signature: file.Signature,
						ByteSize:  file.ByteSize,
					},
				},
				configItemSpec: &table.ConfigItemSpec{
					Name: file.Name,
					Path: file.Path,
				},
			})
		}
	}

	return downloads, nil
}

// TemplateExport 模板导出
func (c *configExport) TemplateExport(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())
	templateSpaceId := chi.URLParam(r, "template_space_id")
	tsId, _ := strconv.Atoi(templateSpaceId)
	if tsId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}
	templateId := chi.URLParam(r, "template_id")
	tId, _ := strconv.Atoi(templateId)

	resp, err := c.cfgClient.GetLatestTemplateVersionsInSpace(kt.RpcCtx(), &pbcs.GetLatestTemplateVersionsInSpaceReq{
		BizId:           kt.BizID,
		TemplateSpaceId: uint32(tsId),
		TemplateId:      uint32(tId),
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	fileName := fmt.Sprintf("%s.zip", resp.GetTemplateSpace().GetName())

	if len(resp.GetTemplateSet()) == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("There are no files to download")))
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/zip")
	w.WriteHeader(http.StatusOK)
	// 创建 zip writer，将文件内容写入到 zip 文件中
	zipWriter := zip.NewWriter(w)
	defer func() { _ = zipWriter.Close() }()
	for _, file := range resp.GetTemplateSet() {
		for _, v := range file.TemplateRevision {
			err := c.downloadTmpFileToZip(kt, file.Name, v, zipWriter)
			if err != nil {
				_ = render.Render(w, r, rest.BadRequest(fmt.Errorf("failed to download files: %v", err)))
				return
			}
		}

	}
}

// 下载模板文件且压缩成zip
func (c *configExport) downloadTmpFileToZip(kt *kit.Kit, folderName string,
	revision *pbtr.TemplateRevisionSpec, zipWriter *zip.Writer) error {
	body, contentLength, err := c.provider.Download(kt, revision.ContentSpec.Signature)
	if err != nil {
		return err
	}

	defer body.Close()

	fileName := filepath.Join(folderName, revision.Path, revision.Name)
	trimmedPath := strings.TrimPrefix(fileName, "/")
	writer, err := zipWriter.Create(trimmedPath)
	if err != nil {
		return fmt.Errorf("Error creating ZIP file entry:%s", err.Error())
	}

	n, err := io.Copy(writer, body)
	if err != nil {
		return err
	}

	if n != contentLength {
		return errors.New("download failed file missing")
	}
	return nil
}
