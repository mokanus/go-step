package util

import (
	"math/rand"
	"sync"
	"time"
)

var (
	random     = rand.New(rand.NewSource(time.Now().UnixNano()))
	randomLock = new(sync.Mutex)
)

// 在min和max间随机一个值(包含min,max) 注意：如果min大于等于max，则固定返回min
func Random(min, max int) int {
	if min >= max {
		return min
	}
	randomLock.Lock()
	r := random.Intn(max - min + 1)
	randomLock.Unlock()
	return min + r
}
func RandomBool() bool {
	return Random(1, 2) == 1
}

// 百分比概率
func RandomPercent(pct int) bool {
	return Random(1, 100) <= pct
}

func Rand() float64 {
	randomLock.Lock()
	f := random.Float64()
	randomLock.Unlock()
	return f
}

// 随机一个uint32
func RandUint32() uint32 {
	randomLock.Lock()
	r := random.Uint32()
	randomLock.Unlock()
	return r
}

// 根据权重数组随机,返回索引.返回-1代表随机失败
func RandomByWeight(weights []int) int {
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
		return Random(0, num-1)
	}

	// total大于0，则按权重随机出一个元素
	r := Random(0, total-1)

	var a int
	for i, w := range weights {
		a += w
		if a > r {
			return i
		}
	}

	return -1
}

// 从list中不重复地挑选num个
func RandPickInt32List(list []int32, num int) []int32 {
	// list不够挑，则list整个返回
	if len(list) <= num {
		return list
	}

	// 按num逐个不放回的挑选
	pickList := make([]int32, 0)
	for i := 0; i < num; i++ {
		idx := Random(0, len(list)-1)

		pickList = append(pickList, list[idx])

		remainList := make([]int32, 0)
		for i, x := range list {
			if i != idx {
				remainList = append(remainList, x)
			}
		}
		list = remainList
	}

	return pickList
}

// 从带权重list中按权重不重复地挑选num个
func RandPickInt32ListByWeight(list [][2]int32, num int) []int32 {
	vList := make([]int32, 0)
	wList := make([]int, 0)
	for _, x := range list {
		vList = append(vList, x[0])
		wList = append(wList, int(x[1]))
	}

	// vList不够挑，则vList整个返回
	if len(vList) <= num {
		return vList
	}

	// 按num逐个不放回的挑选
	pickList := make([]int32, 0)
	for i := 0; i < num; i++ {
		idx := RandomByWeight(wList)

		// 随机出错，提前返回
		if idx < 0 {
			return pickList
		}

		// 选中一个放入pickList
		pickList = append(pickList, vList[idx])

		remainVList := make([]int32, 0)
		remainWList := make([]int, 0)
		for i, v := range vList {
			if i != idx {
				remainVList = append(remainVList, v)
			}
		}
		for i, w := range wList {
			if i != idx {
				remainWList = append(remainWList, w)
			}
		}
		vList = remainVList
		wList = remainWList
	}

	return pickList
}

// Shuffle 随机打乱列表的方法
func Shuffle[T any](list []T) []T {
	rand.NewSource(time.Now().UnixNano()) // 设置随机种子
	shuffled := make([]T, len(list))
	copy(shuffled, list)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) // 生成一个 [0, i] 范围内的随机数
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}
