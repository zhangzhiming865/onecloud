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
	"fmt"

	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/mcclient/options"
)

func init() {
	type HostListOptions struct {
		Schedtag  string `help:"List hosts in schedtag"`
		Zone      string `help:"List hosts in zone"`
		Region    string `help:"List hosts in region"`
		Wire      string `help:"List hosts in wire"`
		Image     string `help:"List hosts cached images" json:"cachedimage"`
		Storage   string `help:"List hosts attached to storages"`
		Baremetal string `help:"List hosts that is managed by baremetal system" choices:"true|false"`
		Empty     bool   `help:"show empty host" json:"-"`
		Occupied  bool   `help:"show occupid host" json:"-"`
		Enabled   bool   `help:"Show enabled host only" json:"-"`
		Disabled  bool   `help:"Show disabled host only" json:"-"`
		HostType  string `help:"Host type filter" choices:"baremetal|hypervisor|esxi|kubelet|hyperv|aliyun|azure|qcloud|aws|huawei|ucloud"`
		AnyMac    string `help:"Mac matches one of the host's interface"`

		IsBaremetal *bool `help:"filter host list by is_baremetal=true|false"`

		ResourceType string `help:"Resource type" choices:"shared|prepaid|dedicated"`

		Usable *bool `help:"List all zones that is usable"`

		Hypervisor string `help:"filter hosts by hypervisor"`

		StorageNotAttached bool `help:"List hosts not attach specified storage"`

		options.BaseListOptions
	}
	R(&HostListOptions{}, "host-list", "List hosts", func(s *mcclient.ClientSession, opts *HostListOptions) error {
		params, err := options.ListStructToParams(opts)
		if err != nil {
			return err
		}

		if opts.Empty {
			params.Add(jsonutils.JSONTrue, "is_empty")
		} else if opts.Occupied {
			params.Add(jsonutils.JSONFalse, "is_empty")
		}
		if opts.Enabled {
			params.Add(jsonutils.NewInt(1), "enabled")
		} else if opts.Disabled {
			params.Add(jsonutils.NewInt(0), "enabled")
		}
		result, err := modules.Hosts.List(s, params)
		if err != nil {
			return err
		}
		printList(result, modules.Hosts.GetColumns(s))
		return nil
	})

	type HostDetailOptions struct {
		ID string `help:"ID or name of host"`
	}
	R(&HostDetailOptions{}, "host-show", "Show details of a host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.Get(s, args.ID, nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-ping", "Ping a host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "ping", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-metadata", "Show metadata of a host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.GetMetadata(s, args.ID, nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostOpsOptions struct {
		ID []string `help:"ID or name of hosts"`
	}
	R(&HostOpsOptions{}, "host-enable", "Enable a host", func(s *mcclient.ClientSession, args *HostOpsOptions) error {
		results := modules.Hosts.BatchPerformAction(s, args.ID, "enable", nil)
		printBatchResults(results, modules.Hosts.GetColumns(s))
		return nil
	})

	R(&HostOpsOptions{}, "host-disable", "Disable a host", func(s *mcclient.ClientSession, args *HostOpsOptions) error {
		results := modules.Hosts.BatchPerformAction(s, args.ID, "disable", nil)
		printBatchResults(results, modules.Hosts.GetColumns(s))
		return nil
	})

	R(&HostOpsOptions{}, "host-syncstatus", "Synchronize status of a host", func(s *mcclient.ClientSession, args *HostOpsOptions) error {
		results := modules.Hosts.BatchPerformAction(s, args.ID, "syncstatus", nil)
		printBatchResults(results, modules.Hosts.GetColumns(s))
		return nil
	})

	R(&HostOpsOptions{}, "host-prepare", "Prepare a host for installation", func(s *mcclient.ClientSession, args *HostOpsOptions) error {
		results := modules.Hosts.BatchPerformAction(s, args.ID, "prepare", nil)
		printBatchResults(results, modules.Hosts.GetColumns(s))
		return nil
	})

	R(&HostDetailOptions{}, "host-ipmi", "Get IPMI information of a host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.GetSpecific(s, args.ID, "ipmi", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-logininfo", "Get SSH login information of a host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		srvid, e := modules.Hosts.GetId(s, args.ID, nil)
		if e != nil {
			return e
		}
		i, e := modules.Hosts.GetLoginInfo(s, srvid, nil)
		if e != nil {
			return e
		}
		printObject(i)
		return nil
	})

	R(&HostDetailOptions{}, "host-vnc", "Get VNC information of a host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.GetSpecific(s, args.ID, "vnc", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostPropertyOptions struct {
	}

	R(&HostPropertyOptions{}, "baremetal-register-script", "Get online baremetal register script", func(s *mcclient.ClientSession, args *HostPropertyOptions) error {
		result, err := modules.Hosts.Get(s, "bm-start-register-script", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostUpdateOptions struct {
		ID                string  `help:"ID or Name of Host"`
		Name              string  `help:"New name of the host"`
		Desc              string  `help:"New Description of the host"`
		CpuCommitBound    float64 `help:"CPU overcommit upper bound at this host"`
		MemoryCommitBound float64 `help:"Memory overcommit upper bound at this host"`
		MemoryReserved    string  `help:"Memory reserved"`
		CpuReserved       int64   `help:"CPU reserved"`
		HostType          string  `help:"Change host type, CAUTION!!!!" choices:"hypervisor|kubelet|esxi|baremetal"`
		AccessIp          string  `help:"Change access ip, CAUTION!!!!"`
	}
	R(&HostUpdateOptions{}, "host-update", "Update information of a host", func(s *mcclient.ClientSession, args *HostUpdateOptions) error {
		params := jsonutils.NewDict()
		if len(args.Name) > 0 {
			params.Add(jsonutils.NewString(args.Name), "name")
		}
		if len(args.Desc) > 0 {
			params.Add(jsonutils.NewString(args.Desc), "description")
		}
		if args.CpuCommitBound > 0.0 {
			params.Add(jsonutils.NewFloat(args.CpuCommitBound), "cpu_cmtbound")
		}
		if args.MemoryCommitBound > 0.0 {
			params.Add(jsonutils.NewFloat(args.MemoryCommitBound), "mem_cmtbound")
		}
		if len(args.MemoryReserved) > 0 {
			params.Add(jsonutils.NewString(args.MemoryReserved), "mem_reserved")
		}
		if args.CpuReserved > 0 {
			params.Add(jsonutils.NewInt(args.CpuReserved), "cpu_reserved")
		}
		if len(args.HostType) > 0 {
			params.Add(jsonutils.NewString(args.HostType), "host_type")
		}
		if len(args.AccessIp) > 0 {
			params.Add(jsonutils.NewString(args.AccessIp), "access_ip")
		}
		if params.Size() == 0 {
			return fmt.Errorf("Not data to update")
		}
		result, err := modules.Hosts.Update(s, args.ID, params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostConvertOptions struct {
		ID         string   `help:"Host ID or Name"`
		Name       string   `help:"New name of the converted host"`
		HOSTTYPE   string   `help:"Convert host type" choices:"hypervisor|esxi|kubelet|hyperv"`
		Image      string   `help:"Template image to install"`
		Raid       string   `help:"Raid to deploy" choices:"raid0|raid1|raid10|raid5|none"`
		RaidConfig []string `help:"Baremetal raid config"`
		Disk       []string `help:"Disk descriptions" metavar:"DISK"`
		Net        []string `help:"Network descriptions" metavar:"NETWORK"`
	}
	R(&HostConvertOptions{}, "host-convert-hypervisor", "Convert a baremetal into a hypervisor", func(s *mcclient.ClientSession, args *HostConvertOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.HOSTTYPE), "host_type")
		if len(args.Name) > 0 {
			params.Add(jsonutils.NewString(args.Name), "name")
		}
		if len(args.RaidConfig) > 0 && len(args.Raid) > 0 {
			return fmt.Errorf("Cannot specifiy raidconfig and raid simultaneously")
		} else if len(args.RaidConfig) > 0 {
			for i := 0; i < len(args.RaidConfig); i += 1 {
				params.Add(jsonutils.NewString(args.RaidConfig[i]), fmt.Sprintf("baremetal_disk_config.%d", i))
			}
		} else if len(args.Raid) > 0 {
			params.Add(jsonutils.NewString(args.Raid), "raid")
		}
		if len(args.Disk) > 0 && len(args.Image) > 0 {
			return fmt.Errorf("Cannot specify disk and image simultaneously")
		} else if len(args.Disk) > 0 {
			for i := 0; i < len(args.Disk); i += 1 {
				params.Add(jsonutils.NewString(args.Disk[i]), fmt.Sprintf("disk.%d", i))
			}
		} else if len(args.Image) > 0 {
			params.Add(jsonutils.NewString(args.Image), "image")
		}
		if len(args.Net) > 0 {
			for i := 0; i < len(args.Net); i += 1 {
				params.Add(jsonutils.NewString(args.Net[i]), fmt.Sprintf("net.%d", i))
			}
		}
		result, err := modules.Hosts.PerformAction(s, args.ID, "convert-hypervisor", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-undo-convert", "Undo converting a host to hypervisor", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "undo-convert", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-maintenance", "Reboot host into PXE offline OS, do maintenance jobs", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "maintenance", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-unmaintenance", "Reboot host back into disk installed OS", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "unmaintenance", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-start", "Power on host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "start", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-stop", "Power off host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "stop", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-reset", "Power reset host", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "reset", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostOpsOptions{}, "host-delete", "Delete host record", func(s *mcclient.ClientSession, args *HostOpsOptions) error {
		results := modules.Hosts.BatchDelete(s, args.ID, nil)
		printBatchResults(results, modules.Hosts.GetColumns(s))
		return nil
	})

	type HostAddNetIfOptions struct {
		ID     string `help:"ID or Name of host"`
		WIRE   string `help:"ID or Name of wire to attach"`
		MAC    string `help:"Mac address of NIC"`
		Type   string `help:"Nic type" choices:"admin|ipmi"`
		IpAddr string `help:"IP address"`
	}
	R(&HostAddNetIfOptions{}, "host-add-netif", "Host add a NIC", func(s *mcclient.ClientSession, args *HostAddNetIfOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.WIRE), "wire")
		params.Add(jsonutils.NewString(args.MAC), "mac")
		params.Add(jsonutils.JSONTrue, "link_up")
		if len(args.Type) > 0 {
			params.Add(jsonutils.NewString(args.Type), "nic_type")
		}
		if len(args.IpAddr) > 0 {
			params.Add(jsonutils.NewString(args.IpAddr), "ip_addr")
		}
		result, err := modules.Hosts.PerformAction(s, args.ID, "add-netif", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostRemoveNetIfOptions struct {
		ID  string `help:"ID or Name of host"`
		MAC string `help:"MAC of NIC to remove"`
	}
	R(&HostRemoveNetIfOptions{}, "host-remove-netif", "Remove NIC from host", func(s *mcclient.ClientSession, args *HostRemoveNetIfOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.MAC), "mac")
		result, err := modules.Hosts.PerformAction(s, args.ID, "remove-netif", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostEnableNetIfOptions struct {
		ID       string `help:"ID or Name of host"`
		MAC      string `help:"MAC of NIC to enable"`
		Ip       string `help:"IP address"`
		Network  string `help:"network to connect"`
		Reserved bool   `help:"fetch IP from reserved pool"`
	}
	R(&HostEnableNetIfOptions{}, "host-enable-netif", "Enable a network interface for a host", func(s *mcclient.ClientSession, args *HostEnableNetIfOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.MAC), "mac")
		if len(args.Ip) > 0 {
			params.Add(jsonutils.NewString(args.Ip), "ip_addr")
			if args.Reserved {
				params.Add(jsonutils.JSONTrue, "reserve")
			}
		}
		if len(args.Network) > 0 {
			params.Add(jsonutils.NewString(args.Network), "network")
		}
		result, err := modules.Hosts.PerformAction(s, args.ID, "enable-netif", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostDisableNetIfOptions struct {
		ID      string `help:"ID or Name of host"`
		MAC     string `help:"MAC of NIC to disable"`
		Reserve bool   `help:"Reserve the IP address"`
	}
	R(&HostDisableNetIfOptions{}, "host-disable-netif", "Disable a network interface", func(s *mcclient.ClientSession, args *HostDisableNetIfOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.MAC), "mac")
		if args.Reserve {
			params.Add(jsonutils.JSONTrue, "reserve")
		}
		result, err := modules.Hosts.PerformAction(s, args.ID, "disable-netif", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	R(&HostDetailOptions{}, "host-remove-all-netifs", "Remvoe all netifs expect admin&ipmi netifs", func(s *mcclient.ClientSession, args *HostDetailOptions) error {
		result, err := modules.Hosts.PerformAction(s, args.ID, "remove-all-netifs", nil)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostCacheImageActionOptions struct {
		ID     string `help:"ID or name of host"`
		IMAGE  string `help:"ID or name of image"`
		Force  bool   `help:"Force refresh cache, even if the image exists in cache"`
		Format string `help:"image format" choices:"iso|vmdk|qcow2|vhd"`
	}
	R(&HostCacheImageActionOptions{}, "host-cache-image", "Ask a host to cache a image", func(s *mcclient.ClientSession, args *HostCacheImageActionOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.IMAGE), "image")
		if args.Force {
			params.Add(jsonutils.JSONTrue, "is_force")
		}
		host, err := modules.Hosts.PerformAction(s, args.ID, "cache-image", params)
		if err != nil {
			return err
		}
		printObject(host)
		return nil
	})

	type HostUncacheImageActionOptions struct {
		ID    string `help:"ID or name of host"`
		IMAGE string `help:"ID or name of image"`
	}
	R(&HostUncacheImageActionOptions{}, "host-uncache-image", "Ask a host to remove image from a cache", func(s *mcclient.ClientSession, args *HostUncacheImageActionOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.IMAGE), "image")
		host, err := modules.Hosts.PerformAction(s, args.ID, "uncache-image", params)
		if err != nil {
			return err
		}
		printObject(host)
		return nil
	})

	type HostCreateOptions struct {
		NAME       string `help:"Name of baremetal"`
		MAC        string `help:"Default MAC address of baremetal"`
		Rack       string `help:"Rack number of baremetal"`
		Slots      string `help:"Slots number of baremetal"`
		IpmiUser   string `help:"IPMI user name"`
		IpmiPasswd string `help:"IPMI user password"`
		IpmiAddr   string `help:"IPMI IP address"`
	}
	R(&HostCreateOptions{}, "host-create", "Create a baremetal host", func(s *mcclient.ClientSession, args *HostCreateOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.NAME), "name")
		params.Add(jsonutils.NewString(args.MAC), "access_mac")
		params.Add(jsonutils.NewString("baremetal"), "host_type")
		if len(args.Rack) > 0 {
			params.Add(jsonutils.NewString(args.Rack), "rack")
		}
		if len(args.Slots) > 0 {
			params.Add(jsonutils.NewString(args.Slots), "slots")
		}
		if len(args.IpmiUser) > 0 {
			params.Add(jsonutils.NewString(args.IpmiUser), "ipmi_username")
		}
		if len(args.IpmiPasswd) > 0 {
			params.Add(jsonutils.NewString(args.IpmiPasswd), "ipmi_password")
		}
		if len(args.IpmiAddr) > 0 {
			params.Add(jsonutils.NewString(args.IpmiAddr), "ipmi_ip_addr")
		}
		result, err := modules.Hosts.Create(s, params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostUndoPrepaidRecycleOptions struct {
		ID string `help:"ID or name of host to undo recycle"`
	}
	R(&HostUndoPrepaidRecycleOptions{}, "host-undo-recycle", "Pull a prepaid server from recycle pool, so that it will not be shared any more", func(s *mcclient.ClientSession, args *HostUndoPrepaidRecycleOptions) error {
		params := jsonutils.NewDict()
		result, err := modules.Hosts.PerformAction(s, args.ID, "undo-prepaid-recycle", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type PrepaidRecycleHostRenewOptions struct {
		ID       string `help:"ID or name of server to renew"`
		DURATION string `help:"Duration of renew, ADMIN only command"`
	}
	R(&PrepaidRecycleHostRenewOptions{}, "host-renew-prepaid-recycle", "Renew a prepaid recycle host", func(s *mcclient.ClientSession, args *PrepaidRecycleHostRenewOptions) error {
		params := jsonutils.NewDict()
		params.Add(jsonutils.NewString(args.DURATION), "duration")
		result, err := modules.Hosts.PerformAction(s, args.ID, "renew-prepaid-recycle", params)
		if err != nil {
			return err
		}
		printObject(result)
		return nil
	})

	type HostSpecOptions struct {
		ID string `help:"ID or name of host"`
	}
	R(&HostSpecOptions{}, "host-spec", "Get host spec info", func(s *mcclient.ClientSession, args *HostSpecOptions) error {
		spec, err := modules.Hosts.GetSpecific(s, args.ID, "spec", nil)
		if err != nil {
			return err
		}
		printObject(spec)
		return nil
	})
}
