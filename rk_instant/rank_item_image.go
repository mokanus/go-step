package rk_instant

const (
	ImageFlagNoChange = 0
	ImageFlagToUpsert = 1
	ImageFlagToRemove = 2
)

// Item的入库镜像
type RankItemImage struct {
	Item *RankItem
	Flag int
}

func NewRankItemImage(key string) *RankItemImage {
	return &RankItemImage{
		Item: &RankItem{
			Key: key,
		},
	}
}

func (self *RankItemImage) CopyFrom(item *RankItem) {
	// 先记为ImageFlagNoChange
	self.Flag = ImageFlagNoChange

	// 然后逐个字段比对和更新，只要有字段值有变，则Flag标记为ImageFlagToUpsert
	imageItem := self.Item
	if item.PlayerUid != imageItem.PlayerUid {
		imageItem.PlayerUid = item.PlayerUid
		self.Flag = ImageFlagToUpsert
	}
	if item.RegionID != imageItem.RegionID {
		imageItem.RegionID = item.RegionID
		self.Flag = ImageFlagToUpsert
	}
	if item.PlayerName != imageItem.PlayerName {
		imageItem.PlayerName = item.PlayerName
		self.Flag = ImageFlagToUpsert
	}
	if item.Decoration != imageItem.Decoration {
		imageItem.Decoration = item.Decoration
		self.Flag = ImageFlagToUpsert
	}
	if item.Value != imageItem.Value {
		imageItem.UpdateTime = item.UpdateTime
		imageItem.Value = item.Value
		imageItem.ExtraData = item.ExtraData
		self.Flag = ImageFlagToUpsert
	}
	if item.Param1 != imageItem.Param1 {
		imageItem.Param1 = item.Param1
		self.Flag = ImageFlagToUpsert
	}
	if item.Param2 != imageItem.Param2 {
		imageItem.Param2 = item.Param2
		self.Flag = ImageFlagToUpsert
	}
	if item.Param3 != imageItem.Param3 {
		imageItem.Param3 = item.Param3
		self.Flag = ImageFlagToUpsert
	}
}
