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

// +build linux

package storageman

import (
	"context"
	"fmt"

	"github.com/ceph/go-ceph/rbd"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/utils"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/appctx"
	"yunion.io/x/onecloud/pkg/hostman/hostutils"
)

type SMEBSDisk struct {
	SBaseDisk
}

func NewMEBSDisk(storage IStorage, id string) *SMEBSDisk {
	var ret = new(SMEBSDisk)
	ret.SBaseDisk = *NewBaseDisk(storage, id)
	return ret
}

func (d *SMEBSDisk) GetType() string {
	return api.STORAGE_MEBS
}

func (d *SMEBSDisk) Probe() error {
	storage := d.Storage.(*SMEBSStorage)
	storageConf := d.Storage.GetStorageConf()
	pool, _ := storageConf.GetString("pool")
	_, err := storage.withImage(pool, d.Id, func(image *rbd.Image) (interface{}, error) {
		return image.GetSize()
	})
	return err
}

func (d *SMEBSDisk) getPath() string {
	storageConf := d.Storage.GetStorageConf()
	pool, _ := storageConf.GetString("pool")
	return fmt.Sprintf("rbd:%s/%s", pool, d.Id)
}

func (d *SMEBSDisk) GetPath() string {
	storage := d.Storage.(*SMEBSStorage)
	return fmt.Sprintf("%s%s", d.getPath(), storage.getStorageConfString())
}

func (d *SMEBSDisk) GetSnapshotDir() string {
	return ""
}

func (d *SMEBSDisk) GetDiskDesc() jsonutils.JSONObject {
	storage := d.Storage.(*SMEBSStorage)
	storageConf := d.Storage.GetStorageConf()
	pool, _ := storageConf.GetString("pool")
	desc := map[string]interface{}{
		"disk_id":     d.Id,
		"disk_format": "raw",
		"disk_path":   d.GetPath(),
		"disk_size":   storage.getImageSizeMb(pool, d.Id),
	}
	return jsonutils.Marshal(desc)
}

func (d *SMEBSDisk) GetDiskSetupScripts(idx int) string {
	return fmt.Sprintf("DISK_%d=%s\n", idx, d.GetPath())
}

func (d *SMEBSDisk) DeleteAllSnapshot() error {
	return fmt.Errorf("Not Impl")
}

func (d *SMEBSDisk) Delete(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	storage := d.Storage.(*SMEBSStorage)
	storageConf := d.Storage.GetStorageConf()
	pool, _ := storageConf.GetString("pool")
	return nil, storage.deleteImage(pool, d.Id)
}

func (d *SMEBSDisk) Resize(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	diskInfo, ok := params.(*jsonutils.JSONDict)
	if !ok {
		return nil, hostutils.ParamsError
	}
	storage := d.Storage.(*SMEBSStorage)
	storageConf := d.Storage.GetStorageConf()
	pool, _ := storageConf.GetString("pool")
	sizeMb, _ := diskInfo.Int("size")
	if err := storage.resizeImage(pool, d.Id, uint64(sizeMb)); err != nil {
		return nil, err
	}

	d.ResizeFs(d.GetPath())
	return d.GetDiskDesc(), nil
}

func (d *SMEBSDisk) PrepareSaveToGlance(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	if err := d.Probe(); err != nil {
		return nil, err
	}
	imageName := fmt.Sprintf("image_cache_%s_%s", d.Id, appctx.AppContextTaskId(ctx))
	imageCache := storageManager.GetStoragecacheById(d.Storage.GetStoragecacheId())
	if imageCache == nil {
		return nil, fmt.Errorf("failed to find image cache for prepare save to glance")
	}
	storage := d.Storage.(*SMEBSStorage)
	pool, _ := storage.GetStorageConf().GetString("pool")
	if err := storage.cloneImage(pool, d.Id, imageCache.GetPath(), imageName); err != nil {
		log.Errorf("clone image %s from pool %s to %s/%s error: %v", d.Id, pool, imageCache.GetPath(), imageName, err)
		return nil, err
	}
	return jsonutils.Marshal(map[string]string{"backup": imageName}), nil
}

func (d *SMEBSDisk) ResetFromSnapshot(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	return nil, fmt.Errorf("Not impl")
}

func (d *SMEBSDisk) CleanupSnapshots(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	storage := d.Storage.(*SMEBSStorage)
	pool, _ := storage.StorageConf.GetString("pool")
	return nil, storage.deleteSnapshot(pool, d.Id, "")
}

func (d *SMEBSDisk) PrepareMigrate(liveMigrate bool) (string, error) {
	return "", fmt.Errorf("Not support")
}

func (d *SMEBSDisk) CreateFromTemplate(ctx context.Context, imageId string, format string, size int64) (jsonutils.JSONObject, error) {
	ret, err := d.createFromTemplate(ctx, imageId, format)
	if err != nil {
		return nil, err
	}

	retSize, _ := ret.Int("disk_size")
	log.Infof("REQSIZE: %d, RETSIZE: %d", size, retSize)
	if size > retSize {
		params := jsonutils.NewDict()
		params.Set("size", jsonutils.NewInt(size))
		return d.Resize(ctx, params)
	}

	return ret, nil
}

func (d *SMEBSDisk) createFromTemplate(ctx context.Context, imageId, format string) (jsonutils.JSONObject, error) {
	var imageCacheManager = storageManager.GetStoragecacheById(d.Storage.GetStoragecacheId())
	if imageCacheManager == nil {
		return nil, fmt.Errorf("failed to find image cache manger for storage %s", d.Storage.GetStorageName())
	}
	imageCache := imageCacheManager.AcquireImage(ctx, imageId, d.GetZone(), "", "")
	if imageCache == nil {
		return nil, fmt.Errorf("failed to qcquire image for storage %s", d.Storage.GetStorageName())
	}
	defer imageCacheManager.ReleaseImage(imageId)
	storage := d.Storage.(*SMEBSStorage)
	destPool, _ := storage.StorageConf.GetString("pool")
	if err := storage.cloneImage(imageCacheManager.GetPath(), imageCache.GetName(), destPool, d.Id); err != nil {
		return nil, err
	}
	return d.GetDiskDesc(), nil
}

func (d *SRBDDisk) CreateFromImageFuse(ctx context.Context, url string, size int64) error {
	return fmt.Errorf("Not support")
}

func (d *SMEBSDisk) CreateRaw(ctx context.Context, sizeMb int, diskFromat string, fsFormat string, encryption bool, diskId string, back string) (jsonutils.JSONObject, error) {
	storage := d.Storage.(*SMEBSStorage)
	pool, _ := storage.StorageConf.GetString("pool")
	if err := storage.createImage(pool, diskId, uint64(sizeMb)); err != nil {
		return nil, err
	}

	if utils.IsInStringArray(fsFormat, []string{"swap", "ext2", "ext3", "ext4", "xfs"}) {
		d.FormatFs(fsFormat, diskId, d.GetPath())
	}

	return d.GetDiskDesc(), nil
}

func (d *SMEBSDisk) PostCreateFromImageFuse() {
	log.Errorf("Not support PostCreateFromImageFuse")
}

func (d *SMEBSDisk) CreateSnapshot(snapshotId string) error {
	storage := d.Storage.(*SMEBSStorage)
	pool, _ := storage.StorageConf.GetString("pool")
	return storage.createSnapshot(pool, d.Id, snapshotId)
}

func (d *SMEBSDisk) DeleteSnapshot(snapshotId, convertSnapshot string, pendingDelete bool) error {
	storage := d.Storage.(*SMEBSStorage)
	pool, _ := storage.StorageConf.GetString("pool")
	return storage.deleteSnapshot(pool, d.Id, snapshotId)
}
