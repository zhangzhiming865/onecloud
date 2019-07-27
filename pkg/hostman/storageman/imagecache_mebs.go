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
	"sync"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	"yunion.io/x/onecloud/pkg/hostman/hostutils"
	"yunion.io/x/onecloud/pkg/hostman/storageman/mebs"
	"yunion.io/x/onecloud/pkg/hostman/storageman/remotefile"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
)

type SMebsImageCache struct {
	imageId string
	cond    *sync.Cond
	Manager IImageCacheManger

	mebsImageCacheManager *SMebsImageCacheManager
	mebsStorage           *SMEBSStorage
	mebsStorageConf       *jsonutils.JSONDict
}

func NewMebsImageCache(imageId string, imagecacheManager IImageCacheManger) *SMebsImageCache {
	log.Infof("mebs NewMebsImageCache imageid %v", imageId)
	imageCache := new(SMebsImageCache)
	imageCache.imageId = imageId
	imageCache.Manager = imagecacheManager
	imageCache.cond = sync.NewCond(new(sync.Mutex))

	imageCache.mebsImageCacheManager = imageCache.Manager.(*SMebsImageCacheManager)
	imageCache.mebsStorage = imageCache.mebsImageCacheManager.storage.(*SMEBSStorage)
	imageCache.mebsStorageConf = imageCache.mebsStorage.GetStorageConf()
	return imageCache
}

func (r *SMebsImageCache) GetName() string {
	log.Infof("mebs GetName imagecache prefix is %v", r.mebsImageCacheManager.Prefix)
	return fmt.Sprintf("%s", r.imageId)
}

func (r *SMebsImageCache) GetPath() string {
	log.Infof("mebs GetPath, imagecachemanager path is %v", r.mebsImageCacheManager.GetPath())
	return ""
}

func (r *SMebsImageCache) Load() bool {
	log.Infof("mebs loading Mebs imagecache %s", r.GetName())
	hostname, _ := r.mebsStorageConf.GetString("mon_host")
	_, err := mebs.MebsImageInfo(r.GetName(), hostname)
	if err != nil {
		log.Infof("image name %v not exist", r.GetName())
		return false
	}
	return true
}

func (r *SMebsImageCache) Acquire(ctx context.Context, zone, srcUrl, format string) bool {
	log.Infof("mebs get local r.image %v zone %v, srcUrl %v, format %v", r.imageId, zone, srcUrl, format)
	if r.Load() {
		return true
	}

	localImageCache := storageManager.LocalStorageImagecacheManager.AcquireImage(ctx, r.imageId, zone, srcUrl, format)
	if localImageCache == nil {
		log.Errorf("failed to acquireimage %s ", r.imageId)
		return false
	}
	if !r.Load() {
		hostname, _ := r.mebsStorageConf.GetString("mon_host")
		log.Infof("mebs convert local image %s to mebs %s", r.imageId, r.GetName())
		success := mebs.UploadMebsImage(localImageCache.GetPath(), r.GetName(), hostname)
		if !success {
			log.Errorf("failed to upload image to mebs")
			return false
		}
	}
	return r.Load()
}

func (r *SMebsImageCache) Release() {
	log.Infof("mebs Release")
	return
}

func (r *SMebsImageCache) Remove(ctx context.Context) error {
	log.Infof("mebs Remove ImageCache %v", r.GetName())
	imageCacheManger := r.Manager.(*SMebsImageCacheManager)
	storage := imageCacheManger.storage.(*SMEBSStorage)
	if err := storage.deleteImage(r.GetName()); err != nil {
		return err
	}

	go func() {
		_, err := modules.Storagecachedimages.Detach(hostutils.GetComputeSession(ctx),
			r.Manager.GetId(), r.imageId, nil)
		if err != nil {
			log.Errorf("Fail to delete host cached image: %s", err)
		}
	}()
	return nil
}

func (r *SMebsImageCache) GetDesc() *remotefile.SImageDesc {
	log.Infof("mebs GetDesc")
	return nil
}

func (r *SMebsImageCache) GetImageId() string {
	log.Infof("mebs GetImageId")
	return r.imageId
}
