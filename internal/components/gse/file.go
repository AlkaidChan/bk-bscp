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

// Package gse provides gse api client.
package gse

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-bscp/internal/components"
	"github.com/TencentBlueKing/bk-bscp/pkg/cc"
)

// TransferFileReq defines transfer file task request
type TransferFileReq struct {
	TimeOutSeconds int                `json:"timeout_seconds"`
	AutoMkdir      bool               `json:"auto_mkdir"`
	UploadSpeed    int                `json:"upload_speed"`
	DownloadSpeed  int                `json:"download_speed"`
	Tasks          []TransferFileTask `json:"tasks"`
}

// TransferFileTask defines transfer file task
type TransferFileTask struct {
	Source TransferFileSource `json:"source"`
	Target TransferFileTarget `json:"target"`
}

// TransferFileSource defines transfer file task source
type TransferFileSource struct {
	FileName string            `json:"file_name"`
	StoreDir string            `json:"store_dir"`
	Agent    TransferFileAgent `json:"agent"`
}

// TransferFileTarget defines transfer file task target
type TransferFileTarget struct {
	FileName string              `json:"file_name"`
	StoreDir string              `json:"store_dir"`
	Agents   []TransferFileAgent `json:"agents"`
}

// TransferFileAgent defines transfer file task agent
type TransferFileAgent struct {
	User          string `json:"user"`
	BkAgentID     string `json:"bk_agent_id"`
	BkContainerID string `json:"bk_container_id"`
}

// CommonTaskRespData defines gse common task response data
type CommonTaskRespData struct {
	Result CommonTaskRespResult `json:"result"`
}

// CommonTaskRespResult defines gse common task response result
type CommonTaskRespResult struct {
	TaskID string `json:"task_id"`
}

// TerminateTransferFileTaskReq defines terminate transfer file task request
type TerminateTransferFileTaskReq struct {
	Agents []TransferFileAgent `json:"agents"`
	TaskID string              `json:"task_id"`
}

// CreateTransferFileTask create sync transfer file task
func CreateTransferFileTask(ctx context.Context, sourceAgentID, sourceContainerID, sourceFileDir, sourceUser,
	filename string, targetFileDir string, targetsAgents []TransferFileAgent) (string, error) {

	// 1. if sourceContainerID is set, means source is node, else is container
	// 2. if targetContainerID is set, means target is node, else is container

	url := fmt.Sprintf("%s/api/v2/task/extensions/async_transfer_file", cc.FeedServer().GSE.Host)
	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		cc.FeedServer().Esb.AppCode, cc.FeedServer().Esb.AppSecret)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		SetBody(TransferFileReq{
			TimeOutSeconds: 600,
			AutoMkdir:      true,
			UploadSpeed:    0,
			DownloadSpeed:  0,
			Tasks: []TransferFileTask{
				{
					Source: TransferFileSource{
						FileName: filename,
						StoreDir: sourceFileDir,
						Agent: TransferFileAgent{
							User:          sourceUser,
							BkAgentID:     sourceAgentID,
							BkContainerID: sourceContainerID,
						},
					},
					Target: TransferFileTarget{
						FileName: filename,
						StoreDir: targetFileDir,
						Agents:   targetsAgents,
					},
				},
			},
		}).
		Post(url)

	if err != nil {
		return "", err
	}

	data := &CommonTaskRespData{}
	if err := components.UnmarshalBKResult(resp, data); err != nil {
		return "", err
	}

	return data.Result.TaskID, nil
}

// TransferFileResultData defines transfer file task result data
type TransferFileResultData struct {
	Version string                         `json:"version"`
	Result  []TransferFileResultDataResult `json:"result"`
}

// TransferFileResultDataResult defines transfer file task result data result
type TransferFileResultDataResult struct {
	Content   TransferFileResultDataResultContent `json:"content"`
	ErrorCode int                                 `json:"error_code"`
	ErrorMsg  string                              `json:"error_msg"`
}

// TransferFileResultDataResultContent defines transfer file task result data result content
type TransferFileResultDataResultContent struct {
	DestAgentID       string `json:"dest_agent_id"`
	DestContainerID   string `json:"dest_container_id"`
	DestFileDir       string `json:"dest_file_dir"`
	DestFileName      string `json:"dest_file_name"`
	Mode              int    `json:"mode"`
	Progress          int    `json:"progress"`
	SourceAgentID     string `json:"source_agent_id"`
	SourceContainerID string `json:"source_container_id"`
	SourceFileDir     string `json:"source_file_dir"`
	SourceFileName    string `json:"source_file_name"`
	Speed             int    `json:"speed"`
	Status            int    `json:"status"`
	StatusInfo        string `json:"status_info"`
	Type              string `json:"type"`
	StartTime         int64  `json:"start_time"`
	EndTime           int64  `json:"end_time"`
	Size              int64  `json:"size"`
}

// TransferFileResult query transfer file task result
func TransferFileResult(ctx context.Context, taskID string) ([]TransferFileResultDataResult, error) {

	url := fmt.Sprintf("%s/api/v2/task/extensions/get_transfer_file_result", cc.FeedServer().GSE.Host)
	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		cc.FeedServer().Esb.AppCode, cc.FeedServer().Esb.AppSecret)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		SetBody(map[string]interface{}{
			"task_id": taskID,
		}).
		Post(url)

	if err != nil {
		return nil, err
	}

	data := &TransferFileResultData{}
	if err := components.UnmarshalBKResult(resp, data); err != nil {
		return nil, err
	}

	return data.Result, nil
}

// TerminateTransferFileTask terminate transfer file task
func TerminateTransferFileTask(ctx context.Context, taskID string, targetsAgents []TransferFileAgent) (string, error) {
	url := fmt.Sprintf("%s/api/v2/task/extensions/async_terminate_transfer_file", cc.FeedServer().GSE.Host)
	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		cc.FeedServer().Esb.AppCode, cc.FeedServer().Esb.AppSecret)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		SetBody(TerminateTransferFileTaskReq{
			TaskID: taskID,
			Agents: targetsAgents,
		}).
		Post(url)

	if err != nil {
		return "", err
	}

	data := &CommonTaskRespData{}
	if err := components.UnmarshalBKResult(resp, data); err != nil {
		return "", err
	}

	return data.Result.TaskID, nil
}
