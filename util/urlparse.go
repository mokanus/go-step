package util

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

var (
	ErrQueryValuesParam = errors.New("Unmarshal() expects struct or map[string]string pointer input. ")
)

func QueryValues(values url.Values, s interface{}) error {
	val := reflect.ValueOf(s)

	// 参数应该为指针
	if val.Kind() != reflect.Ptr {
		return ErrQueryValuesParam
	}

	if val.IsNil() {
		return ErrQueryValuesParam
	}

	val = val.Elem()

	// 传进来的是map，则用遍历填充的方式
	if val.Kind() == reflect.Map {
		pm, ok := s.(*map[string]string)
		if !ok {
			return ErrQueryValuesParam
		}

		for k, value := range values {
			if len(value) > 0 {
				(*pm)[k] = value[0]
			}
		}
		return nil
	}

	// 传进来的是结构体，则遍历结构体字段去values里取值
	if val.Kind() == reflect.Struct {
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			kt := typ.Field(i)
			tag := kt.Tag.Get("json")
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = kt.Name
			}
			sv := val.Field(i)
			uv := values.Get(tag)
			if uv == "" {
				continue
			}
			switch sv.Kind() {
			case reflect.String:
				sv.SetString(uv)
			case reflect.Bool:
				b, err := strconv.ParseBool(uv)
				if err != nil {
					return errors.New(fmt.Sprintf("cast bool has error, expect type: %v, val: %v, query key: %v", sv.Type(), uv, tag))
				}
				sv.SetBool(b)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				n, err := strconv.ParseUint(uv, 10, 64)
				if err != nil || sv.OverflowUint(n) {
					return errors.New(fmt.Sprintf("cast uint has error, expect type: %v, val: %v, query key: %v", sv.Type(), uv, tag))
				}
				sv.SetUint(n)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				n, err := strconv.ParseInt(uv, 10, 64)
				if err != nil || sv.OverflowInt(n) {
					return errors.New(fmt.Sprintf("cast int has error, expect type: %v, val: %v, query key: %v", sv.Type(), uv, tag))
				}
				sv.SetInt(n)
			case reflect.Float32, reflect.Float64:
				n, err := strconv.ParseFloat(uv, sv.Type().Bits())
				if err != nil || sv.OverflowFloat(n) {
					return errors.New(fmt.Sprintf("cast float has error, expect type: %v, val: %v, query key: %v", sv.Type(), uv, tag))
				}
				sv.SetFloat(n)
			default:
				return errors.New(fmt.Sprintf("unsupported type: %v, val: %v, query key: %v", sv.Type(), uv, tag))
			}
		}
		return nil
	}

	return ErrQueryValuesParam
}
