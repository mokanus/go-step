package util

import (
	"encoding/json"
	"fmt"
	"go-step/log"
	"go-step/log/ubiquitous/log/field"
	"sort"
	"strings"
)

func MakeUrlAndSign(params map[string]string, signKey string, ignoreField string) (string, string) {
	// 按键ASCII码值asc排序
	var keys = make([]string, 0, len(params))
	for key := range params {
		if key == ignoreField {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 重组键值对，url参数方式连接
	var sortedUrl string
	for idx, key := range keys {
		if idx == 0 {
			sortedUrl += fmt.Sprintf("%s=%s", key, params[key])
		} else {
			sortedUrl += fmt.Sprintf("&%s=%s", key, params[key])
		}
	}

	// 对拼好了的url参数串进行rawurlencode，然后拼上signKey进行MD5加密
	sign := MD5([]byte(fmt.Sprintf("%s&%s", RawUrlEncode(sortedUrl), signKey)))

	return sortedUrl, sign
}

func MakeUrlAndSign2(params map[string]interface{}, signKey string, ignoreField string) (string, string) {
	// 按键ASCII码值asc排序
	var keys = make([]string, 0, len(params))
	for key := range params {
		if key == ignoreField {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 重组键值对，url参数方式连接
	var kvPairs []string
	for _, key := range keys {
		v := params[key]
		strValue := ""
		if v != nil {
			switch v := v.(type) {
			case string: // 单独处理字符串类型，防止被双引号包裹
				strValue = v
			default: // 通过 json 序列化处理其他的类型值
				jsonBytes, _ := json.Marshal(v)
				strValue = string(jsonBytes)
			}
		}
		kvPairs = append(kvPairs, key+"="+strValue)
	}

	// 拼接参数
	sortedUrl := strings.Join(kvPairs, "&")
	log.GetLogger().Debug("拼接串", field.String("sortedUrl", sortedUrl))
	// 对拼好了的url参数串进行rawurlencode，然后拼上signKey进行MD5加密
	encodedSortedUrl := RawUrlEncode(sortedUrl)
	toMD5 := encodedSortedUrl + "&" + signKey
	sign := MD5([]byte(toMD5))

	return sortedUrl, sign
}
