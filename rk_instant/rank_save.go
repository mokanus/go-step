package rk_instant

import (
	"fmt"
	"go-step/log"
)

func (self *Rank) updateImageMap() {
	self.Lock()
	defer self.Unlock()

	// 情况一：有镜像，而榜单中没有了，那就代表要删除这个镜像，标记为ImageFlagToRemove。
	// 标记为ImageFlagToRemove后，在syncImageMapToDB中，镜像会删除，对应的数据库项也会删除。
	for _, image := range self.imageMap {
		if self.ItemMap[image.Item.Key] == nil {
			image.Flag = ImageFlagToRemove
		}
	}

	// 情况二：榜单中有，而无镜像，那就代表要新建这个镜像，标记为ImageFlagToUpsert。
	// 情况三：榜单中有，已有镜像，那就检查镜像字段值是否有变化，有变化就更新字段值并且将镜像标记为ImageFlagToUpsert。
	// 标记为ImageFlagToUpsert后，在syncImageMapToDB中，镜像会同步到数据库中。
	for _, item := range self.ItemMap {
		image := self.imageMap[item.Key]
		if image == nil {
			image = NewRankItemImage(item.Key)
		}
		image.CopyFrom(item)
		self.imageMap[image.Item.Key] = image
	}
}

func (self *Rank) syncImageMapToDB() {
	// 数据的变更都记录到SaveMap之后，ItemMap就解锁可被更新了。
	// SaveMap再遍历入库，此时SaveMap不用锁住，因为只有这一个协程会访问SaveMap
	for _, image := range self.imageMap {
		imageItem := image.Item
		switch image.Flag {
		case ImageFlagToUpsert:
			log.GetLogger().Debug(fmt.Sprintf("排行榜[%s]镜像%s->%d入库", self.Name, imageItem.Key, imageItem.Value))
			if _, err := self.getDbAgent(self.Name).UpsertId(imageItem.Key, imageItem); err != nil {
				log.GetLogger().Error(fmt.Sprintf("数据库保存%+v失败：%v", imageItem, err))
			}
		case ImageFlagToRemove:
			log.GetLogger().Debug(fmt.Sprintf("排行榜[%s]镜像%s->%d删除", self.Name, imageItem.Key, imageItem.Value))
			delete(self.imageMap, imageItem.Key)
			if err := self.getDbAgent(self.Name).RemoveId(imageItem.Key); err != nil {
				log.GetLogger().Error(fmt.Sprintf("数据库删除%+v失败：%v", imageItem, err))
			}
		}
	}
}
