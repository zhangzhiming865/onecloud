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
	_ "os"
	_ "strings"

	"github.com/pkg/errors"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	_ "yunion.io/x/pkg/utils"

	"yunion.io/x/onecloud/pkg/hostman/storageman/mebs"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	deployapi "yunion.io/x/onecloud/pkg/hostman/hostdeployer/apis"
	"yunion.io/x/onecloud/pkg/hostman/hostdeployer/deployclient"
	"yunion.io/x/onecloud/pkg/hostman/hostutils"
	"yunion.io/x/onecloud/pkg/hostman/options"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	_ "yunion.io/x/onecloud/pkg/util/procutils"
	_ "yunion.io/x/onecloud/pkg/util/qemutils"
)

var (
	ErrMebsNoSuchImage    = errors.New("mebs no such image")
	ErrMebsNoSuchSnapshot = errors.New("mebs no such snapshot")
)

type SMEBSStorage struct {
	SBaseStorage
}

func NewMEBSStorage(manager *SStorageManager, path string) *SMEBSStorage {
	var ret = new(SMEBSStorage)
	ret.SBaseStorage = *NewBaseStorage(manager, path)
	return ret
}

type SMebsStorageFactory struct {
}

func (factory *SMebsStorageFactory) NewStorage(manager *SStorageManager, mountPoint string) IStorage {
	return NewMEBSStorage(manager, mountPoint)
}

func (factory *SMebsStorageFactory) StorageType() string {
	return api.STORAGE_MEBS
}

func init() {
	registerStorageFactory(&SMebsStorageFactory{})
}

func (s *SMEBSStorage) StorageType() string {
	return api.STORAGE_MEBS
}

func (s *SMEBSStorage) GetSnapshotPathByIds(diskId, snapshotId string) string {
	log.Infof("mebs GetSnapshotPathByIds disk id %v snapshotid %v", diskId, snapshotId)
	return ""
}

func (s *SMEBSStorage) GetSnapshotDir() string {
	log.Infof("mebs get snapshot dir")
	return ""
}

func (s *SMEBSStorage) GetFuseTmpPath() string {
	log.Infof("mebs GetFuseTmpPath")
	return ""
}

func (s *SMEBSStorage) GetFuseMountPath() string {
	log.Infof("mebs GetFuseMountPath")
	return ""
}

func (s *SMEBSStorage) GetImgsaveBackupPath() string {
	log.Infof("mebs GetImgsaveBackupPath")
	return ""
}

//Tip Configuration values containing :, @, or = can be escaped with a leading \ character.
func (s *SMEBSStorage) getStorageConfString() string {
	log.Infof("mebs getStorageConfString")
	conf := ""
	for _, key := range []string{"mon_host"} {
		if value, _ := s.StorageConf.GetString(key); len(value) > 0 {
			conf += fmt.Sprintf(":%s=%s", key, value)
		}
	}
	return conf
}

func (s *SMEBSStorage) getImageSizeMb(name string) uint64 {
	log.Infof("mebs getImageSizeMb name %v", name)
	hostname, _ := s.GetStorageConf().GetString("mon_host")
	inf, err := mebs.MebsImageInfo(name, hostname)
	if err != nil {
		log.Errorf("error get image size of %v", name)
		return 0
	}
	return uint64(inf.Virtual_size) / 1024 / 1024
}

func (s *SMEBSStorage) resizeImage(name string, sizeMb uint64) error {
	log.Infof("mebs resizeImage name %v size %v", name, sizeMb)
	hostname, _ := s.GetStorageConf().GetString("mon_host")
	return mebs.MebsResizeImage(name, int64(sizeMb), hostname)
}

func (s *SMEBSStorage) deleteImage(name string) error {
	log.Infof("mebs deleteImage %v", name)
	hostname, _ := s.GetStorageConf().GetString("mon_host")
	return mebs.MebsRemoveImage(name, hostname)
}

// 比较费时
func (s *SMEBSStorage) copyImage(srcImage string, destImage string) error {
	log.Infof("mebs copyImage %v %v", srcImage, destImage)
	return fmt.Errorf("mebs not implement copy image")
}

// 速度快
func (s *SMEBSStorage) cloneImage(srcImage string, destImage string) error {
	log.Infof("close image src %v des %v", srcImage, destImage)
	hostname, _ := s.StorageConf.GetString("mon_host")
	return mebs.MebsCreateVolumeFromImage(srcImage, destImage, hostname)
}

func (s *SMEBSStorage) listImages() ([]string, error) {
	return []string{}, nil
}

func (s *SMEBSStorage) createImage(name string, sizeMb uint64) error {
	log.Infof("mebs createImage %v size %v", name, sizeMb)
	host, _ := s.GetStorageConf().GetString("mon_host")
	return mebs.MebsCreateImage(name, int64(sizeMb), host)
}

func (s *SMEBSStorage) renameImage(src string, dest string) error {
	log.Infof("mebs renameImage src %v dest %v", src, dest)
	return fmt.Errorf("mebs not implement rename image")
}

func (s *SMEBSStorage) createSnapshot(diskId string, snapshotId string) error {
	log.Infof("mebs createSnapshot diskid %v snapshotid %v", diskId, snapshotId)
	hostname, _ := s.StorageConf.GetString("mon_host")
	return mebs.MebsCreateSnapshot(diskId, snapshotId, hostname)
}

func (s *SMEBSStorage) deleteSnapshot(diskId string, snapshotId string) error {
	log.Infof("mebs deleteSnapshot diskid %v snapshotid %v", diskId, snapshotId)
	hostname, _ := s.StorageConf.GetString("mon_host")
	return mebs.MebsDeleteSnapshot(diskId, snapshotId, hostname)
}

func (s *SMEBSStorage) getCapacity() (uint64, error) {
	log.Infof("mebs getCapacity")
	sizeKb := 1024 * 1024 * 1024
	return uint64(sizeKb) / 1024, nil
}

func (s *SMEBSStorage) SyncStorageInfo() (jsonutils.JSONObject, error) {
	log.Infof("SyncStorageInfo MEBS......, storageid %v", s.StorageId)
	content := map[string]interface{}{}
	if len(s.StorageId) > 0 {
		capacity, err := s.getCapacity()
		if err != nil {
			return modules.Storages.PerformAction(hostutils.GetComputeSession(context.Background()), s.StorageId, "offline", nil)
		}
		content = map[string]interface{}{
			"name":     s.StorageName,
			"capacity": capacity,
			"status":   api.STORAGE_ONLINE,
			"zone":     s.GetZone(),
		}
		return modules.Storages.Put(hostutils.GetComputeSession(context.Background()), s.StorageId, jsonutils.Marshal(content))
	}
	return modules.Storages.Get(hostutils.GetComputeSession(context.Background()), s.StorageName, jsonutils.Marshal(content))
}

func (s *SMEBSStorage) GetDiskById(diskId string) IDisk {
	s.DiskLock.Lock()
	defer s.DiskLock.Unlock()
	for i := 0; i < len(s.Disks); i++ {
		if s.Disks[i].GetId() == diskId {
			if s.Disks[i].Probe() == nil {
				return s.Disks[i]
			}
		}
	}
	var disk = NewMEBSDisk(s, diskId)
	if disk.Probe() == nil {
		s.Disks = append(s.Disks, disk)
		return disk
	} else {
		return nil
	}
}

func (s *SMEBSStorage) CreateDisk(diskId string) IDisk {
	s.DiskLock.Lock()
	defer s.DiskLock.Unlock()
	disk := NewMEBSDisk(s, diskId)
	s.Disks = append(s.Disks, disk)
	return disk
}

func (s *SMEBSStorage) Accessible() bool {
	log.Infof("mebs accessible is true")
	return true
}

func (s *SMEBSStorage) SaveToGlance(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	log.Infof("mebs save to SaveToGlancen run %v", params)
	data, ok := params.(*jsonutils.JSONDict)
	if !ok {
		return nil, hostutils.ParamsError
	}

	mebsImageCache := storageManager.GetStoragecacheById(s.GetStoragecacheId())
	if mebsImageCache == nil {
		return nil, fmt.Errorf("failed to find storage image cache for storage %s", s.GetStorageName())
	}

	imagePath, _ := data.GetString("image_path")
	compress := jsonutils.QueryBoolean(data, "compress", true)
	format, _ := data.GetString("format")
	imageId, _ := data.GetString("image_id")
	imageName := imageId
	log.Infof("SaveToGlance imagePath %v compress %v format %v imageId %v", imagePath, compress, format, imageId)

	imagePath = fmt.Sprintf("mebs:%s/%s%s", mebsImageCache.GetPath(), imageName, s.getStorageConfString())

	if err := s.saveToGlance(ctx, imageId, imagePath, compress, format); err != nil {
		log.Errorf("Save to glance failed: %s", err)
		s.onSaveToGlanceFailed(ctx, imageId)
	}

	mebsImageCache.LoadImageCache(imageId)
	_, err := hostutils.RemoteStoragecacheCacheImage(ctx, mebsImageCache.GetId(), imageId, "ready", imagePath)
	if err != nil {
		log.Errorf("Fail to remote cache image: %v", err)
	}
	return nil, nil
}

func (s *SMEBSStorage) onSaveToGlanceFailed(ctx context.Context, imageId string) {
	log.Infof("mebs onSaveToGlanceFailed imageId %v", imageId)
	params := jsonutils.NewDict()
	params.Set("status", jsonutils.NewString("killed"))
	_, err := modules.Images.Update(hostutils.GetImageSession(ctx, s.GetZone()),
		imageId, params)
	if err != nil {
		log.Errorln(err)
	}
}

func (s *SMEBSStorage) saveToGlance(ctx context.Context, imageId, imagePath string, compress bool, format string) error {
	log.Infof("mebs saveToGlance imageid %v imagePath %v compress %v format %v", imageId, imagePath, compress, format)
	ret, err := deployclient.GetDeployClient().SaveToGlance(context.Background(),
		&deployapi.SaveToGlanceParams{DiskPath: imagePath, Compress: compress})
	if err != nil {
		return err
	}

	if len(format) == 0 {
		format = options.HostOptions.DefaultImageSaveFormat
	}

	host, _ := s.GetStorageConf().GetString("mon_host")
	size := mebs.MebsImageSize(imageId, host)

	var params = jsonutils.NewDict()
	if len(ret.OsInfo) > 0 {
		params.Set("os_type", jsonutils.NewString(ret.OsInfo))
	}
	relInfo := ret.ReleaseInfo
	if relInfo != nil {
		params.Set("os_distribution", jsonutils.NewString(relInfo.Distro))
		if len(relInfo.Version) > 0 {
			params.Set("os_version", jsonutils.NewString(relInfo.Version))
		}
		if len(relInfo.Arch) > 0 {
			params.Set("os_arch", jsonutils.NewString(relInfo.Arch))
		}
		if len(relInfo.Version) > 0 {
			params.Set("os_language", jsonutils.NewString(relInfo.Language))
		}
	}
	params.Set("image_id", jsonutils.NewString(imageId))

	_, err = modules.Images.Upload(hostutils.GetImageSession(ctx, s.GetZone()),
		params, nil, size)
	return err
}

func (s *SMEBSStorage) CreateSnapshotFormUrl(ctx context.Context, snapshotUrl, diskId, snapshotPath string) error {
	return fmt.Errorf("Not support")
}

func (s *SMEBSStorage) DeleteSnapshots(ctx context.Context, params interface{}) (jsonutils.JSONObject, error) {
	log.Infof("mebs DeleteSnapshots")
	diskId, ok := params.(string)
	if !ok {
		return nil, hostutils.ParamsError
	}
	return nil, s.deleteSnapshot(diskId, "")
}
