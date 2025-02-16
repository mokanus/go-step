package rk

import (
	"fmt"
	"go-step/log"
	"go-step/log/ubiquitous/log/field"
	"sort"
	"time"
)

func (self *Rank) Sort() {
	self.Lock()
	defer self.Unlock()
	self.sort()
}

// 从itemMap中克隆数据出来排序，得到榜单
// 不对itemMap直接排序的原因是：item的value修改会导致value和排名不一致
func (self *Rank) sort() {
	if !self.needSortFlag {
		return
	}
	self.needSortFlag = false

	t1 := time.Now()

	n := len(self.itemMap)

	itemList := make([]*RankItem, 0, n)

	// 克隆
	for _, item := range self.itemMap {
		clone := new(RankItem)
		*clone = *item
		itemList = append(itemList, clone)
	}

	// 排序
	sort.Slice(itemList, func(i, j int) bool {
		if itemList[i].Value == itemList[j].Value {
			return itemList[i].UpdateTime < itemList[j].UpdateTime
		} else {
			return itemList[i].Value > itemList[j].Value
		}
	})

	// 在前3名有条件限制的情况下，给Rank字段赋值
	toPickNum := 3
	if toPickNum > n {
		toPickNum = n
	}
	pickedNum := 0
	for rank := int32(1); rank <= 3; rank++ {
		limit := self.top3Limit[rank-1]
		for i := pickedNum; i < toPickNum; i++ {
			if itemList[i].Value >= limit {
				itemList[i].Rank = rank
				pickedNum++
				break
			}
		}
	}
	var rank int32 = 4
	for i := pickedNum; i < len(itemList); i++ {
		itemList[i].Rank = rank
		rank++
	}

	// 删除Rank>MaxRank的那些item。对应的，这些item也从itemMap和数据库移除。
	removedNum := 0
	for i := n - 1; i >= 0; i-- {
		if itemList[i].Rank > self.MaxRank {
			if err := self.getDbAgent(self.Name).RemoveId(itemList[i].Key); err == nil {
				delete(self.itemMap, itemList[i].Key)
				removedNum++
				log.GetLogger().Info("排行榜删除", field.String("name", self.Name), field.Any("rank_items", itemList[i]))
			} else {
				log.GetLogger().Error("排行榜删除失败！", field.String("name", self.Name), field.Any("rank_items", itemList[i]), field.Error(err))
				break
			}
		} else {
			break
		}
	}
	if removedNum > 0 {
		itemList = itemList[:n-removedNum]
	}

	// 缓存最终榜单
	self.ItemSortedList = itemList

	// 建立key到排序索引的映射表
	self.ItemKeyToIndex = make(map[string]int)
	for i, item := range itemList {
		self.ItemKeyToIndex[item.Key] = i
	}

	log.GetLogger().Debug(fmt.Sprintf("排行榜[%s] 数量：[%d-%d=%d]，排序用时：%v", self.Name, n, removedNum, len(itemList), time.Now().Sub(t1)))
}
