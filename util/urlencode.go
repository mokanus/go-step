package util

import (
	"net/url"
	"strings"
)

// URL编码：除-_.之外的所有非字母数字字符都将被替换成%后跟两位十六进制数；空格编码为+
// 等同于php的urlencode
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}

// URL还原：将被UrlEncode后的编码串还原为原始URL
func UrlDecode(str string) (string, error) {
	return url.QueryUnescape(str)
}

// 新URL编码：UrlEncode之后，将空格编码而得的+转为%20
func RawUrlEncode(str string) string {
	return strings.Replace(UrlEncode(str), "+", "%20", -1)
}

// 新URL还原：先将%20转为+后，交由UrlDecode进行编码串的还原
func RawUrlDecode(str string) (string, error) {
	return UrlDecode(strings.Replace(str, "%20", "+", -1))
}
