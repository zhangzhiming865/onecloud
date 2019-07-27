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

package storagedrivers

import (
	"context"
	//	"fmt"
	//	"strings"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
)

type SMebsStorageDriver struct {
	SBaseStorageDriver
}

func init() {
	driver := SMebsStorageDriver{}
	models.RegisterStorageDriver(&driver)
}

func (self *SMebsStorageDriver) GetStorageType() string {
	return api.STORAGE_MEBS
}

func (self *SMebsStorageDriver) ValidateCreateData(ctx context.Context, userCred mcclient.TokenCredential, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	log.Infof("SMebsStorageDriver:param data:%v", data)
	conf := jsonutils.NewDict()
	/*for _, v := range []string{"mebs_host","mebs_port"} {
		if !data.Contains(v) {
			return nil, httperrors.NewMissingParameterError(v)
		}
		value, _ := data.GetString(v)
		conf.Add(jsonutils.NewString(value), strings.TrimPrefix(v, "mebs_"))
	}*/
	if !data.Contains("mebs_host") {
		return nil, httperrors.NewMissingParameterError("mebs_host")
	}
	mebs_host, _ := data.GetString("mebs_host")
	if !data.Contains("mebs_port") {
		return nil, httperrors.NewMissingParameterError("mebs_port")
	}
	mebs_port, _ := data.GetString("mebs_port")
	new_host := mebs_host
	new_host += ":"
	new_host += mebs_port
	conf.Add(jsonutils.NewString("mon_host"), new_host)
	/*if timeout, _ := data.Int("rbd_timeout"); timeout > 0 {
		conf.Add(jsonutils.NewInt(timeout), "rados_osd_op_timeout")
		conf.Add(jsonutils.NewInt(timeout), "rados_mon_op_timeout")
		conf.Add(jsonutils.NewInt(timeout), "client_mount_timeout")
	}*/

	storages := []models.SStorage{}
	q := models.StorageManager.Query().Equals("storage_type", api.STORAGE_MEBS)
	if err := db.FetchModelObjects(models.StorageManager, q, &storages); err != nil {
		return nil, httperrors.NewGeneralError(err)
	}

	inputHost, _ := conf.GetString("mon_host")
	for i := 0; i < len(storages); i++ {
		mhost, _ := storages[i].StorageConf.GetString("mon_host")
		if inputHost == mhost {
			return nil, httperrors.NewDuplicateResourceError("This MEBS Storage[%s/%s] has already exist", storages[i].Name, inputHost)
		}
	}

	data.Set("storage_conf", conf)

	return data, nil
}

func (self *SMebsStorageDriver) PostCreate(ctx context.Context, userCred mcclient.TokenCredential, storage *models.SStorage, data jsonutils.JSONObject) {
	storages := []models.SStorage{}
	q := models.StorageManager.Query().Equals("storage_type", api.STORAGE_MEBS)
	if err := db.FetchModelObjects(models.StorageManager, q, &storages); err != nil {
		log.Errorf("fetch storages error: %v", err)
		return
	}
	mebsHost, _ := data.GetString("mebs_host")
	mebsPort, _ := data.GetString("mebs_port")
	host := mebsHost
	host += ":"
	host += mebsPort
	for i := 0; i < len(storages); i++ {
		mhost, _ := storages[i].StorageConf.GetString("mon_host")
		if mhost == host {
			_, err := db.Update(storage, func() error {
				storage.StoragecacheId = storages[i].StoragecacheId
				return nil
			})
			if err != nil {
				log.Errorf("Update storagecacheId error: %v", err)
				return
			}
		}
	}
	/*
		if len(storage.StoragecacheId) == 0 {
			sc := &models.SStoragecache{}
			sc.SetModelManager(models.StoragecacheManager, sc)
			sc.Name = fmt.Sprintf("imagecache-%s", storage.Id)
			pool, _ := data.GetString("rbd_pool")
			sc.Path = fmt.Sprintf("rbd:%s", pool)
			if err := models.StoragecacheManager.TableSpec().Insert(sc); err != nil {
				log.Errorf("insert storagecache for storage %s error: %v", storage.Name, err)
				return
			}
			_, err := db.Update(storage, func() error {
				storage.StoragecacheId = sc.Id
				return nil
			})
			if err != nil {
				log.Errorf("update storagecache info for storage %s error: %v", storage.Name, err)
			}
		}*/
}
