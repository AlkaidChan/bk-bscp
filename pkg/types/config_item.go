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

package types

import (
	"github.com/TencentBlueKing/bk-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bscp/pkg/runtime/filter"
)

// ListConfigItemsOption defines options to list config item.
type ListConfigItemsOption struct {
	BizID  uint32             `json:"biz_id"`
	AppID  uint32             `json:"app_id"`
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
}

// Validate the list config item options
func (opt *ListConfigItemsOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid app id, should >= 1")
	}

	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id,app_id because it's a required field in the option.
		RuleFields: table.ConfigItemColumns.WithoutColumn("biz_id", "app_id"),
	}
	if err := opt.Filter.Validate(exprOpt); err != nil {
		return err
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListConfigItemDetails defines the response details of requested ListConfigItemsOption.
type ListConfigItemDetails struct {
	Count   uint32              `json:"count"`
	Details []*table.ConfigItem `json:"details"`
}

// CIUniqueKey defines struct of unique key of config item.
type CIUniqueKey struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// ListConfigItemCount define the list configuration item statistical structure
type ListConfigItemCount struct {
	AppID uint32 `json:"app_id"`
	Count uint64 `json:"count"`
}
