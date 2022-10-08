package task

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// MMUE 平台登录用户名密码
type mmueLoginInfo struct {
	Username string `json:"username" form:"username" bson:"username" gorm:"username"`
	Password string `json:"pwd" form:"pwd" bson:"pwd" gorm:"pwd"`
}

// MMUE 平台登录信息
type MMUE struct {
	AuthUrl   string `json:"auth_rul" form:"auth_rul" gorm:"auth_rul"`
	TestUrl   string `json:"test_url" form:"test_url" gorm:"test_url"`
	Token     string
	LoginInfo *mmueLoginInfo
}

// MMUE 平台登录方法
func (m *MMUE) mLogin() {
	newHeaderByte, _ := json.Marshal(m.LoginInfo)
	payload := strings.NewReader(string(newHeaderByte))

	req, _ := http.NewRequest("POST", m.AuthUrl+"/mmue/api/login", payload)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	if res.StatusCode == 200 {
		body, _ := ioutil.ReadAll(res.Body)
		bodyString := string(body)
		response := mLoginResponse{}
		err := json.Unmarshal([]byte(bodyString), &response)
		if err != nil {
			return
		}

		m.Token = response.Data.Token
	}

	defer res.Body.Close()
}

// MMUE 平台登录接口响应结构体
type mLoginResponse struct {
	Code int `json:"code"`
	Data struct {
		Data struct {
			UserId     int         `json:"user_id"`
			UserName   string      `json:"user_name"`
			UserPower  string      `json:"user_power"`
			TenantId   string      `json:"tenant_id"`
			TenantName string      `json:"tenant_name"`
			TenantLogo string      `json:"tenant_logo"`
			IsRocuser  string      `json:"is_rocuser"`
			LibValue   interface{} `json:"lib_value"`
			AgentId    interface{} `json:"agent_id"`
		} `json:"data"`
		Token string `json:"token"`
	} `json:"data"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// HTTPReqInfo http请求信息结构体
type HTTPReqInfo struct {
	Method  string
	Url     string
	Payload io.Reader
}

// MMUE平台发起http请求
func (m *MMUE) mRequest(mReq HTTPReqInfo) []byte {
	if m.Token == "" {
		m.mLogin()
	}
	req, _ := http.NewRequest(mReq.Method, m.TestUrl+mReq.Url, mReq.Payload)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", m.Token)
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Origin", m.TestUrl)
	req.Header.Add("Referer", m.TestUrl+"/app/client")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == 401 {
		m.mLogin()
		return m.mRequest(mReq)
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)

	body, _ := ioutil.ReadAll(res.Body)
	return body
}

// 知识图谱会话接口
func (m *MMUE) mChat(spaces, query string) *chatResponse {
	//spaces [{"space_name":"common_kg"},{"space_name":"shici_1549937929118056448"}]
	r := HTTPReqInfo{
		Method:  "POST",
		Url:     "/graph/kgqa/v1/chat",
		Payload: strings.NewReader(fmt.Sprintf(`{"spaces": %s, "question": "%s"}`, spaces, query)),
	}
	if m.TestUrl != m.AuthUrl {
		r.Url = "/kgqa/v1/chat"
	}
	var c chatResponse
	err := json.Unmarshal(m.mRequest(r), &c)
	if err != nil {
		return nil
	}
	return &c
}

// 知识图谱会话响应结构体
type chatResponse struct {
	Code int `json:"code"`
	Data struct {
		Type       string `json:"@type"`
		EntityName string `json:"entity_name"`
		Disambi    string `json:"disambi"`
		Answer     string `json:"answer"`
		Attr       struct {
			Describ string `json:"describ"`
		} `json:"attr"`
		Source  string `json:"source"`
		TraceId string `json:"trace_id"`
	} `json:"data"`
}
