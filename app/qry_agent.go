package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-step/log"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// 每一次qry请求会创建一个QryAgent来负责执行请求。
type QryAgent struct {
	err        error
	addr       string
	channelUid string
}

type QryResponse struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func (self *QryAgent) Post(path string, qryData interface{}, qryResp interface{}) error {
	qryBody, err := json.Marshal(qryData)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}

	req, err := http.NewRequest("POST", self.addr+"/"+path, bytes.NewReader(qryBody))
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}

	req.Header.Set("X-My-Qry-Admin-Token", qryAdminToken)
	if self.channelUid != "" {
		req.Header.Set("X-My-Qry-Channel-Uid", self.channelUid)
	}

	client := &http.Client{}
	client.Timeout = time.Second * 3
	resp, err := client.Do(req)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}

	if resp.StatusCode != 200 {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) 返回状态码为%d，错误内容为: %s", self.addr, path, resp.StatusCode, strings.TrimSpace(string(respBody))))
		return ErrServer
	}

	respWrap := new(QryResponse)
	respWrap.Data = qryResp
	if err := json.Unmarshal(respBody, respWrap); err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}

	// 如果返回status为0，代表出错了，出错信息在msg里
	// 如果返回status不为0，代表成功了，成功的数据在data里
	if respWrap.Status == 0 {
		return errors.New(respWrap.Msg)
	}

	return nil
}

func (self *QryAgent) PostBody(path string, qryBody []byte, qryResp interface{}) error {
	req, err := http.NewRequest("POST", self.addr+"/"+path, bytes.NewReader(qryBody))
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}

	req.Header.Set("X-My-Qry-Admin-Token", qryAdminToken)
	if self.channelUid != "" {
		req.Header.Set("X-My-Qry-Channel-Uid", self.channelUid)
	}

	client := &http.Client{}
	client.Timeout = time.Second * 3
	resp, err := client.Do(req)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}

	if resp.StatusCode != 200 {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) 返回状态码为%d，错误内容为: %s", self.addr, path, resp.StatusCode, strings.TrimSpace(string(respBody))))
		return ErrServer
	}

	respWrap := new(QryResponse)
	respWrap.Data = qryResp
	if err := json.Unmarshal(respBody, respWrap); err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] post(%v) Error: %v", self.addr, path, err))
		return ErrServer
	}

	// 如果返回status为0，代表出错了，出错信息在msg里
	// 如果返回status不为0，代表成功了，成功的数据在data里
	if respWrap.Status == 0 {
		return errors.New(respWrap.Msg)
	}

	return nil
}
