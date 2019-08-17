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
	"yunion.io/x/onecloud/pkg/multicloud/openstack"
	"yunion.io/x/onecloud/pkg/util/shellutils"
)

func init() {
	type EipListOptions struct {
	}
	shellutils.R(&EipListOptions{}, "eip-list", "List eips", func(cli *openstack.SRegion, args *EipListOptions) error {
		eips, err := cli.GetEips()
		if err != nil {
			return err
		}
		printList(eips, 0, 0, 0, []string{})
		return nil
	})

	type EipDeleteOptions struct {
		ID string
	}
	shellutils.R(&EipDeleteOptions{}, "eip-delete", "Delete eip", func(cli *openstack.SRegion, args *EipDeleteOptions) error {
		return cli.DeleteEip(args.ID)
	})

}
