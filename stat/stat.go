package stat

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-step/log"
	"go-step/log/ubiquitous/log/field"
	"os"
	"strings"
	"time"
)

var (
	open              bool
	path              string
	curLoggerFile     *os.File
	curLoggerFilePath string
)

func Config(statPath string) {
	open = false
	path = ""

	if statPath != "" {
		statPath = strings.TrimRight(statPath, "/")
		if err := ensureDir(statPath); err == nil {
			open = true
			path = statPath
		} else {
			log.GetLogger().Error("指定埋点上报输出根目录创建失败，关闭埋点上报输出！", field.String("path", statPath))
		}
	}

	tip := fmt.Sprintf("开启埋点上报：%t", open)
	if open {
		tip += fmt.Sprintf("，上报输出根目录：%s", path)
	}
	log.GetLogger().Info(tip)
}

func Write(content map[string]interface{}) {
	if !open {
		return
	}

	contentBytes, err := json.Marshal(content)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("埋点上报%v时，编码失败：%v", content, err))
		return
	}

	now := time.Now()
	if logger := getLogger(now); logger != nil {
		if _, err := logger.WriteString(fmt.Sprintf("%s\n", string(contentBytes))); err != nil {
			log.GetLogger().Error("埋点上报写入报错", field.Error(err))
		}
	}
}

// ----------------------------------------------------------------------------
// 内部函数
// ----------------------------------------------------------------------------
func getLogger(now time.Time) *os.File {
	y, m, d := now.Date()

	// 构造埋点文件路径
	filePath := fmt.Sprintf("%s/%d-%02d-%02d.log", path, y, m, d)

	// 如果跨天了，将已有logger关闭，并确保当天的日志路径创建好
	if curLoggerFile == nil || curLoggerFilePath != filePath {
		if curLoggerFile != nil {
			curLoggerFile.Close()
			curLoggerFile = nil
		}

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil || file == nil {
			log.GetLogger().Error("创建埋点上报文件%s失败", field.String("path", filePath), field.Error(err))
			return nil
		}

		curLoggerFile = file
		curLoggerFilePath = filePath
	}

	return curLoggerFile
}

func ensureDir(dir string) error {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// 目录不存在，尝试创建
			return os.MkdirAll(dir, 0777)
		} else {
			// 判断函数执行失败
			return err
		}
	} else {
		if !fileInfo.IsDir() {
			// 指定的路径不是目录，而是一个文件，也返回错误
			return errors.New("指定路径已有同名文件存在，无法创建目录")
		} else {
			// 指定的目录已存在，返回成功
			return nil
		}
	}
}
