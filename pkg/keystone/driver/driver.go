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

package driver

import (
	"context"

	api "yunion.io/x/onecloud/pkg/apis/identity"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type IIdentityBackendClass interface {
	SingletonInstance() bool
	SyncMethod() string
	Name() string
	NewDriver(idpId, idpName, template, targetDomainId string, autoCreateProject bool, conf api.TIdentityProviderConfigs) (IIdentityBackend, error)
}

type IIdentityBackend interface {
	Authenticate(ctx context.Context, identity mcclient.SAuthenticationIdentity) (*api.SUserExtended, error)
	Sync(ctx context.Context) error
	Probe(ctx context.Context) error
}
