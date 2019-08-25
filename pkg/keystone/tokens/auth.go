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

package tokens

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/utils"
	"yunion.io/x/sqlchemy"

	api "yunion.io/x/onecloud/pkg/apis/identity"
	"yunion.io/x/onecloud/pkg/keystone/driver"
	"yunion.io/x/onecloud/pkg/keystone/models"
	"yunion.io/x/onecloud/pkg/keystone/options"
	"yunion.io/x/onecloud/pkg/mcclient"
)

func authUserByTokenV2(ctx context.Context, input mcclient.SAuthenticationInputV2) (*api.SUserExtended, error) {
	return authUserByToken(ctx, input.Auth.Token.Id)
}

func authUserByTokenV3(ctx context.Context, input mcclient.SAuthenticationInputV3) (*api.SUserExtended, error) {
	return authUserByToken(ctx, input.Auth.Identity.Token.Id)
}

func authUserByToken(ctx context.Context, tokenStr string) (*api.SUserExtended, error) {
	token := SAuthToken{}
	err := token.ParseFernetToken(tokenStr)
	if err != nil {
		return nil, errors.Wrap(err, "token.ParseFernetToken")
	}
	return models.UserManager.FetchUserExtended(token.UserId, "", "", "")
}

func authUserByPasswordV2(ctx context.Context, input mcclient.SAuthenticationInputV2) (*api.SUserExtended, error) {
	ident := mcclient.SAuthenticationIdentity{}
	ident.Methods = []string{api.AUTH_METHOD_PASSWORD}
	ident.Password.User.Name = input.Auth.PasswordCredentials.Username
	ident.Password.User.Password = input.Auth.PasswordCredentials.Password
	ident.Password.User.Domain.Id = api.DEFAULT_DOMAIN_ID
	return authUserByIdentity(ctx, ident)
}

func authUserByIdentityV3(ctx context.Context, input mcclient.SAuthenticationInputV3) (*api.SUserExtended, error) {
	return authUserByIdentity(ctx, input.Auth.Identity)
}

func authUserByIdentity(ctx context.Context, ident mcclient.SAuthenticationIdentity) (*api.SUserExtended, error) {
	var idpId string

	if len(ident.Password.User.Name) == 0 && len(ident.Password.User.Id) == 0 && len(ident.Password.User.Domain.Id) == 0 && len(ident.Password.User.Domain.Name) == 0 {
		return nil, ErrEmptyAuth
	}
	if len(ident.Password.User.Name) > 0 && len(ident.Password.User.Id) == 0 && len(ident.Password.User.Domain.Id) == 0 && len(ident.Password.User.Domain.Name) == 0 {
		q := models.UserManager.Query().Equals("name", ident.Password.User.Name)
		usrCnt, err := q.CountWithError()
		if err != nil {
			return nil, errors.Wrap(err, "Query user by name")
		}
		if usrCnt > 1 {
			return nil, sqlchemy.ErrDuplicateEntry
		} else if usrCnt == 0 {
			/*idp, err := models.IdentityProviderManager.GetAutoCreateUserProvider()
			if err != nil {
				return nil, errors.Wrap(err, "IdentityProviderManager.GetAutoCreateUserProvider")
			}
			idpId = idp.Id
			*/
			return nil, sqlchemy.ErrEmptyQuery
		} else {
			// userCnt == 1
			usr := models.SUser{}
			usr.SetModelManager(models.UserManager, &usr)
			err := q.First(&usr)
			if err != nil {
				return nil, errors.Wrap(err, "Query user")
			}
			ident.Password.User.Domain.Id = usr.DomainId
			idmap, err := models.IdmappingManager.FetchEntity(usr.Id, api.IdMappingEntityUser)
			if err != nil && err != sql.ErrNoRows {
				return nil, errors.Wrap(err, "IdmappingManager.FetchEntity")
			}
			if idmap == nil { // sql
				idpId = api.DEFAULT_IDP_ID
			} else {
				idpId = idmap.IdpId
			}
		}
	} else {
		usrExt, err := models.UserManager.FetchUserExtended(ident.Password.User.Id, ident.Password.User.Name,
			ident.Password.User.Domain.Id, ident.Password.User.Domain.Name)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "UserManager.FetchUserExtended")
		}

		if err == sql.ErrNoRows {
			// no such user locally, query domain idp
			domain, err := models.DomainManager.FetchDomain(ident.Password.User.Domain.Id, ident.Password.User.Domain.Name)
			if err != nil {
				return nil, errors.Wrap(err, "DomainManager.FetchDomain")
			}
			mapping, err := models.IdmappingManager.FetchEntity(domain.Id, api.IdMappingEntityDomain)
			if err != nil {
				return nil, errors.Wrap(err, "IdmappingManager.FetchEntity")
			}
			idpId = mapping.IdpId
		} else {
			// user exists, query user's idp
			idpId = usrExt.IdpId
		}
	}

	if len(idpId) == 0 {
		idpId = api.DEFAULT_IDP_ID
	}
	idpObj, err := models.IdentityProviderManager.FetchById(idpId)
	if err != nil {
		return nil, errors.Wrap(err, "IdentityProviderManager.FetchById")
	}

	idp := idpObj.(*models.SIdentityProvider)

	if idp.Status != api.IdentityDriverStatusConnected && idp.Status != api.IdentityDriverStatusDisconnected {
		return nil, errors.Error(fmt.Sprintf("invalid idp status %s", idp.Status))
	}

	conf, err := idp.GetConfig(true)
	if err != nil {
		return nil, errors.Wrap(err, "GetConfig")
	}

	backend, err := driver.GetDriver(idp.Driver, idp.Id, idp.Name, idp.Template, idp.TargetDomainId, idp.AutoCreateProject.Bool(), conf)
	if err != nil {
		return nil, errors.Wrap(err, "driver.GetDriver")
	}

	usr, err := backend.Authenticate(ctx, ident)
	if err != nil {
		return nil, errors.Wrap(err, "Authenticate")
	}

	if idp.Status == api.IdentityDriverStatusDisconnected {
		idp.MarkConnected(ctx, models.GetDefaultAdminCred())
	}

	return usr, nil
}

func AuthenticateV3(ctx context.Context, input mcclient.SAuthenticationInputV3) (*mcclient.TokenCredentialV3, error) {
	var user *api.SUserExtended
	var err error
	if len(input.Auth.Identity.Methods) != 1 {
		return nil, ErrInvalidAuthMethod
	}
	method := input.Auth.Identity.Methods[0]
	if method == api.AUTH_METHOD_TOKEN {
		// auth by token
		user, err = authUserByTokenV3(ctx, input)
		if err != nil {
			return nil, errors.Wrap(err, "authUserByTokenV3")
		}
	} else {
		// auth by other methods, password, openid, saml, etc...
		user, err = authUserByIdentityV3(ctx, input)
		if err != nil {
			return nil, errors.Wrap(err, "authUserByIdentityV3")
		}
	}
	// user not found
	if user == nil {
		return nil, ErrUserNotFound
	}
	// user is not enabled
	if !user.Enabled {
		return nil, ErrUserDisabled
	}

	if !user.DomainEnabled {
		return nil, ErrDomainDisabled
	}

	token := SAuthToken{}
	token.UserId = user.Id
	token.Method = method
	token.AuditIds = []string{utils.GenRequestId(16)}
	now := time.Now().UTC()
	token.ExpiresAt = now.Add(time.Duration(options.Options.TokenExpirationSeconds) * time.Second)
	token.Context = input.Auth.Context

	if len(input.Auth.Scope.Project.Id) == 0 && len(input.Auth.Scope.Project.Name) == 0 && len(input.Auth.Scope.Domain.Id) == 0 && len(input.Auth.Scope.Domain.Name) == 0 {
		// unscoped auth
		return token.getTokenV3(ctx, user, nil, nil)
	}
	var projExt *models.SProjectExtended
	var domain *models.SDomain
	if len(input.Auth.Scope.Project.Id) > 0 || len(input.Auth.Scope.Project.Name) > 0 {
		project, err := models.ProjectManager.FetchProject(
			input.Auth.Scope.Project.Id,
			input.Auth.Scope.Project.Name,
			input.Auth.Scope.Project.Domain.Id,
			input.Auth.Scope.Project.Domain.Name,
		)
		if err != nil {
			return nil, errors.Wrap(err, "ProjectManager.FetchProject")
		}
		// if project.Enabled.IsFalse() {
		// 	return nil, ErrProjectDisabled
		// }
		projExt, err = project.FetchExtend()
		if err != nil {
			return nil, errors.Wrap(err, "project.FetchExtend")
		}
		token.ProjectId = project.Id
	} else {
		domain, err = models.DomainManager.FetchDomain(input.Auth.Scope.Domain.Id,
			input.Auth.Scope.Domain.Name)
		if err != nil {
			return nil, errors.Wrap(err, "DomainManager.FetchDomain")
		}
		if domain.Enabled.IsFalse() {
			return nil, ErrDomainDisabled
		}
		token.DomainId = domain.Id
	}
	return token.getTokenV3(ctx, user, projExt, domain)
}

func AuthenticateV2(ctx context.Context, input mcclient.SAuthenticationInputV2) (*mcclient.TokenCredentialV2, error) {
	var user *api.SUserExtended
	var err error
	var method string
	if len(input.Auth.Token.Id) > 0 {
		// auth by token
		user, err = authUserByTokenV2(ctx, input)
		if err != nil {
			return nil, errors.Wrap(err, "authUserByTokenV2")
		}
		method = api.AUTH_METHOD_TOKEN
	} else {
		// auth by password
		user, err = authUserByPasswordV2(ctx, input)
		if err != nil {
			return nil, errors.Wrap(err, "authUserByPasswordV2")
		}
		method = api.AUTH_METHOD_PASSWORD
	}
	// user not found
	if user == nil {
		return nil, ErrUserNotFound
	}
	// user is not enabled
	if !user.Enabled {
		return nil, ErrUserDisabled
	}

	if !user.DomainEnabled {
		return nil, ErrDomainDisabled
	}

	token := SAuthToken{}
	token.UserId = user.Id
	token.Method = method
	token.AuditIds = []string{utils.GenRequestId(16)}
	now := time.Now().UTC()
	token.ExpiresAt = now.Add(time.Duration(options.Options.TokenExpirationSeconds) * time.Second)
	token.Context = input.Auth.Context

	if len(input.Auth.TenantId) == 0 && len(input.Auth.TenantName) == 0 {
		// unscoped auth
		return token.getTokenV2(ctx, user, nil)
	}
	project, err := models.ProjectManager.FetchProject(
		input.Auth.TenantId,
		input.Auth.TenantName,
		api.DEFAULT_DOMAIN_ID, "")
	if err != nil {
		return nil, errors.Wrap(err, "ProjectManager.FetchProject")
	}
	// if project.Enabled.IsFalse() {
	// 	return nil, ErrProjectDisabled
	// }
	token.ProjectId = project.Id
	projExt, err := project.FetchExtend()
	if err != nil {
		return nil, errors.Wrap(err, "project.FetchExtend")
	}

	return token.getTokenV2(ctx, user, projExt)
}
