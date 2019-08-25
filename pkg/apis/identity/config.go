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

package identity

import "yunion.io/x/jsonutils"

const (
	QueryScopeOne = "one"
	QUeryScopeSub = "sub"
)

type TIdentityProviderConfigs map[string]map[string]jsonutils.JSONObject

type SLDAPIdpConfigBaseOptions struct {
	Url      string `json:"url,omitempty" help:"LDAP server URL" required:"true"`
	Suffix   string `json:"suffix,omitempty" required:"true"`
	User     string `json:"user,omitempty" required:"true"`
	Password string `json:"password,omitempty" required:"true"`
}

type SLDAPIdpConfigSingleDomainOptions struct {
	SLDAPIdpConfigBaseOptions

	UserTreeDN  string `json:"user_tree_dn,omitempty" help:"Base user tree distinguished name" required:"true"`
	GroupTreeDN string `json:"group_tree_dn,omitempty" help:"Base group tree distinguished name" required:"true"`
}

type SLDAPIdpConfigMultiDomainOptions struct {
	SLDAPIdpConfigBaseOptions

	DomainTreeDN string `json:"domain_tree_dn,omitempty" help:"Base domain tree distinguished name" required:"true"`
}

type SLDAPIdpConfigOptions struct {
	Url        string `json:"url,omitempty" help:"LDAP server URL" required:"true"`
	Suffix     string `json:"suffix,omitempty" required:"true"`
	QueryScope string `json:"query_scope,omitempty" help:"Query scope" choices:"one|sub"`

	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`

	DomainTreeDN        string `json:"domain_tree_dn,omitempty" help:"Domain tree root node dn(distinguished name)"`
	DomainFilter        string `json:"domain_filter,omitempty"`
	DomainObjectclass   string `json:"domain_objectclass,omitempty"`
	DomainIdAttribute   string `json:"domain_id_attribute,omitempty"`
	DomainNameAttribute string `json:"domain_name_attribute,omitempty"`
	DomainQueryScope    string `json:"domain_query_scope,omitempty" help:"Query scope" choices:"one|sub"`

	UserTreeDN              string   `json:"user_tree_dn,omitempty" help:"User tree distinguished name"`
	UserFilter              string   `json:"user_filter,omitempty"`
	UserObjectclass         string   `json:"user_objectclass,omitempty"`
	UserIdAttribute         string   `json:"user_id_attribute,omitempty"`
	UserNameAttribute       string   `json:"user_name_attribute,omitempty"`
	UserEnabledAttribute    string   `json:"user_enabled_attribute,omitempty"`
	UserEnabledMask         int64    `json:"user_enabled_mask,allowzero" default:"-1"`
	UserEnabledDefault      string   `json:"user_enabled_default,omitempty"`
	UserEnabledInvert       bool     `json:"user_enabled_invert,allowfalse"`
	UserAdditionalAttribute []string `json:"user_additional_attribute_mapping,omitempty" token:"user_additional_attribute"`
	UserQueryScope          string   `json:"user_query_scope,omitempty" help:"Query scope" choices:"one|sub"`

	GroupTreeDN          string `json:"group_tree_dn,omitempty" help:"Group tree distinguished name"`
	GroupFilter          string `json:"group_filter,omitempty"`
	GroupObjectclass     string `json:"group_objectclass,omitempty"`
	GroupIdAttribute     string `json:"group_id_attribute,omitempty"`
	GroupNameAttribute   string `json:"group_name_attribute,omitempty"`
	GroupMemberAttribute string `json:"group_member_attribute,omitempty"`
	GroupMembersAreIds   bool   `json:"group_members_are_ids,allowfalse"`
	GroupQueryScope      string `json:"group_query_scope,omitempty" help:"Query scope" choices:"one|sub"`
}

const (
	IdpTemplateMSSingleDomain       = "msad_one_domain"
	IdpTemplateMSMultiDomain        = "msad_multi_domain"
	IdpTemplateOpenLDAPSingleDomain = "openldap_one_domain"
)

var (
	IdpTemplateDriver = map[string]string{
		IdpTemplateMSSingleDomain:       IdentityDriverLDAP,
		IdpTemplateMSMultiDomain:        IdentityDriverLDAP,
		IdpTemplateOpenLDAPSingleDomain: IdentityDriverLDAP,
	}
)
