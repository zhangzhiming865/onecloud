// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package modules

import (
	"yunion.io/x/onecloud/pkg/multicloud/huawei/client/auth"
)

type SNatSRuleManager struct {
	SResourceManager
}

func NewNatSManager(regionId string, projectId string, signer auth.Signer, debug bool) *SNatSRuleManager {
	return &SNatSRuleManager{SResourceManager: SResourceManager{
		SBaseManager:  NewBaseManager(signer, debug),
		ServiceName:   ServiceNameNAT,
		Region:        regionId,
		ProjectId:     "",
		version:       "v2.0",
		Keyword:       "snat_rule",
		KeywordPlural: "snat_rules",

		ResourceKeyword: "snat_rules",
	}}
}
