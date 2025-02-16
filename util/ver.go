package util

import (
	"strconv"
	"strings"
)

type Ver struct {
	Str string
	Int []int
}

func ParseVer(ver string) (*Ver, bool) {
	sp := strings.Split(ver, ".")
	if len(sp) != 3 {
		return nil, false
	}

	list := make([]int, 0)
	for _, s := range sp {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, false
		}
		list = append(list, v)
	}

	return &Ver{Str: ver, Int: list}, true
}

// 大于
func (self *Ver) GT(other *Ver) bool {
	if self.Int[0] != other.Int[0] {
		return self.Int[0] > other.Int[0]
	}
	if self.Int[1] != other.Int[1] {
		return self.Int[1] > other.Int[1]
	}
	return self.Int[2] > other.Int[2]
}

// 大于等于
func (self *Ver) GTE(other *Ver) bool {
	if self.Int[0] != other.Int[0] {
		return self.Int[0] > other.Int[0]
	}
	if self.Int[1] != other.Int[1] {
		return self.Int[1] > other.Int[1]
	}
	return self.Int[2] >= other.Int[2]
}

// 小于
func (self *Ver) LT(other *Ver) bool {
	if self.Int[0] != other.Int[0] {
		return self.Int[0] < other.Int[0]
	}
	if self.Int[1] != other.Int[1] {
		return self.Int[1] < other.Int[1]
	}
	return self.Int[2] < other.Int[2]
}

// 小于等于
func (self *Ver) LTE(other *Ver) bool {
	if self.Int[0] != other.Int[0] {
		return self.Int[0] < other.Int[0]
	}
	if self.Int[1] != other.Int[1] {
		return self.Int[1] < other.Int[1]
	}
	return self.Int[2] <= other.Int[2]
}
