package main

type Config struct {
	BaseUrl  string `json:"baseUrl"`
	App      string `json:"app"`
	Server   string `json:"server"`
	Comment  string `json:"comment"`
	Filename string `json:"filename"`
}

type FindServerRsp struct {
	RetCode int    `json:"ret_code"`
	ErrMsg  string `json:"err_msg"`
	Data    []struct {
		Id           int    `json:"id"`
		Application  string `json:"application"`
		ServerName   string `json:"server_name"`
		NodeName     string `json:"node_name"`
		ServerType   string `json:"server_type"`
		EnableSet    bool   `json:"enable_set"`
		SettingState string `json:"setting_state"`
		PresentState string `json:"present_state"`
		PatchTime    string `json:"patch_time"`
	} `json:"data"`
}

type UploadFileRsp struct {
	RetCode int    `json:"ret_code"`
	ErrMsg  string `json:"err_msg"`
	Data    struct {
		Id     int    `json:"id"`
		Server string `json:"server"`
		Tgz    string `json:"tgz"`
	} `json:"data"`
}

type Parameters struct {
	BakFlag    bool   `json:"bak_flag"`
	PatchId    string `json:"patch_id"`
	UpdateText string `json:"update_text"`
}

type Item struct {
	Command    string     `json:"command"`
	Parameters Parameters `json:"parameters"`
	ServerId   string     `json:"server_id"`
}

type AddTaskReq struct {
	Items  []Item `json:"items"`
	Serial bool   `json:"serial"`
}

type AddTaskRsp struct {
	RetCode int    `json:"ret_code"`
	ErrMsg  string `json:"err_msg"`
	Data    string `json:"data"`
}

type CheckStatusRsp struct {
	RetCode int    `json:"ret_code"`
	ErrMsg  string `json:"err_msg"`
	Data    struct {
		Items []struct {
			Serial bool   `json:"serial"`
			Status int    `json:"status"`
			TaskId string `json:"task_no"`
		} `json:"items"`
	} `json:"data"`
}