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

package dao

import (
	"errors"
	"fmt"
	"strings"

	rawgen "gorm.io/gen"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bscp/internal/criteria/constant"
	"github.com/TencentBlueKing/bk-bscp/internal/dal/gen"
	"github.com/TencentBlueKing/bk-bscp/internal/dal/utils"
	"github.com/TencentBlueKing/bk-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bscp/pkg/types"
)

// AppTemplateBinding supplies all the app template binding related operations.
type AppTemplateBinding interface {
	// CreateWithTx create one app template binding instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, atb *table.AppTemplateBinding) (uint32, error)
	// Update one app template binding's info.
	Update(kit *kit.Kit, atb *table.AppTemplateBinding, removeTemplateSetId uint32) error
	// UpdateWithTx Update one app template binding's info with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, atb *table.AppTemplateBinding) error
	// BatchUpdateWithTx batch update app template binding's instances with transaction.
	BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.AppTemplateBinding) error
	// List app template bindings with options.
	List(kit *kit.Kit, bizID, appID uint32, opt *types.BasePage) ([]*table.AppTemplateBinding, int64, error)
	// Delete one app template binding instance.
	Delete(kit *kit.Kit, atb *table.AppTemplateBinding) error
	// DeleteByAppIDWithTx delete one app template binding instance by app id with transaction.
	DeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32) error
	// ListAppTemplateBindingByAppIds 按 AppId 列出应用模板绑定
	ListAppTemplateBindingByAppIds(kit *kit.Kit, bizID uint32, appIDs []uint32) ([]*table.AppTemplateBinding, error)
	// GetAppTemplateBindingByAppID 通过业务和服务ID获取模板绑定关系
	GetAppTemplateBindingByAppID(kit *kit.Kit, bizID, appID uint32) (*table.AppTemplateBinding, error)
	// UpsertWithTx create or update one template variable instance with transaction.
	UpsertWithTx(kit *kit.Kit, tx *gen.QueryTx, atb *table.AppTemplateBinding) error
	// GetBindingAppByTemplateSetID 通过套餐ID获取绑定的服务
	GetBindingAppByTemplateSetID(kit *kit.Kit, bizID uint32, templateSetIDs []uint32) ([]*table.AppTemplateBinding, error)
	// GetAppTemplateBindingByAppIDWithTx get app template binding with transaction instance through biz id and app id
	GetAppTemplateBindingByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID,
		appID uint32) (*table.AppTemplateBinding, error)
}

var _ AppTemplateBinding = new(appTemplateBindingDao)

type appTemplateBindingDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// GetAppTemplateBindingByAppIDWithTx get app template binding with transaction instance through biz id and app id
func (dao *appTemplateBindingDao) GetAppTemplateBindingByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32,
	appID uint32) (*table.AppTemplateBinding, error) {

	m := dao.genQ.AppTemplateBinding

	return tx.AppTemplateBinding.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Take()
}

// GetBindingAppByTemplateSetID 通过套餐ID获取绑定的服务
func (dao *appTemplateBindingDao) GetBindingAppByTemplateSetID(kit *kit.Kit, bizID uint32, templateSetIDs []uint32) (
	[]*table.AppTemplateBinding, error) {

	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID))

	conds := make([]rawgen.Condition, 0)
	for k, v := range templateSetIDs {
		if k == 0 {
			conds = append(conds, q.Where(utils.RawCond("JSON_CONTAINS(?,?)", utils.Field{
				Field: m.TemplateSetIDs,
			}, fmt.Sprintf("%d", v))))
		} else {
			conds = append(conds, q.Or(utils.RawCond("JSON_CONTAINS(?,?)", utils.Field{
				Field: m.TemplateSetIDs,
			}, fmt.Sprintf("%d", v))))
		}
	}

	return q.Where(conds...).Find()
}

// BatchUpdateWithTx batch update app template binding's instances with transaction.
func (dao *appTemplateBindingDao) BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	data []*table.AppTemplateBinding) error {
	if len(data) == 0 {
		return nil
	}
	for _, g := range data {
		if err := g.ValidateUpdate(); err != nil {
			return err
		}
		if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
			return err
		}
	}
	return tx.AppTemplateBinding.WithContext(kit.Ctx).Save(data...)
}

// ListAppTemplateBindingByAppIds 按 AppId 列出应用模板绑定
func (dao *appTemplateBindingDao) ListAppTemplateBindingByAppIds(kit *kit.Kit, bizID uint32, appIDs []uint32) (
	[]*table.AppTemplateBinding, error) {

	m := dao.genQ.AppTemplateBinding
	return dao.genQ.AppTemplateBinding.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.In(appIDs...)).
		Find()
}

// UpsertWithTx create or update one template variable instance with transaction.
func (dao *appTemplateBindingDao) UpsertWithTx(kit *kit.Kit, tx *gen.QueryTx, atb *table.AppTemplateBinding) error {
	m := tx.AppTemplateBinding
	q := tx.AppTemplateBinding.WithContext(kit.Ctx)

	tSet := tx.TemplateSet
	tSpace := tx.TemplateSpace
	tssn := []types.TemplateSetSpaceName{}

	// 查询空间名称及套餐名称
	err := tSet.WithContext(kit.Ctx).Select(tSet.Name.As("template_set_name"), tSpace.Name.As("template_space_name")).
		Join(tSpace.WithContext(kit.Ctx), tSpace.ID.EqCol(tSet.TemplateSpaceID)).
		Where(tSet.ID.In(atb.Spec.TemplateSetIDs...)).Scan(&tssn)
	if err != nil {
		return err
	}

	var templateSpaceNames []string
	templateSpaceNamesM := make(map[string]struct{})
	var templateSetNames []string
	for _, v := range tssn {
		// 去重
		if _, ok := templateSpaceNamesM[v.TemplateSpaceName]; !ok {
			templateSpaceNames = append(templateSpaceNames, v.TemplateSpaceName)
			templateSpaceNamesM[v.TemplateSpaceName] = struct{}{}
		}
		templateSetNames = append(templateSetNames, v.TemplateSetName)
	}

	resInstance := fmt.Sprintf(constant.TemplateSpaceName+constant.ResSeparator+constant.TemplateSetName,
		strings.Join(templateSpaceNames, constant.NameSeparator),
		strings.Join(templateSetNames, constant.NameSeparator))

	old, findErr := q.Where(m.BizID.Eq(atb.Attachment.BizID), m.AppID.Eq(atb.Attachment.AppID)).Take()

	// if old exists, update it.
	if findErr == nil {
		atb.ID = old.ID
		if _, err = tx.AppTemplateBinding.WithContext(kit.Ctx).
			Where(m.BizID.Eq(atb.Attachment.BizID), m.ID.Eq(atb.ID)).
			Select(m.Bindings, m.TemplateSpaceIDs, m.TemplateSetIDs, m.TemplateIDs, m.TemplateRevisionIDs,
				m.LatestTemplateIDs, m.Creator, m.Reviser, m.UpdatedAt).
			Updates(atb); err != nil {
			return err
		}

		// audit
		ad := dao.auditDao.Decorator(kit, atb.Attachment.BizID, &table.AuditField{
			ResourceInstance: resInstance,
			Status:           enumor.Success,
			AppId:            atb.Attachment.AppID,
		}).PrepareUpdate(&table.ConfigItem{ID: atb.ID})
		if err = ad.Do(tx.Query); err != nil {
			return err
		}

	} else if errors.Is(findErr, gorm.ErrRecordNotFound) {
		// if old not exists, create it.
		id, err := dao.idGen.One(kit, table.Name(atb.TableName()))
		if err != nil {
			return err
		}
		atb.ID = id
		if err = tx.AppTemplateBinding.WithContext(kit.Ctx).Create(atb); err != nil {
			return err
		}
		// audit
		ad := dao.auditDao.Decorator(kit, atb.Attachment.BizID, &table.AuditField{
			ResourceInstance: resInstance,
			Status:           enumor.Success,
			AppId:            atb.Attachment.AppID,
		}).PrepareCreate(&table.AppTemplateBinding{ID: atb.ID})
		if err = ad.Do(tx.Query); err != nil {
			return err
		}
	}

	return nil
}

// GetAppTemplateBindingByAppID 通过业务和服务ID获取模板绑定关系
func (dao *appTemplateBindingDao) GetAppTemplateBindingByAppID(kit *kit.Kit, bizID uint32, appID uint32) (
	*table.AppTemplateBinding, error) {

	m := dao.genQ.AppTemplateBinding
	return dao.genQ.AppTemplateBinding.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Take()
}

// CreateWithTx create one app template binding instance with transaction.
func (dao *appTemplateBindingDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.AppTemplateBinding) (
	uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}
	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate a app template binding id and update to app template binding.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	q := tx.AppTemplateBinding.WithContext(kit.Ctx)
	if err = q.Create(g); err != nil {
		return 0, err
	}

	tSet := tx.TemplateSet
	tSpace := tx.TemplateSpace
	tssn := []types.TemplateSetSpaceName{}

	// 查询空间名称及套餐名称
	err = tSet.WithContext(kit.Ctx).Select(tSet.Name.As("template_set_name"), tSpace.Name.As("template_space_name")).
		Join(tSpace.WithContext(kit.Ctx), tSpace.ID.EqCol(tSet.TemplateSpaceID)).
		Where(tSet.ID.In(g.Spec.TemplateSetIDs...)).Scan(&tssn)
	if err != nil {
		return 0, err
	}

	var templateSpaceNames []string
	templateSpaceNamesM := make(map[string]struct{})
	var templateSetNames []string
	for _, v := range tssn {
		// 去重
		if _, ok := templateSpaceNamesM[v.TemplateSpaceName]; !ok {
			templateSpaceNames = append(templateSpaceNames, v.TemplateSpaceName)
			templateSpaceNamesM[v.TemplateSpaceName] = struct{}{}
		}
		templateSetNames = append(templateSetNames, v.TemplateSetName)
	}

	resInstance := fmt.Sprintf(constant.TemplateSpaceName+constant.ResSeparator+constant.TemplateSetName,
		strings.Join(templateSpaceNames, constant.NameSeparator),
		strings.Join(templateSetNames, constant.NameSeparator))
	// audit
	ad := dao.auditDao.Decorator(kit, g.Attachment.BizID, &table.AuditField{
		ResourceInstance: resInstance,
		Status:           enumor.Success,
		AppId:            g.Attachment.AppID,
	}).PrepareCreate(&table.AppTemplateBinding{ID: g.ID})

	if err = ad.Do(tx.Query); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Update one app template binding instance.
func (dao *appTemplateBindingDao) Update(kit *kit.Kit, g *table.AppTemplateBinding, removeTemplateSetId uint32) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}
	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return err
	}
	// 删除操作，当前update接口是被当作删除使用，获取当前记录做审计

	tSet := dao.genQ.TemplateSet
	tSpace := dao.genQ.TemplateSpace
	tssn := []types.TemplateSetSpaceName{}

	// 查询空间名称及套餐名称
	err := tSet.WithContext(kit.Ctx).Select(tSet.Name.As("template_set_name"), tSpace.Name.As("template_space_name")).
		Join(tSpace.WithContext(kit.Ctx), tSpace.ID.EqCol(tSet.TemplateSpaceID)).
		Where(tSet.ID.In(removeTemplateSetId)).Scan(&tssn)
	if err != nil {
		return err
	}

	var templateSpaceNames []string
	var templateSetNames []string
	for _, v := range tssn {
		// 无需去重
		templateSpaceNames = append(templateSpaceNames, v.TemplateSpaceName)
		templateSetNames = append(templateSetNames, v.TemplateSetName)
	}

	resInstance := fmt.Sprintf(constant.TemplateSpaceName+constant.ResSeparator+constant.TemplateSetName,
		strings.Join(templateSpaceNames, constant.NameSeparator),
		strings.Join(templateSetNames, constant.NameSeparator))

	// audit
	ad := dao.auditDao.Decorator(kit, g.Attachment.BizID, &table.AuditField{
		ResourceInstance: resInstance,
		Status:           enumor.Success,
		AppId:            g.Attachment.AppID,
	}).PrepareDelete(&table.AppTemplateBinding{ID: g.ID})

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		m := tx.AppTemplateBinding
		q := tx.AppTemplateBinding.WithContext(kit.Ctx)
		if _, err = q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Bindings, m.TemplateSpaceIDs, m.TemplateSetIDs, m.TemplateIDs, m.TemplateRevisionIDs,
				m.LatestTemplateIDs, m.Creator, m.Reviser, m.UpdatedAt).
			Updates(g); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genQ.Transaction(updateTx); err != nil {
		return err
	}

	return nil
}

// UpdateWithTx Update one app template binding's info with transaction.
func (dao *appTemplateBindingDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	g *table.AppTemplateBinding) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}
	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := tx.AppTemplateBinding
	q := tx.AppTemplateBinding.WithContext(kit.Ctx)

	if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
		Select(m.Bindings, m.TemplateSpaceIDs, m.TemplateSetIDs, m.TemplateIDs, m.TemplateRevisionIDs,
			m.LatestTemplateIDs, m.Creator, m.Reviser, m.UpdatedAt).
		Updates(g); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	tSet := dao.genQ.TemplateSet
	tSpace := dao.genQ.TemplateSpace
	tssn := []types.TemplateSetSpaceName{}

	// 查询空间名称及套餐名称
	err := tSet.WithContext(kit.Ctx).Select(tSet.Name.As("template_set_name"), tSpace.Name.As("template_space_name")).
		Join(tSpace.WithContext(kit.Ctx), tSpace.ID.EqCol(tSet.TemplateSpaceID)).
		Where(tSet.ID.In(g.Spec.TemplateSetIDs...)).Scan(&tssn)
	if err != nil {
		return err
	}

	var templateSpaceNames []string
	templateSpaceNamesM := make(map[string]struct{})
	var templateSetNames []string
	for _, v := range tssn {
		// 去重
		if _, ok := templateSpaceNamesM[v.TemplateSpaceName]; !ok {
			templateSpaceNames = append(templateSpaceNames, v.TemplateSpaceName)
			templateSpaceNamesM[v.TemplateSpaceName] = struct{}{}
		}
		templateSetNames = append(templateSetNames, v.TemplateSetName)
	}

	resInstance := fmt.Sprintf(constant.TemplateSpaceName+constant.ResSeparator+constant.TemplateSetName,
		strings.Join(templateSpaceNames, constant.NameSeparator),
		strings.Join(templateSetNames, constant.NameSeparator))

	// audit
	ad := dao.auditDao.Decorator(kit, g.Attachment.BizID, &table.AuditField{
		ResourceInstance: resInstance,
		Status:           enumor.Success,
		AppId:            g.Attachment.AppID,
	}).PrepareUpdate(&table.AppTemplateBinding{ID: g.ID})

	if err = ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}

// List app template bindings with options.
func (dao *appTemplateBindingDao) List(kit *kit.Kit, bizID, appID uint32,
	opt *types.BasePage) ([]*table.AppTemplateBinding, int64, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)

	d := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID))
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// Delete one app template binding instance.
func (dao *appTemplateBindingDao) Delete(kit *kit.Kit, g *table.AppTemplateBinding) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	// 更新操作, 获取当前记录做审计
	tSet := dao.genQ.TemplateSet
	tSpace := dao.genQ.TemplateSpace
	tssn := []types.TemplateSetSpaceName{}

	// 查询空间名称及套餐名称
	err = tSet.WithContext(kit.Ctx).Select(tSet.Name.As("template_set_name"), tSpace.Name.As("template_space_name")).
		Join(tSpace.WithContext(kit.Ctx), tSpace.ID.EqCol(tSet.TemplateSpaceID)).
		Where(tSet.ID.In(oldOne.Spec.TemplateSetIDs...)).Scan(&tssn)
	if err != nil {
		return err
	}

	var templateSpaceNames []string
	templateSpaceNamesM := make(map[string]struct{})
	var templateSetNames []string
	for _, v := range tssn {
		// 去重
		if _, ok := templateSpaceNamesM[v.TemplateSpaceName]; !ok {
			templateSpaceNames = append(templateSpaceNames, v.TemplateSpaceName)
			templateSpaceNamesM[v.TemplateSpaceName] = struct{}{}
		}
		templateSetNames = append(templateSetNames, v.TemplateSetName)
	}

	resInstance := fmt.Sprintf(constant.TemplateSpaceName+constant.ResSeparator+constant.TemplateSetName,
		strings.Join(templateSpaceNames, constant.NameSeparator),
		strings.Join(templateSetNames, constant.NameSeparator))

	// audit
	ad := dao.auditDao.Decorator(kit, g.Attachment.BizID, &table.AuditField{
		ResourceInstance: resInstance,
		Status:           enumor.Success,
		AppId:            oldOne.Attachment.AppID,
	}).PrepareDelete(&table.AppTemplateBinding{ID: g.ID})

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		m := tx.AppTemplateBinding
		q := tx.AppTemplateBinding.WithContext(kit.Ctx)
		if _, err = q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genQ.Transaction(deleteTx); err != nil {
		return err
	}

	return nil
}

// DeleteByAppIDWithTx delete one app template binding instance by app id with transaction.
func (dao *appTemplateBindingDao) DeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32) error {
	m := tx.AppTemplateBinding
	q := tx.AppTemplateBinding.WithContext(kit.Ctx)
	_, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Delete()
	return err
}

// validateAttachmentExist validate if attachment resource exists before operating template
func (dao *appTemplateBindingDao) validateAttachmentExist(kit *kit.Kit, am *table.AppTemplateBindingAttachment) error {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(am.AppID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template attached app %d is not exist", am.AppID)
		}
		return fmt.Errorf("get template attached app failed, err: %v", err)
	}

	return nil
}
