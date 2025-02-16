package app

import (
	"encoding/json"
	"fmt"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/log/ubiquitous/log/field"
	"github.com/mokanus/go-step/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type QRY struct {
	name       string
	r          *http.Request
	w          http.ResponseWriter
	channelUid string
	qryType    string
}

func NewQRY(r *http.Request, w http.ResponseWriter, channelUid string) *QRY {
	return &QRY{
		name:       r.RemoteAddr,
		r:          r,
		w:          w,
		channelUid: channelUid,
		qryType:    strings.TrimPrefix(r.URL.Path, "/"),
	}
}

func (self *QRY) Type() string {
	return self.qryType
}

func (self *QRY) Path() string {
	return self.r.URL.Path
}

func (self *QRY) Body() ([]byte, error) {
	return ioutil.ReadAll(self.r.Body)
}

func (self *QRY) IP() string {
	ipPort := strings.Split(self.r.RemoteAddr, ":")
	if len(ipPort) >= 1 {
		return ipPort[0]
	}
	return ""
}

func (self *QRY) Name() string {
	return self.name
}

func (self *QRY) ChannelUid() string {
	return self.channelUid
}

func (self *QRY) Read(msg interface{}) bool {
	if msg == nil {
		log.GetLogger().Info("[EXTER] <-", field.String("method", self.r.Method), field.String("path", self.r.URL.Path))
		return true
	}

	if self.r.Method == "GET" {
		if err := util.QueryValues(self.r.URL.Query(), msg); err != nil {
			log.GetLogger().Error(fmt.Sprintf("[EXTER] <- %s:%s：%v", self.r.Method, self.r.URL.Path, err))
			return false
		}
		log.GetLogger().Info(fmt.Sprintf("[EXTER] <- %s:%s %+v <= %s", self.r.Method, self.r.URL.Path, msg, self.r.URL.RawQuery))
	} else {
		body, err := ioutil.ReadAll(self.r.Body)
		if err != nil {
			log.GetLogger().Error(fmt.Sprintf("[EXTER] <- %s:%s", self.r.Method, self.r.URL.Path), field.Error(err))
			return false
		}
		if len(body) == 0 {
			log.GetLogger().Error(fmt.Sprintf("[EXTER] <- %s:%s：包体为空", self.r.Method, self.r.URL.Path))
			return false
		}

		if strings.Contains(self.r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			values, err := url.ParseQuery(string(body))
			if err != nil {
				log.GetLogger().Error(fmt.Sprintf("[EXTER] <- %s:%s：解析body(%s)为values失败", self.r.Method, self.r.URL.Path, string(body)), field.Error(err))
				return false
			}
			if err := util.QueryValues(values, msg); err != nil {
				log.GetLogger().Error(fmt.Sprintf("[EXTER] <- %s:%s：解析body(%s)成json失败", self.r.Method, self.r.URL.Path, string(body)), field.Error(err))
				return false
			}
		} else {
			if err := json.Unmarshal(body, msg); err != nil {
				log.GetLogger().Error(fmt.Sprintf("[EXTER] <- %s:%s：解析json(%s)失败", self.r.Method, self.r.URL.Path, string(body)))
				return false
			}
		}

		log.GetLogger().Info(fmt.Sprintf("[EXTER] <- %s:%s %+v <= %s(Content-Type:%s)", self.r.Method, self.r.URL.Path, msg, string(body), self.r.Header.Get("Content-Type")))
	}

	return true
}

func (self *QRY) SetHeader(key, value string) {
	self.w.Header().Set(key, value)
}

func (self *QRY) StartSession(username string) {
	session := &qryAuthSession{
		Token:    MakeRUID("", 1),
		Username: username,
		Expired:  util.NowUint32() + qryAuthSessionExpireSeconds,
	}

	// 服务端保存该token对应的session
	addQryAuthSession(session)

	// 客户端那里设置cookie来保存随机串
	http.SetCookie(self.w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: time.Now().Add(time.Duration(qryAuthSessionExpireSeconds) * time.Second),
	})
}

func (self *QRY) CheckSession() bool {
	// 只有dev和rel要验证账号
	if Env != EnvDev && Env != EnvRel {
		return true
	}

	// 从请求的Cookie里拿token
	c, err := self.r.Cookie("session_token")
	if err != nil || c == nil || c.Value == "" {
		return false
	}

	session := getQryAuthSession(c.Value)
	if session == nil {
		return false
	}

	if util.NowUint32() > session.Expired {
		return false
	}

	// 检查会话时，如果未过期，就对会话进行续期
	session.Expired = util.NowUint32() + qryAuthSessionExpireSeconds
	http.SetCookie(self.w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: time.Now().Add(time.Duration(qryAuthSessionExpireSeconds) * time.Second),
	})

	return true
}

// 返回字符串
func (self *QRY) RspPlain(format string, a ...interface{}) {
	text := fmt.Sprintf(format, a...)
	self.w.Write([]byte(text))
	log.GetLogger().Info("[EXTER] -> ", field.String("method", self.r.Method), field.String("path", self.r.URL.Path), field.String("text", text))
}

// 将结构体转为json后返回
func (self *QRY) RspJSON(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.GetLogger().Error("[EXTER] -> 返回结果编码异常", field.String("method", self.r.Method), field.String("path", self.r.URL.Path))
		self.w.Write([]byte("返回结果编码异常"))
		return
	}
	self.w.Write(data)
	log.GetLogger().Info("[EXTER] -> ", field.String("method", self.r.Method), field.String("path", self.r.URL.Path), field.String("data", string(data)))
}

// 返回文件内容
func (self *QRY) RspFile(filePath string, vars map[string]string) {
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.GetLogger().Error(fmt.Sprintf("文件不存在：%s", filePath))
		http.NotFound(self.w, self.r)
		return
	}

	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("文件%s读错误：%v", filePath, err))
		http.Error(self.w, err.Error(), 500)
		return
	}

	if fileInfo.IsDir() {
		log.GetLogger().Error(fmt.Sprintf("文件为目录：%s", filePath))
		http.NotFound(self.w, self.r)
		return
	}

	if vars != nil {
		bin, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.GetLogger().Error(fmt.Sprintf("文件不存在：%s", filePath))
			http.NotFound(self.w, self.r)
			return
		}
		content := string(bin)
		for name, value := range vars {
			content = strings.Replace(content, name, value, -1)
		}
		log.GetLogger().Debug(fmt.Sprintf("返回文件：%s", filePath))
		self.w.Write([]byte(content))
	} else {
		log.GetLogger().Debug(fmt.Sprintf("返回文件：%s", filePath))
		http.ServeFile(self.w, self.r, filePath)
	}
}

// 返回http格式的错误（http code固定用502：Bad Gateway）
func (self *QRY) RspError(text string) {
	log.GetLogger().Error(fmt.Sprintf("[EXTER] -> %s", text))
	http.Error(self.w, text, 502)
}

// 注意：以下的返回格式是游戏内的Post、星云请求、AdminJS共用的
// 当要返回成功数据时，字段为：
// status: 1
// msg: "success"
// data: interface{}
// 当要返回失败提示时，字段为：
// status: 0
// msg: string

func (self *QRY) RspSuccess(data interface{}) {
	result := make(map[string]interface{})
	result["status"] = 1
	result["msg"] = "success"
	result["data"] = data

	resultBytes, err := json.Marshal(result)
	if err != nil {
		self.RspFail("返回结果编码异常")
		return
	}

	log.GetLogger().Info(fmt.Sprintf("[EXTER] -> %s:%s %s", self.r.Method, self.r.URL.Path, string(resultBytes)))

	self.w.Write(resultBytes)
}

func (self *QRY) RspFail(msg string) {
	log.GetLogger().Error(fmt.Sprintf("[EXTER] -> %s:%s: %s", self.r.Method, self.r.URL.Path, msg))
	self.w.Write([]byte(fmt.Sprintf(`{"status":0,"msg":"%s"}`, msg)))
}
