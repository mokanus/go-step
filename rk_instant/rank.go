package rk_instant

import (
	"fmt"
	"github.com/mokanus/go-step/app"
	"github.com/mokanus/go-step/g"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/util"
	"sync"
	"time"
)

type Rank struct {
	*sync.RWMutex

	Name           string // 排行榜名，也是排行榜在数据库中表名
	Capacity       int
	ItemMap        map[string]*RankItem
	ItemSortedList []*RankItem

	// 数据库代理获取函数
	getDbAgent func(string) *app.DbAgent
	blackLv    func(string) int32

	// Item的数据库镜像
	imageMap map[string]*RankItemImage
}

func NewRank(name string, capacity int, getDbAgent func(string) *app.DbAgent, blackLv func(string) int32) *Rank {
	return &Rank{
		RWMutex:    new(sync.RWMutex),
		Name:       name,
		Capacity:   capacity,
		getDbAgent: getDbAgent,
		blackLv:    blackLv,
	}
}

// 系统调用：排行榜定时入库
func (self *Rank) Loop(delay, interval int) {
	if interval == 0 {
		log.GetLogger().Error(fmt.Sprintf("Rank Loop的interval参数不能为0！！"))
		return
	}
	g.Go(func() {
		if delay > 0 {
			time.Sleep(time.Second * time.Duration(delay))
		}
		for {
			self.Save()
			time.Sleep(time.Second * time.Duration(interval))
		}
	})
}

// 系统调用：加载排行榜
func (self *Rank) Load() error {
	rawList := make([]*RankItem, 0, self.Capacity)
	if err := self.getDbAgent(self.Name).FindSortAll(&rawList, "-value", "updatetime"); err != nil {
		return err
	}

	imageMap := make(map[string]*RankItemImage, self.Capacity)
	itemMap := make(map[string]*RankItem, self.Capacity)
	itemSortedList := make([]*RankItem, 0, self.Capacity)

	existMap := make(map[string]*RankItem)
	rankIndex := 0
	for _, item := range rawList {
		// 数据库中读取出来的，要原原本本放到镜像Map进去
		itemImage := NewRankItemImage(item.Key)
		itemImage.CopyFrom(item)
		itemImage.Flag = ImageFlagNoChange
		imageMap[itemImage.Item.Key] = itemImage

		// 而真正的榜单，要过滤掉黑名单、过滤掉历史数据的重复项
		if self.blackLv(item.PlayerUid) > 0 {
			continue
		}

		// 旧数据处理：旧的种族排行榜一个玩家有多条记录，只保留一条，所以用PlayerUid覆盖Key
		// 因为已排序，要去重，就只保留第一个就行
		// 因为别的榜Key就是PlayerUid，所以不受影响
		item.Key = item.PlayerUid
		if existMap[item.Key] != nil {
			continue
		}
		existMap[item.Key] = item

		// 真正榜单收录真正有效的数据。被过滤掉的条目，因为在镜像中有，所以在下一次loop中，会被从数据库移除
		item.RankIndex = rankIndex
		itemMap[item.Key] = item
		itemSortedList = append(itemSortedList, item)

		rankIndex++
	}

	self.Lock()
	defer self.Unlock()

	self.ItemMap = itemMap
	self.ItemSortedList = itemSortedList
	self.imageMap = imageMap

	return nil
}

// 系统调用：保存排行榜
func (self *Rank) Save() {
	self.updateImageMap()
	self.syncImageMapToDB()
}

// 返回指定排序索引的item，不在榜上则为nil
func (self *Rank) Item(rankIndex int) *RankItem {
	self.Lock()
	defer self.Unlock()
	if rankIndex < 0 || rankIndex >= len(self.ItemSortedList) {
		return nil
	}
	return self.ItemSortedList[rankIndex]
}

// 返回指定key的排名（从1开始的正式排名）。未上榜返回0。
func (self *Rank) Rank(key string) int32 {
	self.Lock()
	defer self.Unlock()
	item := self.ItemMap[key]
	if item == nil {
		return 0
	}
	return int32(item.RankIndex) + 1
}

// 玩家value有变更新榜单
func (self *Rank) Update(key, playerUid string, regionID int32, playerName, decoration string, value int64, param1 int32, param2 int64, param3 int32, extraData []byte) {
	self.Lock()
	defer self.Unlock()

	item := self.ItemMap[key]
	if item == nil {
		//// 1. 玩家当前不在榜上
		numItem := len(self.ItemSortedList)
		if numItem < self.Capacity {
			// 1.1 玩家当前不在榜上，且榜未满：附到榜单末尾，然后冒泡
			item = &RankItem{
				Key:        key,
				PlayerUid:  playerUid,
				RegionID:   regionID,
				PlayerName: playerName,
				Decoration: decoration,
				Value:      value,
				Param1:     param1,
				Param2:     param2,
				Param3:     param3,
				ExtraData:  extraData,
				UpdateTime: util.NowInt64(),
				RankIndex:  numItem,
			}
			self.ItemMap[item.Key] = item
			self.ItemSortedList = append(self.ItemSortedList, item)
			self.bubble(item)
		} else {
			// 1.2 玩家当前不在榜上，且榜已满
			tailItem := self.ItemSortedList[numItem-1]
			if value > tailItem.Value {
				// 1.2.1 玩家当前不在榜上，且榜已满，但玩家超过最后一名了：玩家取代最后一名，然后冒泡
				delete(self.ItemMap, tailItem.Key)
				item := &RankItem{
					Key:        key,
					PlayerUid:  playerUid,
					RegionID:   regionID,
					PlayerName: playerName,
					Decoration: decoration,
					Value:      value,
					Param1:     param1,
					Param2:     param2,
					Param3:     param3,
					ExtraData:  extraData,
					UpdateTime: util.NowInt64(),
					RankIndex:  numItem - 1,
				}
				self.ItemMap[item.Key] = item
				self.ItemSortedList[item.RankIndex] = item
				self.bubble(item)
			} else {
				// 1.2.2 玩家当前不在榜上，且榜已满，且玩家未超过最后一名：不做处理
			}
		}
	} else {
		//// 2. 玩家当前已在榜上：更新item.Value，然后冒泡
		item.PlayerUid = playerUid
		item.RegionID = regionID
		item.PlayerName = playerName
		item.Decoration = decoration
		if item.Value != value {
			item.UpdateTime = util.NowInt64()
		}
		item.Value = value
		item.Param1 = param1
		item.Param2 = param2
		item.Param3 = param3
		item.ExtraData = extraData
		self.bubble(item)
		self.sink(item)
	}
}

// 根据key，将item从榜单移除：常用于一个item原本上榜了，然后value掉到上榜要求值以下了，就从榜单移除
func (self *Rank) KickByKey(key string) {
	self.Lock()
	defer self.Unlock()

	if self.ItemMap[key] == nil {
		return
	}

	index := -1
	for i, item := range self.ItemSortedList {
		if item.Key == key {
			delete(self.ItemMap, item.Key)
		} else {
			index++
			if i != index {
				item.RankIndex = index
				self.ItemSortedList[index] = item
				log.GetLogger().Debug(fmt.Sprintf("[排行榜] 位置%d上的key:%s不为目标key:%s，前移到位置%d上", i, item.Key, key, index))
			}
		}
	}
	log.GetLogger().Debug(fmt.Sprintf("[排行榜] kickByKey前，末位索引为%d，kickByKey后，末位索引为%d", len(self.ItemSortedList)-1, index))

	self.ItemSortedList = self.ItemSortedList[:index+1]
}

// 根据playerUid，将item从榜单移除：常用将一个playerUid拉黑
func (self *Rank) KickByPlayerUid(playerUid string) {
	self.Lock()
	defer self.Unlock()

	index := -1
	for i, item := range self.ItemSortedList {
		if item.PlayerUid == playerUid {
			delete(self.ItemMap, item.Key)
		} else {
			index++
			if i != index {
				item.RankIndex = index
				self.ItemSortedList[index] = item
				log.GetLogger().Debug(fmt.Sprintf("[排行榜] 位置%d上的playerUid:%s不为目标playerUid:%s，前移到位置%d上", i, item.PlayerUid, playerUid, index))
			}
		}
	}
	log.GetLogger().Debug(fmt.Sprintf("[排行榜] kickByPlayerUid前，末位索引为%d，kickByPlayerUid后，末位索引为%d", len(self.ItemSortedList)-1, index))

	self.ItemSortedList = self.ItemSortedList[:index+1]
}

// 玩家名字和装饰物更新到榜单
func (self *Rank) UpdatePlayerInfoByKey(playerUid, playerName, decoration string) {
	self.Lock()
	defer self.Unlock()
	item := self.ItemMap[playerUid]
	if item == nil {
		return
	}
	item.PlayerName = playerName
	item.Decoration = decoration
}

// 玩家名字和装饰物更新到榜单
func (self *Rank) UpdatePlayerInfoByPlayerUid(playerUid, playerName, decoration string) {
	self.Lock()
	defer self.Unlock()
	for _, item := range self.ItemSortedList {
		if item.PlayerUid == playerUid {
			item.PlayerName = playerName
			item.Decoration = decoration
		}
	}
}

// 清空榜单。如果清空完想立刻保存，就再调用一下Save。
func (self *Rank) Clear() {
	self.ItemSortedList = make([]*RankItem, 0, self.Capacity)
	self.ItemMap = make(map[string]*RankItem, self.Capacity)
}
