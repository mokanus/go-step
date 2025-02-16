package util

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
)

func MD5(source []byte) string {
	var md5gen = md5.New()
	md5gen.Write(source)
	return hex.EncodeToString(md5gen.Sum(nil))
}

func MD5Conf(confBytes []byte) string {
	md5gen := md5.New()
	md5gen.Write(confBytes)
	return hex.EncodeToString(md5gen.Sum(nil))[:5]
}

func MD5Data(dataMap map[string][]byte) string {
	if len(dataMap) == 0 {
		return "empty"
	}

	fileNameList := make([]string, 0, len(dataMap))
	for fileName := range dataMap {
		fileNameList = append(fileNameList, fileName)
	}
	sort.Slice(fileNameList, func(i, j int) bool {
		return fileNameList[i] < fileNameList[j]
	})

	// 按顺序计算md5
	md5gen := md5.New()
	for _, fileName := range fileNameList {
		md5gen.Write(dataMap[fileName])
	}

	return hex.EncodeToString(md5gen.Sum(nil))[:5]
}

func MD5StringList(strings ...string) string {
	result := ""
	for _, str := range strings {
		result += str
	}
	hash := md5.Sum([]byte(result))
	md5str := hex.EncodeToString(hash[:])
	return md5str
}
