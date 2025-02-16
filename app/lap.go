package app

type LAP struct {
	ChannelUid  string
	Fun         interface{}
	Arg         interface{}
	Offline     bool
	OfflineInit bool
}

func NewLAP(channelUid string, fun interface{}, arg interface{}, offline bool, offlineInit bool) *LAP {
	return &LAP{
		ChannelUid:  channelUid,
		Fun:         fun,
		Arg:         arg,
		Offline:     offline,
		OfflineInit: offlineInit,
	}
}
