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

package aws

type SDBInstanceSpec struct {
	VcpuCount  int
	VmemSizeMb int
}

var DBInstanceSpecs = map[string]SDBInstanceSpec{
	"db.t2.micro":    {VcpuCount: 1, VmemSizeMb: 1 * 1024},
	"db.t2.small":    {VcpuCount: 1, VmemSizeMb: 2 * 1024},
	"db.t2.medium":   {VcpuCount: 2, VmemSizeMb: 4 * 1024},
	"db.t2.large":    {VcpuCount: 2, VmemSizeMb: 8 * 1024},
	"db.t2.xlarge":   {VcpuCount: 4, VmemSizeMb: 16 * 1024},
	"db.t2.2xlarge":  {VcpuCount: 8, VmemSizeMb: 32 * 1024},
	"db.m4.large":    {VcpuCount: 2, VmemSizeMb: 8 * 1024},
	"db.m4.xlarge":   {VcpuCount: 4, VmemSizeMb: 16 * 1024},
	"db.m4.2xlarge":  {VcpuCount: 8, VmemSizeMb: 32 * 1024},
	"db.m4.4xlarge":  {VcpuCount: 16, VmemSizeMb: 64 * 1024},
	"db.m4.10xlarge": {VcpuCount: 40, VmemSizeMb: 160 * 1024},
	"db.m4.16xlarge": {VcpuCount: 64, VmemSizeMb: 256 * 1024},
	"db.m3.medium":   {VcpuCount: 1, VmemSizeMb: 3.75 * 1024},
	"db.m3.large":    {VcpuCount: 2, VmemSizeMb: 7.5 * 1024},
	"db.m3.xlarge":   {VcpuCount: 4, VmemSizeMb: 15 * 1024},
	"db.m3.2xlarge":  {VcpuCount: 8, VmemSizeMb: 30 * 1024},
	"db.r3.large":    {VcpuCount: 2, VmemSizeMb: 15.25 * 1024},
	"db.r3.xlarge":   {VcpuCount: 4, VmemSizeMb: 30.5 * 1024},
	"db.r3.2xlarge":  {VcpuCount: 8, VmemSizeMb: 61 * 1024},
	"db.r3.4xlarge":  {VcpuCount: 16, VmemSizeMb: 122 * 1024},
	"db.r3.8xlarge":  {VcpuCount: 32, VmemSizeMb: 244 * 1024},
	"db.r4.large":    {VcpuCount: 2, VmemSizeMb: 15.25 * 1024},
	"db.r4.xlarge":   {VcpuCount: 4, VmemSizeMb: 30.5 * 1024},
	"db.r4.2xlarge":  {VcpuCount: 8, VmemSizeMb: 61 * 1024},
	"db.r4.4xlarge":  {VcpuCount: 16, VmemSizeMb: 122 * 1024},
	"db.r4.8xlarge":  {VcpuCount: 32, VmemSizeMb: 244 * 1024},
	"db.r4.16xlarge": {VcpuCount: 64, VmemSizeMb: 488 * 1024},
}
