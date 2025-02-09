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

package options

import (
	common_options "yunion.io/x/onecloud/pkg/cloudcommon/options"
	"yunion.io/x/onecloud/pkg/cloudcommon/pending_delete"
)

type ComputeOptions struct {
	PortV2 int `help:"Listening port for region V2"`

	DNSServer    string   `help:"Address of DNS server"`
	DNSDomain    string   `help:"Domain suffix for virtual servers"`
	DNSResolvers []string `help:"Upstream DNS resolvers"`

	IgnoreNonrunningGuests        bool    `default:"true" help:"Count memory for running guests only when do scheduling. Ignore memory allocation for non-running guests"`
	DefaultCPUOvercommitBound     float32 `default:"8.0" help:"Default cpu overcommit bound for host, default to 8"`
	DefaultMemoryOvercommitBound  float32 `default:"1.0" help:"Default memory overcommit bound for host, default to 1"`
	DefaultStorageOvercommitBound float32 `default:"1.0" help:"Default storage overcommit bound for storage, default to 1"`

	DefaultSecurityRules      string `help:"Default security rules" default:"allow any"`
	DefaultAdminSecurityRules string `help:"Default admin security rules" default:""`

	DefaultDiskSizeMB int `default:"10240" help:"Default disk size in MB if not specified, default to 10GiB" json:"default_disk_size"`

	pending_delete.SPendingDeleteOptions

	PrepaidExpireCheck              bool `default:"false" help:"clean expired servers or disks"`
	PrepaidExpireCheckSeconds       int  `default:"600" help:"How long to wait to scan expired prepaid VM or disks, default is 10 minutes"`
	ExpiredPrepaidMaxCleanBatchSize int  `default:"50" help:"How many expired prepaid servers can be deleted in a batch"`

	LoadbalancerPendingDeleteCheckInterval int `default:"3600" help:"Interval between checks of pending deleted loadbalancer objects, defaults to 1h"`

	ImageCacheStoragePolicy string `default:"least_used" choices:"best_fit|least_used" help:"Policy to choose storage for image cache, best_fit or least_used"`
	MetricsRetentionDays    int32  `default:"30" help:"Retention days for monitoring metrics in influxdb"`

	DefaultBandwidth int `default:"1000" help:"Default bandwidth"`
	DefaultMtu       int `default:"1500" help:"Default network mtu"`

	DefaultCpuQuota            int `help:"Common CPU quota per tenant, default 200" default:"200"`
	DefaultMemoryQuota         int `default:"204800" help:"Common memory quota per tenant in MB, default 200G"`
	DefaultStorageQuota        int `default:"12288000" help:"Common storage quota per tenant in MB, default 12T"`
	DefaultPortQuota           int `default:"200" help:"Common network port quota per tenant, default 200"`
	DefaultEipQuota            int `default:"10" help:"Common floating IP quota per tenant, default 10"`
	DefaultEportQuota          int `default:"200" help:"Common exit network port quota per tenant, default 200"`
	DefaultBwQuota             int `default:"2000000" help:"Common network port bandwidth in mbps quota per tenant, default 200*10Gbps"`
	DefaultEbwQuota            int `default:"4000" help:"Common exit network port bandwidth quota per tenant, default 4Gbps"`
	DefaultKeypairQuota        int `default:"50" help:"Common keypair quota per tenant, default 50"`
	DefaultGroupQuota          int `default:"50" help:"Common group quota per tenant, default 50"`
	DefaultSecgroupQuota       int `default:"50" help:"Common security group quota per tenant, default 50"`
	DefaultIsolatedDeviceQuota int `default:"200" help:"Common isolated device quota per tenant, default 200"`
	DefaultSnapshotQuota       int `default:"10" help:"Common snapshot quota per tenant, default 10"`
	DefaultBucketQuota         int `default:"100" help:"Common bucket quota per tenant, default 100"`
	DefaultObjectGBQuota       int `default:"100" help:"Common object size quota per tenant in GB, default 100GB"`
	DefaultObjectCntQuota      int `default:"500" help:"Common object count quota per tenant, default 500"`

	SystemAdminQuotaCheck bool `help:"Enable quota check for system admin, default False" default:"false"`

	BaremetalPreparePackageUrl string `help:"Baremetal online register package"`

	// snapshot options
	AutoSnapshotDay               int `default:"1" help:"Days auto snapshot disks, default 1 day"`
	AutoSnapshotHour              int `default:"2" help:"What hour take sanpshot, default 02:00"`
	DefaultMaxSnapshotCount       int `default:"9" help:"Per Disk max snapshot count, default 9"`
	DefaultMaxManualSnapshotCount int `default:"2" help:"Per Disk max manual snapshot count, default 2"`

	// sku sync
	SyncSkusDay  int `default:"1" help:"Days auto sync skus data, default 1 day"`
	SyncSkusHour int `default:"3" help:"What hour start sync skus, default 03:00"`

	// aws instance type file
	DefaultAwsInstanceTypeFile string `default:"/etc/yunion/aws_instance_types.json" help:"aws instance type json file"`

	ConvertHypervisorDefaultTemplate string `help:"Kvm baremetal convert option"`
	ConvertEsxiDefaultTemplate       string `help:"ESXI baremetal convert option"`
	ConvertKubeletDockerVolumeSize   string `default:"256g" help:"Docker volume size"`

	DefaultImageCacheDir string `default:"image_cache"`

	SnapshotCreateDiskProtocol string `help:"Snapshot create disk protocol" choices:"url|fuse" default:"fuse"`

	HostOfflineMaxSeconds        int `help:"Maximal seconds interval that a host considered offline during which it did not ping region, default is 3 minues" default:"180"`
	HostOfflineDetectionInterval int `help:"Interval to check offline hosts, defualt is half a minute" default:"30"`

	MinimalIpAddrReusedIntervalSeconds int `help:"Minimal seconds when a release IP address can be reallocate" default:"30"`

	CloudSyncWorkerCount         int `help:"how many current synchronization threads" default:"2"`
	CloudAutoSyncIntervalSeconds int `help:"frequency to check auto sync tasks" default:"30"`
	DefaultSyncIntervalSeconds   int `help:"minimal synchronization interval, default 1 minutes" default:"900"`
	MinimalSyncIntervalSeconds   int `help:"minimal synchronization interval, default 1 minutes" default:"300"`
	MaxCloudAccountErrorCount    int `help:"maximal consecutive error count allow for a cloud account" default:"5"`

	NameSyncResources []string `help:"resources that need synchronization of name"`

	SyncPurgeRemovedResources []string `help:"resources that shoud be purged immediately if found removed"`

	DisconnectedCloudAccountRetryProbeIntervalHours int `help:"interval to wait to probe status of a disconnected cloud account" default:"24"`

	SCapabilityOptions
	common_options.CommonOptions
	common_options.DBOptions
}

type SCapabilityOptions struct {
	MinDataDiskCount   int `help:"Minimal data disk count" default:"0"`
	MaxDataDiskCount   int `help:"Maximal data disk count" default:"12"`
	MinNicCount        int `help:"Minimal nic count" default:"1"`
	MaxNormalNicCount  int `help:"Maximal nic count" default:"8"`
	MaxManagedNicCount int `help:"Maximal managed nic count" default:"1"`
}

var (
	Options ComputeOptions
)
