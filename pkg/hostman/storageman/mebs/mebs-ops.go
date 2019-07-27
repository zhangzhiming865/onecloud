/**
* @author
*    fuyuandi fuyuandi2008@163.com
*    zhangzhiming zhangzhiming865@163.com
* ©2019 fuyuandi and zhangzhiming. All Rights Reserved.
**/
package mebs

import (
	_ "github.com/pkg/errors"
	"yunion.io/x/log"

	"fmt"
	"time"
)

const (
	DefaultMebsSize = 1024 * 1024 * 1024
)

func UploadMebsImage(localFilePath, imageName, hostname string) bool {
	log.Infof("UploadMebsImage localFilePath %v imageName %v hostname %v", localFilePath, imageName, hostname)
	prog, err := UploadTemplate(hostname, TempleInfo{File_name: localFilePath, Description: "", Template_name: imageName})
	if err != nil {
		log.Infof("mebs error UploadTemplate, local file path %v, imageName %v", localFilePath, imageName)
		return false
	}
	// try for one hour
	time.Sleep(time.Second * 5)
	for i := 0; i < 3600; i++ {
		progSt, err := GetProgess(hostname, prog)
		if err != nil {
			log.Errorf("error get progress ret %v, imagename %v", err, imageName)
			return false
		}
		time.Sleep(time.Second)
		if progSt.Progress_status == 100 {
			return true
		}
	}
	log.Infof("mebs upload Template failed for too long time")
	return false
}

func MebsImageInfo(imageName, hostname string) (VolInfo_t, error) {
	return GetVol(hostname, imageName, false)
}

func MebsImageSize(imageName, hostname string) int64 {
	info, err := MebsImageInfo(imageName, hostname)
	if err != nil {
		log.Infof("error get image info by hostname %v", hostname)
		return 0
	}
	return info.Virtual_size
}

func MebsResizeImage(imageName string, sizeMb int64, hostname string) error {
	_, err := Resize(hostname, imageName, sizeMb*1024*1024, false)
	return err
}

func MebsRemoveImage(imageName, hostname string) error {
	_, err := RemoveVol(hostname, imageName, false)
	return err
}

func MebsCreateVolumeFromImage(imageName, diskName, hostname string) error {
	_, err := CreateVolume(hostname, diskName, 0, imageName)
	return err
}

func MebsCreateSnapshot(diskName, snapName, hostname string) error {
	return fmt.Errorf("error Mebs not Implement")
}

func MebsCreateImage(name string, sizeMb int64, hostname string) error {
	_, err := CreateVolume(hostname, name, sizeMb*1024*1024, "")
	return err
}

func MebsDeleteSnapshot(diskName, snapName, hostname string) error {
	return fmt.Errorf("error Mebs not Implement")
}
