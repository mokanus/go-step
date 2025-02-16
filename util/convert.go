package util

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ----------------------------------------------------------------------------
// 字符串转int32相关
// ----------------------------------------------------------------------------
func StrToInt32(str string) (int32, error) {
	v, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func StrMustToInt32(str string) int32 {
	v, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return int32(v)
}

func StrMustToInt64(str string) int64 {
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func StrToInt32List(str string) (list []int32) {
	if str == "" {
		return
	}
	if strings.Contains(str, ":") {
		return
	}
	sp := strings.Split(str, ";")
	for _, s := range sp {
		v, err := StrToInt32(s)
		if err == nil {
			list = append(list, v)
		}
	}
	return
}

func StrToTwoInt32List(str string) [][2]int32 {
	list := make([][2]int32, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt32List(frag)
		if len(item) != 2 {
			continue
		}
		list = append(list, [2]int32{item[0], item[1]})
	}

	return list
}

func StrToThreeInt32List(str string) [][3]int32 {
	list := make([][3]int32, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt32List(frag)
		if len(item) != 3 {
			continue
		}
		list = append(list, [3]int32{item[0], item[1], item[2]})
	}

	return list
}

func StrToFourInt32List(str string) [][4]int32 {
	list := make([][4]int32, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt32List(frag)
		if len(item) != 4 {
			continue
		}
		list = append(list, [4]int32{item[0], item[1], item[2], item[3]})
	}

	return list
}

// ----------------------------------------------------------------------------
// 字符串转int64相关
// ----------------------------------------------------------------------------
func StrToInt64(str string) (v int64, err error) {
	return strconv.ParseInt(str, 10, 64)
}

func StrToInt64List(str string) (list []int64) {
	if str == "" {
		return
	}
	sp := strings.Split(str, ";")
	for _, s := range sp {
		v, err := StrToInt64(s)
		if err == nil {
			list = append(list, v)
		}
	}
	return
}

func StrToTwoInt64List(str string) [][2]int64 {
	list := make([][2]int64, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt64List(frag)
		if len(item) != 2 {
			continue
		}
		list = append(list, [2]int64{item[0], item[1]})
	}

	return list
}

func StrToThreeInt64List(str string) [][3]int64 {
	list := make([][3]int64, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt64List(frag)
		if len(item) != 3 {
			continue
		}
		list = append(list, [3]int64{item[0], item[1], item[2]})
	}

	return list
}

func StrToFourInt64List(str string) [][4]int64 {
	list := make([][4]int64, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt64List(frag)
		if len(item) != 4 {
			continue
		}
		list = append(list, [4]int64{item[0], item[1], item[2], item[3]})
	}

	return list
}

func StrToFiveInt64List(str string) [][5]int64 {
	list := make([][5]int64, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt64List(frag)
		list = append(list, [5]int64{item[0], item[1], item[2], item[3], item[4]})
	}

	return list
}

func ScienceToInt64(str string) int64 {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(fmt.Sprintf("配置错误 %s", str))
	}
	return int64(v)
}

func StrToSomeInt64List(str string) [][]int64 {
	list := make([][]int64, 0)
	frags := strings.Split(str, "|")
	for _, frag := range frags {
		list = append(list, StrToInt64List(frag))
	}
	return list
}

// ----------------------------------------------------------------------------
// 字符串转uint32相关
// ----------------------------------------------------------------------------
func StrToUint32(str string) (uint32, error) {
	v, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// ----------------------------------------------------------------------------
// 字符串转uint64相关
// ----------------------------------------------------------------------------
func StrToUint64(str string) (v uint64, err error) {
	return strconv.ParseUint(str, 10, 64)
}

func StrToUint64List(str string) (list []uint64) {
	sp := strings.Split(str, ";")
	for _, s := range sp {
		v, err := StrToUint64(s)
		if err == nil {
			list = append(list, v)
		}
	}
	return
}

func StrToTwoUint64List(str string) [][2]uint64 {
	list := make([][2]uint64, 0)

	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToUint64List(frag)
		if len(item) != 2 {
			continue
		}
		list = append(list, [2]uint64{item[0], item[1]})
	}

	return list
}

// ----------------------------------------------------------------------------
// 字符串转float64相关
// ----------------------------------------------------------------------------
func StrToFloat64List(str string) (list []float64) {
	if str == "" {
		return
	}
	sp := strings.Split(str, ";")
	for _, s := range sp {
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			list = append(list, v)
		}
	}
	return
}

func StrToTwoFloat64List(str string) [][2]float64 {
	list := make([][2]float64, 0)
	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToFloat64List(frag)
		if len(item) != 2 {
			continue
		}
		list = append(list, [2]float64{item[0], item[1]})
	}
	return list
}

func StrToThreeFloat64List(str string) [][3]float64 {
	list := make([][3]float64, 0)
	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToFloat64List(frag)
		if len(item) != 3 {
			continue
		}
		list = append(list, [3]float64{item[0], item[1], item[2]})
	}
	return list
}

func StrToSomeFloat64List(str string) [][]float64 {
	list := make([][]float64, 0)
	frags := strings.Split(str, "|")
	for _, frag := range frags {
		list = append(list, StrToFloat64List(frag))
	}
	return list
}

// ----------------------------------------------------------------------------
// 整型转字符串相关
// ----------------------------------------------------------------------------
func Int32ToStr(v int32) string {
	return fmt.Sprintf("%d", v)
}

func Int32ListToStr(list []int32) string {
	str := ""
	for _, v := range list {
		if str != "" {
			str += ";"
		}
		str += fmt.Sprintf("%d", v)
	}
	return str
}

func TwoInt32ListToStr(list [][2]int32) string {
	str := ""
	for _, item := range list {
		if str != "" {
			str += "|"
		}
		str += fmt.Sprintf("%d;%d", item[0], item[1])
	}
	return str
}

func ThreeInt32ListToStr(list [][3]int32) string {
	str := ""
	for _, item := range list {
		if str != "" {
			str += "|"
		}
		str += fmt.Sprintf("%d;%d;%d", item[0], item[1], item[2])
	}
	return str
}

// ----------------------------------------------------------------------------
// 字符串与整型map互转相关
// ----------------------------------------------------------------------------
func StrToInt32Map(str string) map[int32]int32 {
	dict := make(map[int32]int32)
	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt32List(frag)
		if len(item) != 2 {
			continue
		}
		dict[item[0]] += item[1]
	}
	return dict
}

func Int32MapToStr(dict map[int32]int32) string {
	str := ""
	for k, v := range dict {
		if str != "" {
			str += "|"
		}
		str += fmt.Sprintf("%d;%d", k, v)
	}
	return str
}

func StrToInt64Map(str string) map[int64]int64 {
	dict := make(map[int64]int64)
	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := StrToInt64List(frag)
		if len(item) != 2 {
			continue
		}
		dict[item[0]] += item[1]
	}
	return dict
}

func Int64MapToStr(dict map[int64]int64) string {
	str := ""
	for k, v := range dict {
		if str != "" {
			str += "|"
		}
		str += fmt.Sprintf("%d;%d", k, v)
	}
	return str
}

func StrToStrUin32Map(str string) map[string]uint32 {
	dict := make(map[string]uint32)
	frags := strings.Split(str, "|")
	for _, frag := range frags {
		item := strings.Split(frag, ";")
		if len(item) != 2 {
			continue
		}
		num, err := StrToUint32(item[1])
		if err != nil {
			continue
		}
		dict[item[0]] = num
	}
	return dict
}

func StrUin32MapToStr(dict map[string]uint32) string {
	str := ""
	for k, v := range dict {
		if str != "" {
			str += "|"
		}
		str += fmt.Sprintf("%v;%d", k, v)
	}
	return str
}

// 驼峰转蛇形小写  RankArenaScore -> rank_arena_score
func ToSnakeCase(str string) string {
	var final []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				final = append(final, '_')
			}
			final = append(final, unicode.ToLower(r))
		} else {
			final = append(final, r)
		}
	}
	return string(final)
}

func MergeInt32Map(a, b map[int32]int64) (c map[int32]int64) {
	c = make(map[int32]int64)
	for k, v := range a {
		c[k] += v
	}
	for k, v := range b {
		c[k] += v
	}
	return c
}
