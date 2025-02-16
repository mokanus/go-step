package rk

import (
	"github.com/globalsign/mgo/bson"
	"github.com/mokanus/go-step/app"
	"github.com/mokanus/go-step/g"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/log/ubiquitous/log/field"
	"github.com/mokanus/go-step/util"
	"strings"
	"sync"
	"time"
)

type Rank struct {
	*sync.RWMutex

	Name           string // 排行榜名，也是排行榜在数据库中表名
	MaxRank        int32
	ItemSortedList []*RankItem
	ItemKeyToIndex map[string]int

	// 对应数据库缓存
	itemMap map[string]*RankItem

	// 数据库代理获取函数、黑名单获取函数
	getDbAgent func(string) *app.DbAgent
	blackLv    func(string) int32

	// 前3名的上榜条件
	top3Limit [3]int64

	// 是否要排序的标识
	needSortFlag bool
}

func NewRank(name string, maxRank int32, top3Limit []int64, getDbAgent func(string) *app.DbAgent, blackLv func(string) int32) *Rank {
	rank := &Rank{
		RWMutex:        new(sync.RWMutex),
		Name:           name,
		MaxRank:        maxRank,
		ItemSortedList: make([]*RankItem, 0),
		ItemKeyToIndex: make(map[string]int),
		itemMap:        make(map[string]*RankItem),
		getDbAgent:     getDbAgent,
		blackLv:        blackLv,
	}

	if len(top3Limit) == 3 {
		rank.top3Limit[0] = top3Limit[0]
		rank.top3Limit[1] = top3Limit[1]
		rank.top3Limit[2] = top3Limit[2]
	}

	return rank
}

func (self *Rank) Loop() {
	g.Go(func() {
		// 初始1-3000毫秒做一个delay，避免各排行榜同时执行
		time.Sleep(time.Millisecond * time.Duration(util.Random(1, 3000)))
		// 每3秒，将itemMap克隆成list进行排序，得到榜单
		for {
			self.Sort()
			time.Sleep(time.Second * 3)
		}
	})
}

// 系统调用：加载排行榜
func (self *Rank) Load() error {
	itemList := make([]*RankItem, 0, self.MaxRank)
	if err := self.getDbAgent(self.Name).FindAll(nil, &itemList); err != nil {
		return err
	}

	// 读取到内存的map中做镜像缓存
	itemMap := make(map[string]*RankItem, self.MaxRank)
	for _, item := range itemList {
		if self.blackLv(item.PlayerUid) > 0 {
			if err := self.getDbAgent(self.Name).RemoveId(item.Key); err != nil {
				log.GetLogger().Error("排行榜[%s]加载时删除%+v失败！%v", field.String("name", self.Name), field.Any("item", item), field.Error(err))
			}
		} else {
			itemMap[item.Key] = item
		}
	}

	self.Lock()
	defer self.Unlock()

	self.itemMap = itemMap
	self.needSortFlag = true

	self.sort()

	return nil
}

// 返回指定key的排名（从1开始的正式排名）。未上榜返回0。
func (self *Rank) Rank(key string) int32 {
	self.Lock()
	defer self.Unlock()
	index, ok := self.ItemKeyToIndex[key]
	if !ok {
		return 0
	}
	if index < 0 || index >= len(self.ItemSortedList) {
		return 0
	}
	return self.ItemSortedList[index].Rank
}

// 返回指定排序索引的item，不在榜上则为nil。注意：只是列表第1位的玩家，他排名不一定是第一名！！
func (self *Rank) ItemByIndex(index int) *RankItem {
	self.Lock()
	defer self.Unlock()
	if index < 0 || index >= len(self.ItemSortedList) {
		return nil
	}
	return self.ItemSortedList[index]
}

// 返回指定key的item
func (self *Rank) ItemByKey(key string) *RankItem {
	self.Lock()
	defer self.Unlock()
	index, ok := self.ItemKeyToIndex[key]
	if !ok {
		return nil
	}
	if index < 0 || index >= len(self.ItemSortedList) {
		return nil
	}
	return self.ItemSortedList[index]
}

// 玩家value有变更新榜单
func (self *Rank) Update(key, playerUid string, regionID int32, playerName, decoration string, value int64,
	param1 int32, param2 int64, param3 int32, extraData []byte) {
	self.Lock()
	defer self.Unlock()

	item := self.itemMap[key]

	if item != nil {
		// 玩家之前在榜上：
		// 那就不会掉出榜，最坏是到最后一名
		if item.Value != value {
			updateTime := util.NowInt64()
			update := make(map[string]interface{})
			if item.PlayerUid != playerUid {
				update["playeruid"] = playerUid
			}
			if item.RegionID != regionID {
				update["regionid"] = regionID
			}
			if item.PlayerName != playerName {
				update["playername"] = playerName
			}
			if item.Decoration != decoration {
				update["decoration"] = decoration
			}
			if item.Param1 != param1 {
				update["param1"] = param1
			}
			if item.Param2 != param2 {
				update["param2"] = param2
			}
			if item.Param3 != param3 {
				update["param3"] = param3
			}
			if item.Value != value {
				update["value"] = value
			}
			if item.UpdateTime != updateTime {
				update["updatetime"] = updateTime
			}

			if extraData != nil {
				update["extradata"] = extraData
			}

			if len(update) > 0 {
				if err := self.getDbAgent(self.Name).UpdateId(item.Key, bson.M{"$set": update}); err != nil {
					log.GetLogger().Error("排行榜更新失败！", field.String("name", self.Name), field.String("key", item.Key), field.Any("update", update), field.Error(err))
					return
				}
				log.GetLogger().Info("排行榜更新", field.String("name", self.Name), field.String("key", item.Key), field.Any("update", update))
				item.PlayerUid = playerUid
				item.RegionID = regionID
				item.PlayerName = playerName
				item.Decoration = decoration
				item.Param1 = param1
				item.Param2 = param2
				item.Param3 = param3
				item.Value = value
				item.ExtraData = extraData
				item.UpdateTime = updateTime
				self.needSortFlag = true
			}
		}
	} else {
		// 玩家之前不在榜上：
		// 如果还没有最后一名，或最后一名还没到MaxRank，或最后一名已达到MaxRank但玩家的value超过最后一名了，则新增该玩家
		// 注意：这样的话，ItemMap在下次sort前，如果有多名新增玩家，则ItemMap数量是有可能短暂超出MaxRank的。但在下次sort时，超出的那些，会被移除掉。
		n := len(self.ItemSortedList)
		var tailItem *RankItem
		if n > 0 {
			tailItem = self.ItemSortedList[n-1]
		}
		if tailItem == nil || tailItem.Rank < self.MaxRank || value > tailItem.Value {
			newItem := &RankItem{
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
			}
			if err := self.getDbAgent(self.Name).Insert(newItem); err != nil {
				log.GetLogger().Error("排行榜新增失败！", field.String("name", self.Name), field.Any("item", newItem), field.Error(err))
				return
			}
			if !strings.HasPrefix(newItem.PlayerUid, "robot") {
				log.GetLogger().Info("排行榜新增", field.String("name", self.Name), field.Any("item", newItem))
			}
			self.itemMap[newItem.Key] = newItem
			self.needSortFlag = true
		}
	}
}

// 根据key，将item从榜单移除：从itemMap移除即可，榜单会定时同步生成
func (self *Rank) KickByKey(key string) {
	self.Lock()
	defer self.Unlock()

	item := self.itemMap[key]
	if item == nil {
		return
	}

	if err := self.getDbAgent(self.Name).RemoveId(item.Key); err == nil {
		log.GetLogger().Info("排行榜删除", field.String("name", self.Name), field.Any("item", item))
		delete(self.itemMap, item.Key)
		self.needSortFlag = true
	} else {
		log.GetLogger().Error("排行榜删除失败！", field.String("name", self.Name), field.Any("item", item), field.Error(err))
	}
}

// 玩家名字和装饰物更新到榜单
func (self *Rank) UpdatePlayerInfoByKey(key, playerName, decoration string) {
	self.Lock()
	defer self.Unlock()

	item := self.itemMap[key]
	if item == nil {
		return
	}

	update := make(map[string]interface{})
	if item.PlayerName != playerName {
		update["playername"] = playerName
	}
	if item.Decoration != decoration {
		update["decoration"] = decoration
	}
	if len(update) > 0 {
		if err := self.getDbAgent(self.Name).UpdateId(item.Key, bson.M{"$set": update}); err != nil {
			log.GetLogger().Error("排行榜更新失败！", field.String("name", self.Name), field.String("key", item.Key), field.Any("update", update), field.Error(err))
			return
		}
		log.GetLogger().Info("排行榜更新", field.String("name", self.Name), field.String("key", item.Key), field.Any("update", update))
		item.PlayerName = playerName
		item.Decoration = decoration

		// 榜单上的名字也修改
		if index, ok := self.ItemKeyToIndex[key]; ok {
			if index >= 0 || index < len(self.ItemSortedList) {
				sortedItem := self.ItemSortedList[index]
				sortedItem.PlayerName = playerName
				sortedItem.Decoration = decoration
			}
		}
	}
}

// 清空榜单
func (self *Rank) Clear() {
	self.Lock()
	defer self.Unlock()
	if _, err := self.getDbAgent(self.Name).RemoveAll(nil); err != nil {
		log.GetLogger().Error("排行榜清空表失败！", field.String("name", self.Name), field.Error(err))
		return
	}
	log.GetLogger().Info("排行榜清空！", field.String("name", self.Name))
	self.itemMap = make(map[string]*RankItem)
	self.needSortFlag = true
}
