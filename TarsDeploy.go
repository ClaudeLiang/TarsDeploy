package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"mime/multipart"
	"strconv"
	"time"
)

var (
	logger Logger
	cfg    Config
)

func findServer(app string, server string) (int, error) {
	logger.Infof("\n【获取服务列表】")
	b := []byte("")
	query := fmt.Sprintf(`tree_node_id=1%s.5%s`, app, server)
	res, err := HttpRequest("GET", cfg.BaseUrl + "/pages/server/api/server_list?" + query, &b, &map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		logger.Errorf("\t获取服务列表失败:%s", err.Error())
		return -1, err
	}
	//logger.Debugf("\t获取服务列表结果:%s", res)
	var rsp FindServerRsp
	err = json.Unmarshal([]byte(res), &rsp)
	if err != nil {
		logger.Errorf("\t获取服务列表结果格式不合法:%s", err.Error())
		return -1, err
	}
	if rsp.RetCode != 200 {
		logger.Errorf("\t%s", rsp.ErrMsg)
		return -1, errors.New(rsp.ErrMsg)
	}
	if len(rsp.Data) == 0 {
		logger.Errorf("\t获取服务列表为空")
		return -1, errors.New("获取服务列表为空")
	}
	svr := rsp.Data[0]
	logger.Infof("\t应用名:%s\n\t服务名:%s\n\t节点:%s\n\t服务类型:%s\n\t启用set:%t\n\t设置状态:%s\n\t实时状态:%s\n\t发布时间:%s",
		svr.Application, svr.ServerName, svr.NodeName, svr.ServerType, svr.EnableSet, svr.SettingState, svr.PresentState, svr.PatchTime)
	return svr.Id, nil
}

func uploadFile(app string, server string, file string, comment string) (int, error) {
	logger.Infof("\n【上传文件】")
	taskId := strconv.Itoa(int(time.Now().Unix())) + "000"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Errorf("\t找不到文件:%s", file)
		return -1, err
	}
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	_ = w.WriteField("application", app)
	_ = w.WriteField("module_name", server)
	_ = w.WriteField("comment", comment)
	_ = w.WriteField("task_id", taskId)
	fw, _ := w.CreateFormFile("suse", file)
	_, err = fw.Write(buf)
	if err != nil {
		logger.Errorf("\t写入文件到formData失败:%s", err.Error())
		return -1, err
	}
	_ = w.Close()
	b := body.Bytes()
	contentType := w.FormDataContentType()
	res, err := HttpRequest("POST", cfg.BaseUrl + "/pages/server/api/upload_patch_package", &b, &map[string]string{
		"Content-Type": contentType,
	})
	if err != nil {
		logger.Errorf("\t上传文件到tars平台失败:%s", err.Error())
		return -1, err
	}
	//logger.Debugf("\t上传结果:%s", res)
	var rsp UploadFileRsp
	err = json.Unmarshal([]byte(res), &rsp)
	if err != nil {
		logger.Errorf("\t上传文件结果格式不合法:%s", err.Error())
		return -1, err
	}
	if rsp.RetCode != 200 {
		logger.Errorf("\t%s", rsp.ErrMsg)
		return -1, errors.New(rsp.ErrMsg)
	}
	data := rsp.Data
	logger.Infof("\t服务:%s\n\t上传包:%s",
		data.Server, data.Tgz)
	return data.Id, nil
}

func addTask(svrId int, uploadId int) (string, error) {
	logger.Infof("\n【创建发布任务】")
	req := AddTaskReq{
		Items: []Item{{
			ServerId: strconv.Itoa(svrId),
			Command:  "patch_tars",
			Parameters: Parameters{
				BakFlag:    false,
				PatchId:    strconv.Itoa(uploadId),
				UpdateText: "",
			},
		}},
		Serial: true,
	}
	b, _ := json.Marshal(&req)
	res, err := HttpRequest("POST", cfg.BaseUrl + "/pages/server/api/add_task", &b, &map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		logger.Errorf("\t创建发布任务失败:%s\n", err.Error())
		return "", err
	}
	var rsp AddTaskRsp
	err = json.Unmarshal([]byte(res), &rsp)
	if err != nil {
		logger.Errorf("\t创建发布任务结果格式不合法:%s", err.Error())
		return "", err
	}
	if rsp.RetCode != 200 {
		logger.Errorf("\t%s", rsp.ErrMsg)
		return "", errors.New(rsp.ErrMsg)
	}
	logger.Debugf("\t创建发布任务成功，任务ID:%s\n", rsp.Data)
	return rsp.Data, nil
}

func checkStatus(taskId string) (bool, error) {
	logger.Infof("检查发布状态...")
	b := []byte("")
	query := fmt.Sprintf(`task_no=%s`, taskId)
	res, err := HttpRequest("GET", cfg.BaseUrl + "/pages/server/api/task?" + query, &b, &map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		logger.Errorf("\t检查发布状态失败:%s", err.Error())
		return false, err
	}
	//logger.Debugf("\t上传结果:%s", res)
	var rsp CheckStatusRsp
	err = json.Unmarshal([]byte(res), &rsp)
	if err != nil {
		logger.Errorf("\t检查发布状态结果格式不合法:%s", err.Error())
		return false, err
	}
	if rsp.RetCode != 200 {
		logger.Errorf("\t%s", rsp.ErrMsg)
		return false, errors.New(rsp.ErrMsg)
	}
	items := rsp.Data.Items
	for _, item := range items {
		if item.TaskId != taskId {
			continue
		}
		if item.Status == 2 {
			logger.Infof("\t\n发布成功!")
			return true, nil
		} else if item.Status == 3 {
			logger.Infof("\t\n发布失败!")
			return true, nil
		}
	}
	return false, nil
}

func main() {
	logger = L{}
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		logger.Errorf("读取配置文件错误:%s", err.Error())
		return
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		logger.Errorf("配置文件格式错误:%s", err.Error())
		return
	}
	if svrId, err := findServer(cfg.App, cfg.Server); err == nil {
		if uploadId, err := uploadFile(cfg.App, cfg.Server, cfg.Filename, cfg.Comment); err == nil {
			if taskId, err := addTask(svrId, uploadId); err == nil {
				ticker := time.NewTicker(time.Second)
				result := make(chan int)
				go func() {
					for range ticker.C {
						if status, err := checkStatus(taskId); err == nil {
							if status {
								ticker.Stop()
								result<-1
							}
						}
					}
				}()
				<-result
			}
			return
		}
	}
	logger.Errorf("\n【操作失败】")
}
