package app

// 由于逻辑层的协议号是PktType、RpcType类型，类型是在逻辑层定义的，所以必须通过该接口来接收逻辑层的协议号
type ProtoCode interface {
	Value() uint16
}

type Manager interface {
	Init()
	Stop()
}
