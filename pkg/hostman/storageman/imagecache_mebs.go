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
	"sync"

	"yunion.io/x/log"

	"yunion.io/x/onecloud/pkg/hostman/hostutils"
	"yunion.io/x/onecloud/pkg/hostman/options"
	"yunion.io/x/onecloud/pkg/hostman/storageman/remotefile"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/util/procutils"
	"yunion.io/x/onecloud/pkg/util/qemuimg"
	"yunion.io/x/onecloud/pkg/util/qemutils"
)

type SMebsImageCache struct {
	imageId string
	cond    *sync.Cond
	Manager IImageCacheManger
}

func NewMebsImageCache(imageId string, imagecacheManager IImageCacheManger) *SMebsImageCache {
	imageCache := new(SMebsImageCache)
	imageCache.imageId = imageId
	imageCache.Manager = imagecacheManager
	imageCache.cond = sync.NewCond(new(sync.Mutex))
	return imageCache
}

func (r *SMebsImageCache) GetName() string {
	imageCacheManger := r.Manager.(*SMebsImageCacheManager)
	return fmt.Sprintf("%s%s", imageCacheManger.Prefix, r.imageId)
}

func (r *SMebsImageCache) GetPath() string {
	imageCacheManger := r.Manager.(*SMebsImageCacheManager)
	storage := imageCacheManger.storage.(*SMEBSStorage)
	return fmt.Sprintf("rbd:%s/%s%s", imageCacheManger.GetPath(), r.GetName(), storage.getStorageConfString())
}

func (r *SMebsImageCache) Load() bool {
	log.Debugf("loading Mebs imagecache %s", r.GetPath())
	origin, err := qemuimg.NewQemuImage(r.GetPath())
	if err != nil {
		return false
	}
	return origin.IsValid()
}

func (r *SMebsImageCache) Acquire(ctx context.Context, zone, srcUrl, format string) bool {
	localImageCache := storageManager.LocalStorageImagecacheManager.AcquireImage(ctx, r.imageId, zone, srcUrl, format)
	if localImageCache == nil {
		log.Errorf("failed to acquireimage %s ", r.imageId)
		return false
	}
	if !r.Load() {
		log.Debugf("convert local image %s to rbd pool %s", r.imageId, r.Manager.GetPath())

		_, err := procutils.NewCommand(qemutils.GetQemuImg(), "convert", "-O", "raw", localImageCache.GetPath(), r.GetPath()).Run()
		if err != nil {
			log.Errorf("failed to convert image %s", options.HostOptions.ServersPath)
			return false
		}
	}
	return r.Load()
}

func (r *SMebsImageCache) Release() {
	return
}

func (r *SMebsImageCache) Remove(ctx context.Context) error {
	imageCacheManger := r.Manager.(*SMebsImageCacheManager)
	storage := imageCacheManger.storage.(*SMEBSStorage)
	if err := storage.deleteImage(r.Manager.GetPath(), r.GetName()); err != nil {
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
	return nil
}

func (r *SMebsImageCache) GetImageId() string {
	return r.imageId
}
