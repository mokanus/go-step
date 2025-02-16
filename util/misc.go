package util

import (
	"errors"
	"os"
)

// uint32类型做减法，如果结果为负数，将得到一个大的正数，从而导致程序上隐晦的BUG
func Uint32Sub(a, b uint32) uint32 {
	if a > b {
		return a - b
	} else {
		return 0
	}
}

// uint64类型做减法，如果结果为负数，将得到一个大的正数，从而导致程序上隐晦的BUG
func Uint64Sub(a, b uint64) uint64 {
	if a > b {
		return a - b
	} else {
		return 0
	}
}

// int取小
func IntMin(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

// int取大
func IntMax(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

// int32取差值
func Int32Diff(a, b int32) int32 {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

// int32取小
func Int32Min(a, b int32) int32 {
	if a <= b {
		return a
	} else {
		return b
	}
}

// int32取大
func Int32Max(a, b int32) int32 {
	if a >= b {
		return a
	} else {
		return b
	}
}

func CheckInList(value int32, list []int32) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

// int32列表取最小
func Int32ListMinIndex(list []int32) int {
	if len(list) == 0 {
		return -1
	}
	var minV int32
	var minI int
	for i, v := range list {
		if i == 0 {
			minV = v
			minI = i
		} else {
			if v < minV {
				minV = v
				minI = i
			}
		}
	}
	return minI
}

func BoolToUint8(v bool) uint8 {
	if v {
		return 1
	} else {
		return 0
	}
}

func EnsureDir(dir string) error {
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

func IsFileExist(filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		if fileInfo.IsDir() {
			return false, errors.New("同名路径为目录")
		} else {
			return true, nil
		}
	}
}

func IsDirExist(dirPath string) (bool, error) {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		if !fileInfo.IsDir() {
			return false, errors.New("同名路径为文件")
		} else {
			return true, nil
		}
	}
}
