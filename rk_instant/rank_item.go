package rk_instant

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
	ExtraData  []byte
	UpdateTime int64
	RankIndex  int `bson:"-"`
}
