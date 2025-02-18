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

package shell

import (
	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/mcclient/options"
)

func init() {
	type UserListOptions struct {
		options.BaseListOptions
		Name          string `help:"Filter by name"`
		OrderByDomain string `help:"order by domain name" choices:"asc|desc"`
		Role          string `help:"Filter by role"`
	}
	R(&UserListOptions{}, "user-list", "List users", func(s *mcclient.ClientSession, args *UserListOptions) error {
		params, err := options.ListStructToParams(args)
		if err != nil {
			return err
		}
		result, err := modules.UsersV3.List(s, params)
		if err != nil {
			return err
		}
		printList(result, modules.UsersV3.GetColumns(s))
		return nil
	})

	type UserDetailOptions struct {
		ID     string `help:"ID of user"`
		Domain string `help:"Domain"`
		System bool   `help:"show system user"`
	}
	R(&UserDetailOptions{}, "user-show", "Show details of user", func(s *mcclient.ClientSession, args *UserDetailOptions) error {
		query := jsonutils.NewDict()
		if len(args.Domain) > 0 {
			domainId, err := modules.Domains.GetId(s, args.Domain, nil)
			if err != nil {
				return err
			}
			query.Add(jsonutils.NewString(domainId), "domain_id")
		}
		if args.System {
			query.Add(jsonutils.JSONTrue, "system")
		}
		user, e := modules.UsersV3.Get(s, args.ID, query)
		if e != nil {
			return e
		}
		printObject(user)
		return nil
	})
	R(&UserDetailOptions{}, "user-delete", "Delete user", func(s *mcclient.ClientSession, args *UserDetailOptions) error {
		query := jsonutils.NewDict()
		if len(args.Domain) > 0 {
			domainId, err := modules.Domains.GetId(s, args.Domain, nil)
			if err != nil {
				return err
			}
			query.Add(jsonutils.NewString(domainId), "domain_id")
		}
		uid, e := modules.UsersV3.GetId(s, args.ID, query)
		if e != nil {
			return e
		}
		_, e = modules.UsersV3.Delete(s, uid, nil)
		if e != nil {
			return e
		}
		return nil
	})

	R(&UserDetailOptions{}, "user-project-list", "List projects of user", func(s *mcclient.ClientSession, args *UserDetailOptions) error {
		query := jsonutils.NewDict()
		if len(args.Domain) > 0 {
			domainId, err := modules.Domains.GetId(s, args.Domain, nil)
			if err != nil {
				return err
			}
			query.Add(jsonutils.NewString(domainId), "domain_id")
		}
		uid, err := modules.UsersV3.GetId(s, args.ID, query)
		if err != nil {
			return err
		}
		projects, e := modules.UsersV3.GetProjects(s, uid)
		if e != nil {
			return e
		}
		printList(projects, modules.Projects.GetColumns(s))
		return nil
	})

	R(&UserDetailOptions{}, "user-group-list", "List groups of user", func(s *mcclient.ClientSession, args *UserDetailOptions) error {
		query := jsonutils.NewDict()
		if len(args.Domain) > 0 {
			domainId, err := modules.Domains.GetId(s, args.Domain, nil)
			if err != nil {
				return err
			}
			query.Add(jsonutils.NewString(domainId), "domain_id")
		}
		uid, err := modules.UsersV3.GetId(s, args.ID, query)
		if err != nil {
			return err
		}
		groups, e := modules.UsersV3.GetGroups(s, uid)
		if e != nil {
			return e
		}
		printList(groups, modules.Groups.GetColumns(s))
		return nil
	})

	type UserTenantRoleOptions struct {
		ID     string `help:"ID of user"`
		Tenant string `help:"ID of tenant"`
	}
	R(&UserTenantRoleOptions{}, "user-role-list", "List roles of user", func(s *mcclient.ClientSession, args *UserTenantRoleOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.ID), "id")
		if len(args.Tenant) > 0 {
			params.Add(jsonutils.NewString(args.Tenant), "tenantId")
		}
		result, err := modules.Users.GetTenantRoleList(s, params)
		if err != nil {
			return err
		}
		printList(modules.JSON2ListResult(result), nil)
		return nil
	})

	type UserCreateOptions struct {
		NAME        string `help:"Name of the new user"`
		Domain      string `help:"Domain"`
		Desc        string `help:"Description"`
		Password    string `help:"Password"`
		Displayname string `help:"Displayname"`
		Email       string `help:"Email"`
		Mobile      string `help:"Mobile"`
		Enabled     bool   `help:"Enabled"`
		Disabled    bool   `help:"Disabled"`

		// DefaultProject string `help:"Default project"`
		SystemAccount bool `help:"is a system account?"`
		NoWebConsole  bool `help:"allow web console access"`
		EnableMfa     bool `help:"enable TOTP mfa"`
	}
	R(&UserCreateOptions{}, "user-create", "Create a user", func(s *mcclient.ClientSession, args *UserCreateOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.NAME), "name")
		if len(args.Domain) > 0 {
			domainId, err := modules.Domains.GetId(s, args.Domain, nil)
			if err != nil {
				return err
			}
			params.Add(jsonutils.NewString(domainId), "domain_id")
		}
		if len(args.Password) > 0 {
			params.Add(jsonutils.NewString(args.Password), "password")
		}
		if len(args.Displayname) > 0 {
			params.Add(jsonutils.NewString(args.Displayname), "displayname")
		}
		if len(args.Desc) > 0 {
			params.Add(jsonutils.NewString(args.Desc), "description")
		}
		if len(args.Email) > 0 {
			params.Add(jsonutils.NewString(args.Email), "email")
		}
		if len(args.Mobile) > 0 {
			params.Add(jsonutils.NewString(args.Mobile), "mobile")
		}
		if args.Enabled && !args.Disabled {
			params.Add(jsonutils.JSONTrue, "enabled")
		} else if !args.Enabled && args.Disabled {
			params.Add(jsonutils.JSONFalse, "enabled")
		}

		if args.SystemAccount {
			params.Add(jsonutils.JSONTrue, "is_system_account")
		}
		if args.NoWebConsole {
			params.Add(jsonutils.JSONFalse, "allow_web_console")
		}
		if args.EnableMfa {
			params.Add(jsonutils.JSONTrue, "enable_mfa")
		}

		/*if len(args.DefaultProject) > 0 {
			projId, err := modules.Projects.GetId(s, args.DefaultProject, nil)
			if err != nil {
				return err
			}
			params.Add(jsonutils.NewString(projId), "default_project_id")
		}*/

		user, err := modules.UsersV3.Create(s, params)
		if err != nil {
			return err
		}
		printObject(user)
		return nil
	})

	type UserUpdateOptions struct {
		ID          string `help:"ID or name of the user"`
		Domain      string `help:"Domain"`
		Name        string `help:"New name of the user"`
		Password    string `help:"New password"`
		Desc        string `help:"Description"`
		Displayname string `help:"Displayname"`
		Email       string `help:"Email"`
		Mobile      string `help:"Mobile"`
		Enabled     bool   `help:"Enabled"`
		Disabled    bool   `help:"Disabled"`

		SystemAccount    bool `help:"Turn on is_system_account"`
		NotSystemAccount bool `help:"Turn off is_system_account"`

		AllowWebConsole    bool `help:"Turn on allow_web_console"`
		DisallowWebConsole bool `help:"Turn off allow_web_console"`

		EnableMfa  bool `help:"turn on enable_mfa"`
		DisableMfa bool `help:"turn off enable_mfa"`

		// DefaultProject string `help:"Default project"`
		// Option []string `help:"User options"`
	}
	R(&UserUpdateOptions{}, "user-update", "Update a user", func(s *mcclient.ClientSession, args *UserUpdateOptions) error {
		query := jsonutils.NewDict()
		if len(args.Domain) > 0 {
			domainId, err := modules.Domains.GetId(s, args.Domain, nil)
			if err != nil {
				return err
			}
			query.Add(jsonutils.NewString(domainId), "domain_id")
		}
		uid, err := modules.UsersV3.GetId(s, args.ID, query)
		if err != nil {
			return err
		}
		params := jsonutils.NewDict()
		if len(args.Name) > 0 {
			params.Add(jsonutils.NewString(args.Name), "name")
		}
		if len(args.Password) > 0 {
			params.Add(jsonutils.NewString(args.Password), "password")
		}
		if len(args.Displayname) > 0 {
			params.Add(jsonutils.NewString(args.Displayname), "displayname")
		}
		if len(args.Desc) > 0 {
			params.Add(jsonutils.NewString(args.Desc), "description")
		}
		if len(args.Email) > 0 {
			params.Add(jsonutils.NewString(args.Email), "email")
		}
		if len(args.Mobile) > 0 {
			params.Add(jsonutils.NewString(args.Mobile), "mobile")
		}
		if args.Enabled && !args.Disabled {
			params.Add(jsonutils.JSONTrue, "enabled")
		} else if !args.Enabled && args.Disabled {
			params.Add(jsonutils.JSONFalse, "enabled")
		}
		if args.SystemAccount {
			params.Add(jsonutils.JSONTrue, "is_system_account")
		} else if args.NotSystemAccount {
			params.Add(jsonutils.JSONFalse, "is_system_account")
		}
		if args.AllowWebConsole {
			params.Add(jsonutils.JSONTrue, "allow_web_console")
		} else if args.DisallowWebConsole {
			params.Add(jsonutils.JSONFalse, "allow_web_console")
		}
		if args.EnableMfa {
			params.Add(jsonutils.JSONTrue, "enable_mfa")
		} else if args.DisableMfa {
			params.Add(jsonutils.JSONFalse, "enable_mfa")
		}
		// if len(args.DefaultProject) > 0 {
		// 	projId, err := modules.Projects.GetId(s, args.DefaultProject, nil)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	params.Add(jsonutils.NewString(projId), "default_project_id")
		// }
		//
		//   if len(args.Option) > 0 {
		//       uoptions := jsonutils.NewDict()
		//       for _, opt := range args.Option {
		//           pos := strings.IndexByte(opt, ':')
		//           key := opt[:pos]
		//           val := opt[pos+1:]
		//           uoptions.Add(jsonutils.NewString(val), key)
		//       }
		//       params.Add(uoptions, "_resource_options")
		//   }
		//
		user, err := modules.UsersV3.Patch(s, uid, params)
		if err != nil {
			return err
		}
		printObject(user)
		return nil
	})

	type UserGroupOptions struct {
		USER   string `help:"User ID or Name"`
		GROUP  string `help:"group ID or Name"`
		Domain string `help:"Domain"`
	}
	R(&UserGroupOptions{}, "user-join-group", "Add a user to a group", func(s *mcclient.ClientSession, args *UserGroupOptions) error {
		uid, gid, err := getUserGroupId(s, args.USER, args.GROUP, args.Domain)
		if err != nil {
			return err
		}
		_, err = modules.UsersV3.PutInContext(s, uid, nil, &modules.Groups, gid)
		if err != nil {
			return err
		}
		return nil
	})
	R(&UserGroupOptions{}, "user-in-group", "Check whether a user belongs a group", func(s *mcclient.ClientSession, args *UserGroupOptions) error {
		uid, gid, err := getUserGroupId(s, args.USER, args.GROUP, args.Domain)
		if err != nil {
			return err
		}
		_, err = modules.UsersV3.HeadInContext(s, uid, nil, &modules.Groups, gid)
		if err != nil {
			return err
		}
		return nil
	})
	R(&UserGroupOptions{}, "user-leave-group", "Remove a user from a group", func(s *mcclient.ClientSession, args *UserGroupOptions) error {
		uid, gid, err := getUserGroupId(s, args.USER, args.GROUP, args.Domain)
		if err != nil {
			return err
		}
		_, err = modules.UsersV3.DeleteInContext(s, uid, nil, &modules.Groups, gid)
		if err != nil {
			return err
		}
		return nil
	})

}
