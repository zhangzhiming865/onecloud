/**
* @author
*    fuyuandi fuyuandi2008@163.com
*    zhangzhiming zhangzhiming865@163.com
* ©2019 fuyuandi and zhangzhiming. All Rights Reserved.
**/
package mebs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"encoding/json"

	//"yunion.io/x/jsonutils"
	"yunion.io/x/log"

	//"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/util/httputils"
)

type VolInfo_t struct {
	Volume_id     int64  `json:"volume_id"`
	Created_time  string `json:"create_time"`
	Deleted_time  string `json:"delete_time"`
	Description   string `json:"description"`
	Format        string `json:"format"`
	Id            int64  `json:"id"`
	Modified_time string `json:"modified_time"`
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	Status        string `json:"status"`
	Struct_       string `json:struct"`
	Virtual_size  int64  `json:"virtual_size"`
}

/* 创建卷
* 输入：
* host: ip:port
* vol_name:卷名称
* size 大小
* template_name 　可为空串
*  返回值volinfo_t.volume_id 卷id
 */
func CreateVolume(host string, vol_name string, size int64, template_name string) (VolInfo_t, error) {
	url := "http://"
	url += host
	url += "/volumes"
	postmap := map[string]interface{}{
		"name":        vol_name,
		"size":        size,
		"description": "",
	}
	if len(template_name) > 0 {
		postmap["template_name"] = template_name
	}
	result := VolInfo_t{
		Volume_id: -1,
	}
	jsonstr, err1 := json.Marshal(postmap)
	if err1 != nil {
		log.Errorf("postmap to json failed %v", err1)
		return result, fmt.Errorf("marshal failed")
	}
	req, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(jsonstr)))
	if httperr != nil {
		log.Errorf("http post failed %v", httperr)
		return result, httperr
	}
	if req.StatusCode != http.StatusOK {
		log.Errorf("http status error %v", req.StatusCode)
		return result, fmt.Errorf("http status error")
	}
	defer req.Body.Close()
	body, reqerr := ioutil.ReadAll(req.Body)
	if reqerr != nil {
		log.Errorf("read body failed %v %s", reqerr, string(body))
		return result, fmt.Errorf("read body failed")
	}
	var resp map[string]interface{}
	err2 := json.Unmarshal(body, &resp)
	if err2 != nil {
		log.Errorf("pares result string to json failed %v", string(body))
		return result, fmt.Errorf("pares result string to json failed")
	}
	_, ok := resp["volume_id"]
	if !ok {
		return result, fmt.Errorf("not volume_id found in body")
	}
	// log.Infof("get result is %v, type %v", string(body), reflect.TypeOf(resp["volume_id"]))
	value, ok2 := resp["volume_id"].(float64)
	if ok2 != true {
		log.Errorf("not volume_id int found")
		return result, fmt.Errorf("not volume_id int found")
	}
	result.Volume_id = int64(value)
	return result, nil
}
func GetVol(host string, ident string, by_id bool) (VolInfo_t, error) {
	url := "http://"
	url += host
	url += "/volumes/"
	url += ident
	url += "?by_id="
	url += strconv.FormatBool(by_id)
	var result VolInfo_t
	req, httperr := httputils.GetDefaultClient().Get(url)
	if httperr != nil {
		log.Errorf("query vol failed %v:", ident)
		return result, fmt.Errorf("query vol failed")
	}
	if req.StatusCode != http.StatusOK {
		log.Errorf("http status error %v", req.StatusCode)
		return result, fmt.Errorf("http status error")
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("read req body faild")
		return result, fmt.Errorf("parse get result failed")
	}
	err2 := json.Unmarshal(body, &result)
	if err2 != nil {
		log.Errorf("Unmarshal result failed %v", string(body))
		return result, fmt.Errorf("Unmarshal result failed")
	}
	return result, nil

}
func GetVols(host string) ([]VolInfo_t, error) {
	url := "http://"
	url += host
	url += "/volumes"
	var result []VolInfo_t
	req, httperr := httputils.GetDefaultClient().Get(url)
	if httperr != nil {
		log.Errorf("query vols failed :", host)
		return result, fmt.Errorf("query vol failed")
	}
	if req.StatusCode != http.StatusOK {
		log.Errorf("http status error %v", req.StatusCode)
		return result, fmt.Errorf("http status error")
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("read req body faild")
		return result, fmt.Errorf("parse get result failed")
	}
	err2 := json.Unmarshal(body, &result)
	if err2 != nil {
		log.Errorf("Unmarshal result failed %v", string(body))
		return result, fmt.Errorf("Unmarshal result failed")
	}
	return result, nil
}

/*
*  返回值：new size of volume
 */
func Resize(host string, ident string, new_size int64, by_id bool) (int64, error) {
	url := "http://"
	url += host
	url += "/volumes/"
	url += ident
	url += "/"
	url += "resize"
	url += "?by_id="
	url += strconv.FormatBool(by_id)
	postmap := map[string]interface{}{
		"new_size": new_size,
	}
	poststr, err1 := json.Marshal(postmap)
	if err1 != nil {
		log.Errorf("tranlate jsonmap to string failed %v", postmap)
		return -1, fmt.Errorf("tranlate jsonmap to string failed")
	}
	req, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(poststr)))
	if httperr != nil {
		log.Errorf("http error %v, url %v map %v", httperr.Error(), url, postmap)
		return -1, fmt.Errorf("http error")
	}
	if req.StatusCode != http.StatusOK {
		log.Errorf("http status error %v", req.StatusCode)
		return -1, fmt.Errorf("http status error")
	}
	defer req.Body.Close()
	var resp map[string]interface{}
	body, err2 := ioutil.ReadAll(req.Body)
	if err2 != nil {
		return -1, fmt.Errorf(err2.Error())
	}
	err3 := json.Unmarshal(body, &resp)
	if err3 != nil {
		return -1, fmt.Errorf(err3.Error())
	}
	_, iscontain := resp["new_size"]
	if !iscontain {
		return -1, fmt.Errorf("not new_size found")
	}
	value, ok := resp["new_size"].(int64)
	if !ok {
		return -1, fmt.Errorf("no new_size int64 found")
	}
	return value, nil

}

type Launch_t struct {
	Host        string `json:"-"`
	Vol_id      int64  `json:"volume_id"`
	Vol_name    string `json:"vol_name"`
	By_id       bool   `json:"by_id"`
	Protocol    string `json:"protocol"`
	Mount_point string `json:"-"`
	Connect     bool   `json:"connect"`
	Restart     bool   `json:"-"`
}
type Con_t struct {
	Device string `json:"device"`
	Url    string `json:"url"`
}

/**
* 输入：
* launch.protocol 协议
* launch.connect   true
* launch.by_id  set to  choose volume id or name
* launch.vol_id or launch.vol_name
* 输出：
* con_t.device
* con_t.url
**/
func LaunchVol(launch Launch_t) (Con_t, error) {
	url := "http://"
	url += launch.Host
	url += "/volumes/"
	if launch.By_id {
		url += string(launch.Vol_id)
	} else {
		url += string(launch.Vol_name)
	}
	url += "/launch?by_id="
	url += strconv.FormatBool(launch.By_id)
	post_map := map[string]interface{}{
		"protocol": launch.Protocol,
		"connect":  launch.Connect,
	}
	var conn Con_t
	poststr, err1 := json.Marshal(post_map)
	if err1 != nil {
		log.Errorf("map to string failed")
		return conn, fmt.Errorf("err:%v", err1)
	}
	req, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(poststr)))
	if httperr != nil {
		log.Errorf("http error :%v", httperr)
		return conn, fmt.Errorf("http error")
	}
	if req.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", req.StatusCode)
		return conn, fmt.Errorf("http status error")
	}
	defer req.Body.Close()
	body, err2 := ioutil.ReadAll(req.Body)
	if err2 != nil {
		log.Errorf("read body stream failed %v", err2)
		return conn, fmt.Errorf("read body stream failed")

	}
	err3 := json.Unmarshal(body, &conn)
	if err3 != nil {
		log.Errorf("parse string to json failed %v", string(body))
		return conn, fmt.Errorf("parse string to json failed")
	}
	return conn, nil
}

func Disconnect(host string, ident string, by_id bool) error {
	url := "http://"
	url += host
	url += "/volumes/"
	url += ident
	url += "/disconnect?by_id="
	url += strconv.FormatBool(by_id)
	postmap := map[string]interface{}{
		"kill": false,
	}
	poststr, err1 := json.Marshal(postmap)
	if err1 != nil {
		log.Errorf("tanslate map to string failed %v", err1)
		return fmt.Errorf("tanslate map to string failed")
	}
	req, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(poststr)))
	if httperr != nil {
		log.Errorf("http error:%v", httperr)
		return fmt.Errorf("http error")
	}
	if req.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", req.StatusCode)
		return fmt.Errorf("http status error")
	}
	defer req.Body.Close()
	_, err2 := ioutil.ReadAll(req.Body)
	if err2 != nil {
		log.Errorf("read body faild %v", err2)
		return fmt.Errorf("read body faild %v", err2)
	}
	return nil
}

func RemoveVol(host string, ident string, by_id bool) (int64, error) {
	url := "http://"
	url += host
	url += "/volumes/"
	url += ident
	url += "?by_id="
	url += strconv.FormatBool(by_id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Errorf("new request faild %v", err)
		return -1, fmt.Errorf("NewRequest error:%v", err)
	}
	res, httperr := http.DefaultClient.Do(req)
	if httperr != nil {
		log.Errorf("Http request failed :%v", httperr)
		return -1, fmt.Errorf("Http request failed %v", httperr)
	}
	if res.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", res.StatusCode)
		return -1, fmt.Errorf("http status error")
	}
	defer res.Body.Close()
	body, err1 := ioutil.ReadAll(res.Body)
	if err1 != nil {
		log.Errorf("read stream failed %v", err1)
		return -1, fmt.Errorf("read stream failed %v", err1)
	}
	var result map[string]interface{}
	err2 := json.Unmarshal(body, &result)
	if err2 != nil {
		log.Errorf("parse result to json faild %v", err2)
		return -1, fmt.Errorf("parse result to json faild %v", err2)
	}
	_, ok := result["volume_id"]
	if !ok {
		log.Errorf("not volume_id found in response")
		return -1, fmt.Errorf("not volume_id found in response")
	}
	value, ok2 := result["volume_id"].(float64)
	if !ok2 {
		log.Errorf("no volume_id int64 found int response")
		return -1, fmt.Errorf("no volume_id int64 found in response")
	}
	return int64(value), nil
}

type TempleInfo struct {
	File_name     string `json:"file_name"`
	Template_name string `json:"template_name"`
	Description   string `json:"description"`
	Template_id   int64  `json:"template_id"`
	Size          int64  `json:"size"`
	Volume_id     int64  `json:"volume_id"`
	Volume_name   string `json:volume_name`
}
type Progress_t struct {
	Progress_key string `json:"progress_key"`
}
type ProgressSt_t struct {
	Progress_status int    `json:"progress_status"`
	Progress_host   string `json:"progress_host"`
}

/*
* 输入：
* temp_info.file_name  本地文件名
* temp_info.template_name
* temp_info.description
* 输出：
* progress_key
 */
func UploadTemplate(host string, temp_info TempleInfo) (Progress_t, error) {
	url := "http://"
	url += host
	url += "/upload_templates"
	param_json := map[string]interface{}{
		"file_name":     temp_info.File_name,
		"description":   temp_info.Description,
		"template_name": temp_info.Template_name,
	}
	var result Progress_t
	param_str, err1 := json.Marshal(param_json)
	if err1 != nil {
		log.Errorf("tanslate map to string failed %v", err1)
		return result, fmt.Errorf("tanslate map to string failed")
	}
	resp, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(param_str)))
	if httperr != nil {
		log.Errorf("httperr %v", httperr)
		return result, fmt.Errorf("httperr")
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", resp.StatusCode)
		return result, fmt.Errorf("http status error")
	}
	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		log.Errorf("read stream failed %v", err1)
		return result, fmt.Errorf("read stream failed")
	}
	err3 := json.Unmarshal(body, &result)
	if err3 != nil {
		log.Errorf("Unmarshal response text faild %v", err3)
		return result, fmt.Errorf("Unmarshal response text faild")
	}
	return result, nil
}

/*
* 输入：
* temp_info.template_id
* temp_info.file_name
* 输出：
*  progress_key
 */
func DownloadTemplate(host string, temp_info TempleInfo) (Progress_t, error) {
	json_params := map[string]interface{}{
		"template_id": temp_info.Template_id,
		"file_name":   temp_info.File_name,
	}
	var progress Progress_t
	params_str, err1 := json.Marshal(json_params)
	if err1 != nil {
		log.Errorf("parse map to str faild")
		return progress, fmt.Errorf("parse map to str faild")
	}
	url := "http://"
	url += host
	url += "/download_templates"
	res, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(params_str)))
	if httperr != nil {
		log.Errorf("httperr:%v", httperr)
		return progress, fmt.Errorf("http error")
	}
	if res.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", res.StatusCode)
		return progress, fmt.Errorf("http status error")
	}
	defer res.Body.Close()
	body, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		log.Errorf("Read res stream error:%v", err2)
		return progress, fmt.Errorf("Read response stream error")
	}
	err3 := json.Unmarshal(body, &progress)
	if err3 != nil {
		log.Errorf("parse response text to json error:%v", err3)
		return progress, fmt.Errorf("parse response text to json error")
	}
	return progress, nil
}

/*
* 输入：
* temp_info.volume_id
* temp_info.template_name
* temp_info.size
* temp_info.description
 */
func CreateTemplate(host string, temp_info TempleInfo) (Progress_t, error) {
	json_params := map[string]interface{}{
		"volume_id":     temp_info.Volume_id,
		"template_name": temp_info.Template_name,
		"size":          temp_info.Size,
		"description":   temp_info.Description,
	}
	var progress Progress_t
	params_str, _ := json.Marshal(json_params)
	url := "http://"
	url += host
	url += "/templates"
	resp, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(params_str)))
	if httperr != nil {
		log.Errorf("httperr:%v", httperr)
		return progress, fmt.Errorf("http error")
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", resp.StatusCode)
		return progress, fmt.Errorf("http status error")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err := json.Unmarshal(body, &progress)
	return progress, err
}

/*
*输入：
*temp_info.Volume_id / temp_info.Volume_name
*输出：
* progress
 */
func UpdateTemplate(host string, temp_info TempleInfo) (Progress_t, error) {
	url := "http://"
	url += host
	url += "/volumes/"
	if len(temp_info.Volume_name) <= 0 {
		url += string(temp_info.Volume_id)
		url += "/commit?by_id=true"
	} else {
		url += string(temp_info.Volume_name)
		url += "/commit?by_id=false"
	}
	var progress Progress_t
	resp, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(""))
	if httperr != nil {
		log.Errorf("httperr:%v", httperr)
		return progress, fmt.Errorf("http error")
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", resp.StatusCode)
		return progress, fmt.Errorf("http status error")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err := json.Unmarshal(body, &progress)
	return progress, err
}

/*
* dst_temp_name 模板名称
*
 */
func MigrateTemplate(host string, dst_temp_name string, dst_manager string, ident string, by_id bool) (Progress_t, error) {
	json_params := map[string]string{
		"dst_template_name": dst_temp_name,
		"dst_manager_add":   dst_manager,
	}
	params_str, _ := json.Marshal(json_params)
	url := "http://"
	url += host
	url += "/templates/"
	url += ident
	url += "/migrate?by_id="
	url += strconv.FormatBool(by_id)
	var progress Progress_t
	resp, httperr := httputils.GetDefaultClient().Post(url, "application/json", strings.NewReader(string(params_str)))
	if httperr != nil {
		log.Errorf("httperr:%v", httperr)
		return progress, fmt.Errorf("http error")
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", resp.StatusCode)
		return progress, fmt.Errorf("http status error")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err := json.Unmarshal(body, &progress)
	return progress, err
}
func SyncTemplate(host string, dst_template_name string, dst_manager string) (Progress_t, error) {
	var progress Progress_t
	return progress, nil
}
func DeleteTemplate(host string, ident string, by_id bool) (int64, error) {
	url := "http://"
	url += host
	url += "/templates/"
	url += ident
	url += "?by_id="
	url += strconv.FormatBool(by_id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Errorf("new request faild %v", err)
		return -1, fmt.Errorf("NewRequest error:%v", err)
	}
	res, httperr := http.DefaultClient.Do(req)
	if httperr != nil {
		log.Errorf("Http request failed :%v", httperr)
		return -1, fmt.Errorf("Http request failed %v", httperr)
	}
	if res.StatusCode != http.StatusOK {
		log.Errorf("http status error %s", res.StatusCode)
		return -1, fmt.Errorf("http status error")
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	var result map[string]int64
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Errorf("Parse body failed %v", err)
		return -1, fmt.Errorf("Parse body failed")
	}
	v, ok := result["template_id"]
	if !ok {
		log.Errorf("No template_id found", string(body))
		return -1, fmt.Errorf("No template_id found")
	}
	return v, nil
}
func GetProgess(host string, progress Progress_t) (ProgressSt_t, error) {
	url := "http://"
	url += host
	url += "/progress/"
	url += progress.Progress_key
	var result ProgressSt_t
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("new request faild %v", err)
		return result, fmt.Errorf("NewRequest error:%v", err)
	}
	res, httperr := http.DefaultClient.Do(req)
	if httperr != nil {
		log.Errorf("Http request failed :%v", httperr)
		return result, fmt.Errorf("Http request failed %v", httperr)
	}
	if res.StatusCode != http.StatusOK {
		log.Errorf("http status error %d", res.StatusCode)
		return result, fmt.Errorf("http status error")
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	log.Infof("progress get body %v, key %v", string(body), progress.Progress_key)
	var pro map[string]string
	err = json.Unmarshal(body, &pro)
	if err != nil {
		log.Errorf("Parse body failed %v", err)
		return result, fmt.Errorf("Parse body failed")
	}
	var progress_host string
	var progress_status string
	var ok bool
	progress_host, ok = pro["progress_host"]
	progress_status, ok = pro["progress_status"]
	if !ok {
		log.Errorf("no progress_status found")
		return result, fmt.Errorf("no progress_status found")
	}
	result.Progress_host = progress_host
	result.Progress_status, err = strconv.Atoi(strings.Split(progress_status, "%")[0])
	return result, err
}
