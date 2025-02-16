package util

import (
	"math/rand"
)

type RandTool struct {
	rand *rand.Rand
}

func NewRandTool(seed int64) *RandTool {
	return &RandTool{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (self *RandTool) Random(min, max int) int {
	if min >= max {
		return min
	}
	r := self.rand.Intn(max - min + 1)
	return min + r
}

func (self *RandTool) RandomByWeight(weights []int) int {
	num := len(weights)

	if num == 0 {
		return -1
	}

	if num == 1 {
		return 0
	}

	var total int
	for _, w := range weights {
		if w < 0 {
			return -1
		}
		total += w
	}

	// 每个元素权重都为0，则相当于平均权重
	if total == 0 {
		return self.Random(0, num-1)
	}

	// total大于0，则按权重随机出一个元素
	r := self.Random(0, total-1)

	var a int
	for i, w := range weights {
		a += w
		if a > r {
			return i
		}
	}

	return -1
}

func (self *RandTool) RandItemByWeight(itemList [][3]int64) (int64, int64) {
	n := len(itemList)
	if n == 0 {
		return 0, 0
	}

	weightList := make([]int, 0, n)
	for _, item := range itemList {
		weightList = append(weightList, int(item[2]))
	}
	item := itemList[self.RandomByWeight(weightList)]

	return item[0], item[1]
}
