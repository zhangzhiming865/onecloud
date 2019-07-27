/**
* @author
*    fuyuandi fuyuandi2008@163.com
*    zhangzhiming zhangzhiming865@163.com
* ©2019 fuyuandi and zhangzhiming. All Rights Reserved.
**/

// +build linux

package storageman

import (
	"context"
	"fmt"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/utils"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/appctx"
	"yunion.io/x/onecloud/pkg/hostman/hostutils"
	"yunion.io/x/onecloud/pkg/hostman/storageman/mebs"
	"yunion.io/x/onecloud/pkg/util/procutils"
)

type SMEBSDisk struct {
	SBaseDisk
	mebsStorage     *SMEBSStorage
	mebsStorageConf *jsonutils.JSONDict
}

func NewMEBSDisk(storage IStorage, id string) *SMEBSDisk {
	var ret = new(SMEBSDisk)
	ret.SBaseDisk = *NewBaseDisk(storage, id)
	ret.mebsStorage = ret.Storage.(*SMEBSStorage)
	ret.mebsStorageConf = ret.mebsStorage.GetStorageConf()
	return ret
}

func (d *SMEBSDisk) GetType() string {
	return api.STORAGE_MEBS
}

func (d *SMEBSDisk) Probe() error {
	log.Infof("mebs Probe, id %v", d.Id)
	hostname, _ := d.mebsStorageConf.GetString("mon_host")
	_, err := mebs.MebsImageInfo(d.Id, hostname)
	return err
}

func (d *SMEBSDisk) doGetLaunchCmds() string {
	hostname, _ := d.mebsStorageConf.GetString("mon_host")
	return "curl -s -X POST http://" + hostname + "/volumes/" + d.Id + "/launch?by_id=false " +
		"-d '{\"protocol\":\"nbd\"}' | sed -n 's/.*url\\\":\\\"\\(.*\\)\\\",\\\".*/\\1/p'"
}

func (d *SMEBSDisk) getLaunchCmds() string {
	log.Infof("mebs getPath")
	return "`" + d.doGetLaunchCmds() + "`"
}

func (d *SMEBSDisk) GetPath() string {
	log.Infof("mebs GetPath, %v", d.Id)
	bashCmd := d.doGetLaunchCmds()
	output, err := procutils.Run("bash", "-c", bashCmd)
	if err != nil {
		return "FailedToGetMebsPath!!!"
	}
	return output[0]
}

func (d *SMEBSDisk) GetSnapshotDir() string {
	log.Infof("mebs GetSnapshotDir")
	return ""
}

func (d *SMEBSDisk) GetDiskDesc() jsonutils.JSONObject {
	log.Infof("mebs GetDiskDesc")
	storage := d.Storage.(*SMEBSStorage)
	storageConf := d.Storage.GetStorageConf()
	log.Infof("mebs storage conf is %v", storageConf)
	desc := map[string]interface{}{
		"disk_id":     d.Id,
		"disk_format": "raw",
		"disk_path":   d.GetPath(),
		"disk_size":   storage.getImageSizeMb(d.Id),
	}
	return jsonutils.Marshal(desc)
}

func (d *SMEBSDisk) GetDiskSetupScripts(idx int) string {
	log.Infof("mebs GetDiskSetupScripts")
	return fmt.Sprintf("DISK_%d=%s\n", idx, d.getLaunchCmds())
}

func (d *SMEBSDisk) DeleteAllSnapshot() error {
	log.Infof("mebs DeleteAllSnapshot")
	return fmt.Errorf("Not Impl")
}

func (d *SMEBSDisk) Delete(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	log.Infof("mebs Delete %v", params)
	storage := d.Storage.(*SMEBSStorage)
	storageConf := d.Storage.GetStorageConf()
	log.Infof("storage conf is %v", storageConf)
	return nil, storage.deleteImage(d.Id)
}

func (d *SMEBSDisk) Resize(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	log.Infof("mebs Resize %v", params)
	diskInfo, ok := params.(*jsonutils.JSONDict)
	if !ok {
		return nil, hostutils.ParamsError
	}
	storage := d.Storage.(*SMEBSStorage)
	storageConf := d.Storage.GetStorageConf()
	log.Infof("mebs storageConf %v", storageConf)
	sizeMb, _ := diskInfo.Int("size")
	if err := storage.resizeImage(d.Id, uint64(sizeMb)); err != nil {
		return nil, err
	}

	d.ResizeFs(d.GetPath())
	return d.GetDiskDesc(), nil
}

func (d *SMEBSDisk) PrepareSaveToGlance(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	log.Infof("mebs PrepareSaveToGlance %v", params)
	if err := d.Probe(); err != nil {
		return nil, err
	}
	imageName := fmt.Sprintf("image_cache_%s_%s", d.Id, appctx.AppContextTaskId(ctx))
	imageCache := storageManager.GetStoragecacheById(d.Storage.GetStoragecacheId())
	if imageCache == nil {
		return nil, fmt.Errorf("failed to find image cache for prepare save to glance")
	}
	storage := d.Storage.(*SMEBSStorage)
	if err := storage.cloneImage(d.Id, imageName); err != nil {
		log.Errorf("clone image %s to %s error: %v", d.Id, imageName, err)
		return nil, err
	}
	return jsonutils.Marshal(map[string]string{"backup": imageName}), nil
}

func (d *SMEBSDisk) ResetFromSnapshot(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	return nil, fmt.Errorf("Not impl")
}

func (d *SMEBSDisk) CleanupSnapshots(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	storage := d.Storage.(*SMEBSStorage)
	return nil, storage.deleteSnapshot(d.Id, "")
}

func (d *SMEBSDisk) PrepareMigrate(liveMigrate bool) (string, error) {
	return "", fmt.Errorf("Not support")
}

func (d *SMEBSDisk) CreateFromTemplate(ctx context.Context, imageId string, format string, size int64) (jsonutils.JSONObject, error) {
	log.Infof("mebs CreateFromTemplate image id %v format %v", imageId, format)
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
	log.Infof("mebs createFromTemplate image id %v format %v", imageId, format)
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
	if err := storage.cloneImage(imageCache.GetName(), d.Id); err != nil {
		return nil, err
	}
	return d.GetDiskDesc(), nil
}

func (d *SMEBSDisk) CreateFromImageFuse(ctx context.Context, url string, size int64) error {
	return fmt.Errorf("Not support")
}

func (d *SMEBSDisk) CreateRaw(ctx context.Context, sizeMb int, diskFromat string, fsFormat string, encryption bool, diskId string, back string) (jsonutils.JSONObject, error) {
	log.Infof("mebs CreateRaw diskid %v, fsFormat %v diskFromat %v", diskId, fsFormat, diskFromat)
	storage := d.Storage.(*SMEBSStorage)
	if err := storage.createImage(diskId, uint64(sizeMb)); err != nil {
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
	log.Infof("mebs CreateSnapshot %v", snapshotId)
	storage := d.Storage.(*SMEBSStorage)
	return storage.createSnapshot(d.Id, snapshotId)
}

func (d *SMEBSDisk) DeleteSnapshot(snapshotId, convertSnapshot string, pendingDelete bool) error {
	log.Infof("mebs DeleteSnapshot %v", snapshotId)
	storage := d.Storage.(*SMEBSStorage)
	return storage.deleteSnapshot(d.Id, snapshotId)
}
