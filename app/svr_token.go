package app

import (
	"fmt"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/util"
	"io/ioutil"
	"sync"
)

var (
	svrToken       uint32
	svrTokenLocker = new(sync.RWMutex)
)

func LoadSvrToken() error {
	fileName := fmt.Sprintf("%s_%s_%d.token", Env, Type, ID)
	exist, err := util.IsFileExist(fileName)
	if err != nil {
		return fmt.Errorf("读取服务器启动token文件失败：%v", err)
	}
	if exist {
		fileBytes, err := ioutil.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("读取服务器启动token文件失败：%v", err)
		}
		token, err := util.StrToUint32(string(fileBytes))
		if err != nil {
			return fmt.Errorf("服务器启动token文件内容损坏：%v", err)
		}
		SetSvrToken(token)
		return nil
	} else {
		// token文件不存在，就默认值为0
		return nil
	}
}

func SetSvrToken(token uint32) {
	svrTokenLocker.Lock()
	defer svrTokenLocker.Unlock()
	svrToken = token
}

func GetSvrToken() uint32 {
	svrTokenLocker.RLock()
	defer svrTokenLocker.RUnlock()
	return svrToken
}

func IncSvrToken() {
	fileName := fmt.Sprintf("%s_%s_%d.token", Env, Type, ID)
	svrTokenLocker.Lock()
	defer svrTokenLocker.Unlock()
	svrToken++
	if err := ioutil.WriteFile(fileName, []byte(fmt.Sprintf("%d", svrToken)), 0644); err != nil {
		log.GetLogger().Error(fmt.Sprintf("更新服务器令牌失败！原因：%v", err))
	}
}
