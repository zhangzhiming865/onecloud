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

package models

import (
	"context"
	"fmt"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/util/compare"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/lockman"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type SAwsCachedLbManager struct {
	SLoadbalancerLogSkipper
	db.SVirtualResourceBaseManager
}

var AwsCachedLbManager *SAwsCachedLbManager

func init() {
	AwsCachedLbManager = &SAwsCachedLbManager{
		SVirtualResourceBaseManager: db.NewVirtualResourceBaseManager(
			SAwsCachedLb{},
			"awscachedlbbs_tbl",
			"awscachedlbb",
			"awscachedlbbs",
		),
	}
	AwsCachedLbManager.SetVirtualObject(AwsCachedLbManager)
}

type SAwsCachedLb struct {
	db.SVirtualResourceBase
	db.SExternalizedResourceBase

	SManagedResourceBase
	SCloudregionResourceBase

	BackendServerId      string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional"` // 后端服务器 实例ID
	BackendId            string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional"` // 本地loadbalancebackend id
	CachedBackendGroupId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional"`
}

func (lbb *SAwsCachedLb) GetCustomizeColumns(context.Context, mcclient.TokenCredential, jsonutils.JSONObject) *jsonutils.JSONDict {
	return nil
}

func (man *SAwsCachedLbManager) GetBackendsByLocalBackendId(backendId string) ([]SAwsCachedLb, error) {
	loadbalancerBackends := []SAwsCachedLb{}
	q := man.Query().Equals("backend_id", backendId)
	if err := db.FetchModelObjects(man, q, &loadbalancerBackends); err != nil {
		return nil, err
	}
	return loadbalancerBackends, nil
}

func (man *SAwsCachedLbManager) CreateAwsCachedLb(ctx context.Context, userCred mcclient.TokenCredential, lbb *SLoadbalancerBackend, cachedLbbg *SAwsCachedLbbg, extLoadbalancerBackend cloudprovider.ICloudLoadbalancerBackend, syncOwnerId mcclient.IIdentityProvider) (*SAwsCachedLb, error) {
	cachedlbb := &SAwsCachedLb{}
	cachedlbb.SetModelManager(man, cachedlbb)

	cachedlbb.CloudregionId = cachedLbbg.CloudregionId
	cachedlbb.ManagerId = cachedLbbg.ManagerId
	cachedlbb.CachedBackendGroupId = cachedLbbg.GetId()
	cachedlbb.BackendId = lbb.GetId()
	cachedlbb.ExternalId = extLoadbalancerBackend.GetGlobalId()

	newName, err := db.GenerateName(man, syncOwnerId, extLoadbalancerBackend.GetName())
	if err != nil {
		return nil, err
	}
	cachedlbb.Name = newName

	if err := cachedlbb.constructFieldsFromCloudLoadbalancerBackend(extLoadbalancerBackend); err != nil {
		return nil, err
	}

	err = man.TableSpec().Insert(lbb)

	if err != nil {
		return nil, err
	}

	SyncCloudProject(userCred, lbb, syncOwnerId, extLoadbalancerBackend, cachedLbbg.ManagerId)

	db.OpsLog.LogEvent(cachedlbb, db.ACT_CREATE, lbb.GetShortDesc(ctx), userCred)

	return cachedlbb, nil
}

func (lbb *SAwsCachedLb) GetCachedBackendGroup() (*SAwsCachedLbbg, error) {
	lbbg, err := db.FetchById(AwsCachedLbbgManager, lbb.CachedBackendGroupId)
	if err != nil {
		return nil, err
	}

	return lbbg.(*SAwsCachedLbbg), nil
}

func (man *SAwsCachedLbManager) getLoadbalancerBackendsByLoadbalancerBackendgroup(loadbalancerBackendgroup *SAwsCachedLbbg) ([]SAwsCachedLb, error) {
	loadbalancerBackends := []SAwsCachedLb{}
	q := man.Query().Equals("cached_backend_group_id", loadbalancerBackendgroup.Id)
	if err := db.FetchModelObjects(man, q, &loadbalancerBackends); err != nil {
		return nil, err
	}
	return loadbalancerBackends, nil
}

func (man *SAwsCachedLbManager) SyncLoadbalancerBackends(ctx context.Context, userCred mcclient.TokenCredential, provider *SCloudprovider, loadbalancerBackendgroup *SAwsCachedLbbg, lbbs []cloudprovider.ICloudLoadbalancerBackend, syncRange *SSyncRange) compare.SyncResult {
	syncOwnerId := provider.GetOwnerId()

	lockman.LockClass(ctx, man, db.GetLockClassKey(man, syncOwnerId))
	defer lockman.ReleaseClass(ctx, man, db.GetLockClassKey(man, syncOwnerId))

	syncResult := compare.SyncResult{}

	dbLbbs, err := man.getLoadbalancerBackendsByLoadbalancerBackendgroup(loadbalancerBackendgroup)
	if err != nil {
		syncResult.Error(err)
		return syncResult
	}

	removed := []SAwsCachedLb{}
	commondb := []SAwsCachedLb{}
	commonext := []cloudprovider.ICloudLoadbalancerBackend{}
	added := []cloudprovider.ICloudLoadbalancerBackend{}

	err = compare.CompareSets(dbLbbs, lbbs, &removed, &commondb, &commonext, &added)
	if err != nil {
		syncResult.Error(err)
		return syncResult
	}

	for i := 0; i < len(removed); i++ {
		err = removed[i].syncRemoveCloudLoadbalancerBackend(ctx, userCred)
		if err != nil {
			syncResult.DeleteError(err)
		} else {
			syncResult.Delete()
		}
	}
	for i := 0; i < len(commondb); i++ {
		err = commondb[i].SyncWithCloudLoadbalancerBackend(ctx, userCred, commonext[i], syncOwnerId)
		if err != nil {
			syncResult.UpdateError(err)
		} else {
			syncMetadata(ctx, userCred, &commondb[i], commonext[i])
			syncResult.Update()
		}
	}
	for i := 0; i < len(added); i++ {
		local, err := man.newFromCloudLoadbalancerBackend(ctx, userCred, loadbalancerBackendgroup, added[i], syncOwnerId)
		if err != nil {
			syncResult.AddError(err)
		} else {
			syncMetadata(ctx, userCred, local, added[i])
			syncResult.Add()
		}
	}
	return syncResult
}

func (lbb *SAwsCachedLb) syncRemoveCloudLoadbalancerBackend(ctx context.Context, userCred mcclient.TokenCredential) error {
	lockman.LockObject(ctx, lbb)
	defer lockman.ReleaseObject(ctx, lbb)

	err := lbb.ValidateDeleteCondition(ctx)
	if err != nil { // cannot delete
		err = lbb.SetStatus(userCred, api.LB_STATUS_UNKNOWN, "sync to delete")
	} else {
		lbb.SetModelManager(AwsCachedLbManager, lbb)
		err := db.DeleteModel(ctx, userCred, lbb)
		if err != nil {
			return err
		}
	}
	return err
}

func (lbb *SAwsCachedLb) constructFieldsFromCloudLoadbalancerBackend(extLoadbalancerBackend cloudprovider.ICloudLoadbalancerBackend) error {
	lbb.Status = extLoadbalancerBackend.GetStatus()

	instance, err := db.FetchByExternalId(GuestManager, extLoadbalancerBackend.GetBackendId())
	if err != nil {
		return err
	}
	guest := instance.(*SGuest)

	lbb.BackendServerId = guest.Id
	return nil
}

func (lbb *SAwsCachedLb) SyncWithCloudLoadbalancerBackend(ctx context.Context, userCred mcclient.TokenCredential, extLoadbalancerBackend cloudprovider.ICloudLoadbalancerBackend, syncOwnerId mcclient.IIdentityProvider) error {
	lbb.SetModelManager(AwsCachedLbManager, lbb)
	diff, err := db.UpdateWithLock(ctx, lbb, func() error {
		return lbb.constructFieldsFromCloudLoadbalancerBackend(extLoadbalancerBackend)
	})
	if err != nil {
		return err
	}
	db.OpsLog.LogSyncUpdate(lbb, diff, userCred)

	SyncCloudProject(userCred, lbb, syncOwnerId, extLoadbalancerBackend, lbb.ManagerId)

	return nil
}

func (man *SAwsCachedLbManager) newFromCloudLoadbalancerBackend(ctx context.Context, userCred mcclient.TokenCredential, loadbalancerBackendgroup *SAwsCachedLbbg, extLoadbalancerBackend cloudprovider.ICloudLoadbalancerBackend, syncOwnerId mcclient.IIdentityProvider) (*SAwsCachedLb, error) {
	localBackendGroup, err := loadbalancerBackendgroup.GetLocalBackendGroup(ctx, userCred)
	if err != nil {
		return nil, err
	} else if localBackendGroup == nil {
		return nil, fmt.Errorf("newFromCloudLoadbalancerBackend localBackendGroup is nil")
	}

	locallbb, err := newLocalBackendFromCloudLoadbalancerBackend(ctx, userCred, localBackendGroup, extLoadbalancerBackend, syncOwnerId)
	if err != nil {
		return nil, err
	}
	lbb := &SAwsCachedLb{}
	lbb.SetModelManager(man, lbb)

	lbb.CloudregionId = loadbalancerBackendgroup.CloudregionId
	lbb.ManagerId = loadbalancerBackendgroup.ManagerId
	lbb.CachedBackendGroupId = loadbalancerBackendgroup.Id
	lbb.BackendId = locallbb.GetId()
	lbb.ExternalId = extLoadbalancerBackend.GetGlobalId()

	newName, err := db.GenerateName(man, syncOwnerId, extLoadbalancerBackend.GetName())
	if err != nil {
		return nil, err
	}
	lbb.Name = newName

	if err := lbb.constructFieldsFromCloudLoadbalancerBackend(extLoadbalancerBackend); err != nil {
		return nil, err
	}

	err = man.TableSpec().Insert(lbb)

	if err != nil {
		return nil, err
	}

	SyncCloudProject(userCred, lbb, syncOwnerId, extLoadbalancerBackend, loadbalancerBackendgroup.ManagerId)

	db.OpsLog.LogEvent(lbb, db.ACT_CREATE, lbb.GetShortDesc(ctx), userCred)

	return lbb, nil
}
