package csv

import (
	"bytes"
	ec "encoding/csv"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go-step/log"
	"go-step/log/ubiquitous/log/field"
	"go-step/util"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrCsv    = errors.New("Wrong csv file. ")
	ErrResult = errors.New("Param result expect ptr of struct ptr slice. ")
)

// 注意！！
// 在项目中使用该函数时，必须晚于proto的init的执行。所以，该函数基于就只限定于tpl的Load函数中调用。tpl的load函数，是可以确保易于proto的init的执行的。
func Unmarshal(content []byte, result interface{}) error {
	reader := ec.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}
	if len(lines) < 4 {
		return ErrCsv
	}
	if len(lines[0]) < 2 {
		return ErrCsv
	}
	if len(lines) == 4 {
		return nil
	}

	// 此时resultT为Slice指针的Type
	resultT := reflect.TypeOf(result)
	if resultT.Kind() != reflect.Ptr {
		return ErrResult
	}

	// 此时resultT为Slice的Type
	resultT = resultT.Elem()
	if resultT.Kind() != reflect.Slice {
		return ErrResult
	}

	// 此时resultT为Slice中结构体的指针的Type
	resultT = resultT.Elem()
	if resultT.Kind() != reflect.Ptr {
		return ErrResult
	}

	// 此时resultT为Slice中的结构体的Type
	resultT = resultT.Elem()
	if resultT.Kind() != reflect.Struct {
		return ErrResult
	}

	tableName := resultT.Name()
	tableName = strings.TrimSuffix(tableName, "Data")

	// 特殊处理：对于Parameter表，原本每一行是一个kv，转置为所有列为k、只有一个数据行为v
	if strings.Contains(tableName, "Parameter") {
		lines = transLines(lines)
	}

	// 遍历该结构体，检查是否有枚举字段，如果有，准备好相关的枚举值映射表
	fieldEnumMap := makeFiledEnumMap(resultT)

	// 此时resultV为Slice的指针的Value
	resultV := reflect.ValueOf(result)
	// 此时resultV为Slice的Value（可用于Append和Set了）
	resultV = resultV.Elem()

	// 列名转索引，索引可用于从line取值
	colMap := make(map[string]int)
	for j, col := range lines[0][1:] {
		colMap[col] = j
	}

	for _, line := range lines[4:] {
		if len(line) < len(lines[0]) {
			return ErrCsv
		}
		if strings.HasPrefix(strings.TrimSpace(line[0]), "#") {
			continue
		}
		recordV := reflect.New(resultT)
		if err := unmarshal(tableName, line[1:], colMap, fieldEnumMap, resultT, recordV.Elem()); err != nil { // 注意：要拿Elem进去填值
			return err
		}
		resultV.Set(reflect.Append(resultV, recordV))
	}

	return nil
}

// resultT是结构体的Type，recordV是结构体的Value
func unmarshal(tableName string, line []string, colMap map[string]int, fieldEnumMap map[int]map[string]int64, recordT reflect.Type, recordV reflect.Value) error {
	// 遍历目标结构体的每个字段，根据字段名，从values中取出字符串类型的值，再根据字段类型，将字符串类型的值解析出来
	for i := 0; i < recordV.NumField(); i++ {
		kt := recordT.Field(i)
		tag := kt.Tag.Get("csv")
		if tag == "-" {
			continue
		}
		index, ok := colMap[kt.Name]
		if !ok {
			log.GetLogger().Error("配置表的结构体字段在表中没有对应的列名！请检查字段命名！", field.String("table_name", tableName), field.String("name", kt.Name))
			continue
		}
		uv := line[index]
		uv = strings.TrimSpace(uv)
		if uv == "" {
			continue
		}
		sv := recordV.Field(i)

		if enumMap := fieldEnumMap[i]; enumMap != nil {
			n, ok := enumMap[uv]
			if !ok {
				// return errors.New(fmt.Sprintf("unexpect enum: %v", uv))
				log.GetLogger().Error(fmt.Sprintf("配置表[%s]中列[%s]配置的枚举[%s]无法识别！请检查：1、枚举配置是否正确；2、服务端是否已生成最新协议！", tableName, kt.Name, uv))
			}
			sv.SetInt(n)
		} else if tag != "" {
			switch tag {
			case "[]int32":
				sv.Set(reflect.ValueOf(util.StrToInt32List(uv)))
				break
			case "[][2]int32":
				sv.Set(reflect.ValueOf(util.StrToTwoInt32List(uv)))
				break
			case "[][3]int32":
				sv.Set(reflect.ValueOf(util.StrToThreeInt32List(uv)))
				break
			case "[][4]int32":
				sv.Set(reflect.ValueOf(util.StrToFourInt32List(uv)))
				break
			case "[]int64":
				sv.Set(reflect.ValueOf(util.StrToInt64List(uv)))
				break
			case "[]float64":
				sv.Set(reflect.ValueOf(util.StrToFloat64List(uv)))
				break
			case "[][2]float64":
				sv.Set(reflect.ValueOf(util.StrToTwoFloat64List(uv)))
				break
			case "[][3]float64":
				sv.Set(reflect.ValueOf(util.StrToThreeFloat64List(uv)))
				break
			case "[][2]int64":
				sv.Set(reflect.ValueOf(util.StrToTwoInt64List(uv)))
				break
			case "[][3]int64":
				sv.Set(reflect.ValueOf(util.StrToThreeInt64List(uv)))
				break
			case "[][4]int64":
				sv.Set(reflect.ValueOf(util.StrToFourInt64List(uv)))
			case "[][5]int64":
				sv.Set(reflect.ValueOf(util.StrToFiveInt64List(uv)))
				break
			case "int64":
				sv.Set(reflect.ValueOf(util.ScienceToInt64(uv)))
				break
			case "time":
				unix, err := util.TimeStrToUnix(uv)
				if err != nil {
					return errors.New(fmt.Sprintf("cast time has error, reason: %v, val: %v, col: %v", err, uv, kt.Name))
				}
				sv.Set(reflect.ValueOf(unix))
			default:
				return errors.New(fmt.Sprintf("unsupported csv tag: %s", tag))
			}
		} else {
			switch sv.Kind() {
			case reflect.String:
				sv.SetString(uv)
			case reflect.Bool:
				b, err := strconv.ParseBool(uv)
				if err != nil {
					return errors.New(fmt.Sprintf("cast bool has error, expect type: %v, val: %v, col: %v", sv.Type(), uv, kt.Name))
				}
				sv.SetBool(b)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				n, err := strconv.ParseUint(uv, 10, 64)
				if err != nil || sv.OverflowUint(n) {
					return errors.New(fmt.Sprintf("cast uint has error, expect type: %v, val: %v, col: %v", sv.Type(), uv, kt.Name))
				}
				sv.SetUint(n)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				n, err := strconv.ParseInt(uv, 10, 64)
				if err != nil || sv.OverflowInt(n) {
					return errors.New(fmt.Sprintf("cast int has error, expect type: %v, val: %v, col: %v", sv.Type(), uv, kt.Name))
				}
				sv.SetInt(n)
			case reflect.Float32, reflect.Float64:
				n, err := strconv.ParseFloat(uv, sv.Type().Bits())
				if err != nil || sv.OverflowFloat(n) {
					return errors.New(fmt.Sprintf("cast float has error, expect type: %v, val: %v, col: %v", sv.Type(), uv, kt.Name))
				}
				sv.SetFloat(n)
			default:
				return errors.New(fmt.Sprintf("unsupported type: %v, val: %v, col: %v", sv.Type(), uv, kt.Name))
			}
		}
	}
	return nil
}

func transLines(lines [][]string) [][]string {
	// 将所有k放在newLine0，作为列名；将所有v放在newLine4，作为数据。第一个元素放空。
	newLine0 := []string{""}
	newLine4 := []string{""}

	newLine0 = append(newLine0)
	for _, line := range lines[4:] {
		newLine0 = append(newLine0, line[1]) // line[1]是k
		newLine4 = append(newLine4, line[2]) // line[2]是v
	}

	return [][]string{
		newLine0, // 列名行
		nil,      // 注释行
		nil,      // 类型行
		nil,      // 备注行
		newLine4, // 数据行，转置后的参数表，只会有一个数据行
	}
}

func makeFiledEnumMap(recordT reflect.Type) map[int]map[string]int64 {
	fieldEnumMap := make(map[int]map[string]int64)
	for i := 0; i < recordT.NumField(); i++ {
		tp := recordT.Field(i).Type.String()
		if strings.HasPrefix(tp, "pb.") {
			enum := strings.TrimSuffix(strings.TrimPrefix(tp, "pb."), "Type")
			enumMap := make(map[string]int64)
			for k, v := range proto.EnumValueMap("pbproto." + strings.TrimPrefix(tp, "pb.")) {
				enumMap[strings.TrimPrefix(k, enum)] = int64(v)
			}
			fieldEnumMap[i] = enumMap
		}
	}
	return fieldEnumMap
}
