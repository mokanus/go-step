package app

import (
	"errors"
	"fmt"
	"go-step/util"
	"io/ioutil"
	"sort"
	"strings"
)

var (
	flatDataLoaders []FlatDataLoader
	nestDataLoaders []NestDataLoader
)

type SubDataFile struct {
	Path string
	Name string
	Data []byte
}

type FlatDataLoader struct {
	name string
	load func([]byte) error
}

type NestDataLoader struct {
	name string
	load func([]*SubDataFile) error
}

func LoadData() (map[string][]byte, string) {
	dataMap := make(map[string][]byte)
	errList := make([]string, 0)
	errList = loadFlatData(dataMap, errList)
	errList = loadNestData(dataMap, errList)
	return dataMap, strings.Join(errList, "|")
}

func loadFlatData(dataMap map[string][]byte, errList []string) []string {
	for _, loader := range flatDataLoaders {
		bytes, err := ioutil.ReadFile(Conf.DataPath + "/" + loader.name)
		if err != nil {
			errList = append(errList, fmt.Sprintf("读取配置文件[%s]出错[%v]！", loader.name, err))
			continue
		}
		if err := loader.load(bytes); err != nil {
			errList = append(errList, fmt.Sprintf("加载配置文件[%s]出错[%v]！", loader.name, err))
			continue
		}
		// flat型的配置文件fileName就为main.go中写的csv文件名
		dataMap[loader.name] = bytes
	}
	return errList
}

func loadNestData(dataMap map[string][]byte, errList []string) []string {
	for _, loader := range nestDataLoaders {
		if !strings.Contains(loader.name, "*") {
			errList = append(errList, fmt.Sprintf("[%s]不是嵌套配置文件路径格式", loader.name))
			continue
		}

		subFileList, err := parseSubDataFiles(loader.name)
		if err != nil {
			errList = append(errList, fmt.Sprintf("从[%s]中罗列子文件列表出错：%v", loader.name, err))
			continue
		}

		subDataFileList := make([]*SubDataFile, 0)

		// 子文件路径格式举例：Map/100/MapWave.csv
		for _, subFile := range subFileList {
			subFilePath, subFileName, err := parseSubDataFilePath(subFile)
			if err != nil {
				errList = append(errList, fmt.Sprintf("解析配置文件[%s]的文件名出错[%v]！", subFile, err))
				continue
			}
			isExist, err := util.IsFileExist(Conf.DataPath + "/" + subFile)
			if err != nil {
				errList = append(errList, fmt.Sprintf("读取配置文件[%s]出错[%v]！", subFile, err))
				continue
			}
			if !isExist {
				continue
			}
			bytes, err := ioutil.ReadFile(Conf.DataPath + "/" + subFile)
			if err != nil {
				errList = append(errList, fmt.Sprintf("读取配置文件[%s]出错[%v]！", subFile, err))
				continue
			}
			dataMap[subFile] = bytes
			subDataFileList = append(subDataFileList, &SubDataFile{Path: subFilePath, Name: subFileName, Data: bytes})
		}

		if len(subDataFileList) == 0 {
			continue
		}

		if err := loader.load(subDataFileList); err != nil {
			errList = append(errList, fmt.Sprintf("加载嵌套文件[%s]下的%d个文件时，出错[%v]！", loader.name, len(subDataFileList), err))
			continue
		}
	}

	return errList
}

func parseSubDataFiles(subFilePath string) ([]string, error) {
	// 举例：["Map", "*", "MapWave.csv"]
	sps := strings.Split(subFilePath, "/")
	if len(sps) != 3 {
		return nil, errors.New("子文件路径格式错误")
	}

	subFileList := make([]string, 0)

	// 举例：先尝试遍历出Map_*_MapWave.csv文件来
	fileInfoList, err := ioutil.ReadDir(Conf.DataPath)
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range fileInfoList {
		if fileInfo.IsDir() {
			continue
		}
		fileName := fileInfo.Name()
		if !strings.HasSuffix(fileName, ".csv") {
			continue
		}
		if !strings.HasPrefix(fileName, sps[0]+"_") {
			continue
		}
		if !strings.HasSuffix(fileName, "_"+sps[2]) {
			continue
		}
		subFileList = append(subFileList, strings.Replace(fileName, "_", "/", -1))
	}

	if len(subFileList) > 0 {
		sort.Slice(subFileList, func(i, j int) bool {
			return subFileList[i] < subFileList[j]
		})
		// 将/转为_再返回
		newSubFileList := make([]string, 0)
		for _, subFile := range subFileList {
			newSubFileList = append(newSubFileList, strings.Replace(subFile, "/", "_", -1))
		}
		return newSubFileList, nil
	}

	// 没有Map_*_MapWave.csv文件，那就是开发环境了，那就罗列出Map/*/MapWave.csv
	dirInfoList, err := ioutil.ReadDir(Conf.DataPath + "/" + sps[0])
	if err != nil {
		return nil, err
	}
	for _, dirInfo := range dirInfoList {
		if !dirInfo.IsDir() {
			continue
		}
		subFileList = append(subFileList, fmt.Sprintf("%s/%s/%s", sps[0], dirInfo.Name(), sps[2]))
	}
	// 先排序
	sort.Slice(subFileList, func(i, j int) bool {
		return subFileList[i] < subFileList[j]
	})
	return subFileList, nil
}

func parseSubDataFilePath(subFile string) (string, string, error) {
	var sps []string
	if strings.Contains(subFile, "/") {
		sps = strings.Split(subFile, "/")
	} else {
		sps = strings.Split(subFile, "_")
	}
	if len(sps) != 3 {
		return "", "", errors.New("格式不对")
	}
	return sps[0] + "/" + sps[1] + "/", sps[2], nil
}
