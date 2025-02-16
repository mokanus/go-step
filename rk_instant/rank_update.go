package rk_instant

// 对指定item进行冒泡
func (self *Rank) bubble(item *RankItem) {
	for {
		// 玩家已经在第1名了
		if item.RankIndex <= 0 {
			break
		}

		preItem := self.ItemSortedList[item.RankIndex-1]

		// 玩家分数没超过前一名，冒泡结束
		if item.Value <= preItem.Value {
			break
		}

		// 玩家分数超过前一名了，则与前一名位置互换
		preItem.RankIndex++
		self.ItemSortedList[item.RankIndex] = preItem
		item.RankIndex--
		self.ItemSortedList[item.RankIndex] = item
	}
}

// 对指定item进行下沉
func (self *Rank) sink(item *RankItem) {
	numItem := len(self.ItemSortedList)
	for {
		// 玩家已经是在最后一名了
		if item.RankIndex >= numItem-1 {
			break
		}

		nxtItem := self.ItemSortedList[item.RankIndex+1]

		// 玩家分数没低于后一名，下沉结束
		if item.Value > nxtItem.Value {
			break
		}

		// 玩家分数低于后一名了，则与后一名位置互换
		nxtItem.RankIndex--
		self.ItemSortedList[item.RankIndex] = nxtItem
		item.RankIndex++
		self.ItemSortedList[item.RankIndex] = item
	}
}
