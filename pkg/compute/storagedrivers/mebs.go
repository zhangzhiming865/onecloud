/**
* @author
*    fuyuandi fuyuandi2008@163.com
*    zhangzhiming zhangzhiming865@163.com
* ©2019 fuyuandi and zhangzhiming. All Rights Reserved.
**/

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
	conf.Add(jsonutils.NewString(new_host), "mon_host")

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
}
