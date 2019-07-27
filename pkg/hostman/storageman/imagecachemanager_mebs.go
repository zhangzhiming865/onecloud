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
	"strings"
	"sync"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/hostman/hostutils"
)

type SMebsImageCacheManager struct {
	SBaseImageCacheManager
	Prefix  string
	storage IStorage
}

func NewMebsImageCacheManager(manager *SStorageManager, cachePath string, storage IStorage, storagecacheId string) *SMebsImageCacheManager {
	imageCacheManager := new(SMebsImageCacheManager)

	imageCacheManager.storagemanager = manager
	imageCacheManager.storagecacaheId = storagecacheId
	imageCacheManager.storage = storage

	imageCacheManager.Prefix = cachePath
	imageCacheManager.cachedImages = make(map[string]IImageCache, 0)
	imageCacheManager.mutex = new(sync.Mutex)
	imageCacheManager.loadCache()
	return imageCacheManager
}

type SMebsImageCacheManagerFactory struct {
}

func (factory *SMebsImageCacheManagerFactory) NewImageCacheManager(manager *SStorageManager, cachePath string, storage IStorage, storagecacheId string) IImageCacheManger {
	return NewMebsImageCacheManager(manager, cachePath, storage, storagecacheId)
}

func (factory *SMebsImageCacheManagerFactory) StorageType() string {
	return api.STORAGE_MEBS
}

func init() {
	registerimageCacheManagerFactory(&SMebsImageCacheManagerFactory{})
}

func (c *SMebsImageCacheManager) loadCache() {
	log.Infof("mebs loadCache")
	c.mutex.Lock()
	defer c.mutex.Unlock()
	storage := c.storage.(*SMEBSStorage)
	images, err := storage.listImages()
	if err != nil {
		log.Errorf("get storage %s images error; %v", c.storage.GetStorageName(), err)
		return
	}
	for _, image := range images {
		if strings.HasPrefix(image, c.Prefix) {
			imageId := strings.TrimPrefix(image, c.Prefix)
			c.LoadImageCache(imageId)
		} else {
			log.Debugf("find image %s from stroage %s", image, c.storage.GetStorageName())
		}
	}
}

func (c *SMebsImageCacheManager) LoadImageCache(imageId string) {
	log.Infof("mebs LoadImageCache")
	imageCache := NewMebsImageCache(imageId, c)
	if imageCache.Load() {
		c.cachedImages[imageId] = imageCache
	}
}

func (c *SMebsImageCacheManager) GetPath() string {
	log.Infof("mebs GetPath")
	return ""
}

func (c *SMebsImageCacheManager) PrefetchImageCache(ctx context.Context, data interface{}) (jsonutils.JSONObject, error) {
	body, ok := data.(*jsonutils.JSONDict)
	if !ok {
		return nil, hostutils.ParamsError
	}
	log.Infof("mebs PrefetchImageCache body %v", body)

	imageId, err := body.GetString("image_id")
	if err != nil {
		return nil, err
	}
	format, _ := body.GetString("format")
	srcUrl, _ := body.GetString("src_url")
	zone, _ := body.GetString("zone")

	cache := c.AcquireImage(ctx, imageId, zone, srcUrl, format)
	if cache == nil {
		log.Errorf("error mebs AcquireImage return nil")
		return nil, fmt.Errorf("error acquire image")
	}

	res := map[string]interface{}{
		"image_id": imageId,
		"path":     cache.GetPath(),
	}
	if desc := cache.GetDesc(); desc != nil {
		res["name"] = desc.Name
		res["size"] = desc.Size
	}
	return jsonutils.Marshal(res), nil
}

func (c *SMebsImageCacheManager) DeleteImageCache(ctx context.Context, data interface{}) (jsonutils.JSONObject, error) {
	log.Infof("mebs DeleteImageCache")
	body, ok := data.(*jsonutils.JSONDict)
	if !ok {
		return nil, hostutils.ParamsError
	}

	imageId, _ := body.GetString("image_id")
	return nil, c.removeImage(ctx, imageId)
}

func (c *SMebsImageCacheManager) removeImage(ctx context.Context, imageId string) error {
	log.Infof("mebs removeImage")
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if img, ok := c.cachedImages[imageId]; ok {
		delete(c.cachedImages, imageId)
		return img.Remove(ctx)
	}
	return nil
}

func (c *SMebsImageCacheManager) AcquireImage(ctx context.Context, imageId, zone, srcUrl, format string) IImageCache {
	log.Infof("mebs AcquireImage imageId %v zone %v srcUrl %v format %v", imageId, zone, srcUrl, format)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	img, ok := c.cachedImages[imageId]
	if !ok {
		img = NewMebsImageCache(imageId, c)
		c.cachedImages[imageId] = img
	}
	if img.Acquire(ctx, zone, srcUrl, format) {
		return img
	}
	return nil
}

func (c *SMebsImageCacheManager) ReleaseImage(imageId string) {
	log.Infof("mebs ReleaseImage")
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if img, ok := c.cachedImages[imageId]; ok {
		img.Release()
	}
}
