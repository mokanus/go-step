package rk

type RankItem struct {
	Key        string `bson:"_id"`
	PlayerUid  string
	RegionID   int32
	PlayerName string
	Decoration string
	Value      int64
	Param1     int32
	Param2     int64
	Param3     int32
	UpdateTime int64
	ExtraData  []byte
	Rank       int32 `bson:"-"`
}

func (r *RankItem) Copy() *RankItem {
	return &RankItem{
		Key:        r.Key,
		PlayerUid:  r.PlayerUid,
		RegionID:   r.RegionID,
		PlayerName: r.PlayerName,
		Decoration: r.Decoration,
		Value:      r.Value,
		Param1:     r.Param1,
		Param2:     r.Param2,
		Param3:     r.Param3,
		ExtraData:  r.ExtraData,
		UpdateTime: r.UpdateTime,
		Rank:       r.Rank,
	}
}
