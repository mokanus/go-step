package app

import "errors"

var (
	ErrServer = errors.New("server error")
)

const (
	EnvLocal = "foo"
	EnvDev   = "dev"
	EnvRel   = "rel"
)

const (
	TypeAdmin = "admin"
	TypeFront = "front"
	TypePay   = "pay"
	TypeGm    = "gm"
	TypeGame  = "game"
	TypeZone  = "zone"
)

const (
	ServiceSeason = "Season"
)

const (
	acceptorHandshakeTimeout = 90 //  注：这里的时间会大于心跳超时，也就是用心跳来关闭闲置的连接

	qryAdminToken               = "JxvMeU5t7RwlP" // 所有来自后台的请求，都带上这个Header做合法校验
	qryPostTimeout              = 3               // qry post的超时时长
	qryAuthSessionExpireSeconds = 24 * 3600       // qry登录会话有效期

	rpcCallTimeout       = 2     // rpc call的超时时长
	rpcCallMaxConcurrent = 60000 // rpc call在同一连接下的最大并发请求数

	rpcRequestMessage  uint8 = 1 // rpc消息类型：请求消息
	rpcResponseMessage uint8 = 2 // rpc消息类型：响应消息

	rpcBodySizeLimit = 4 * 1024 * 1024 // 限定rpc body的大小最大为4M

	rpcChannelIdleRecycleMinutes = 10    // rpc channel空闲超过10分钟就回收
	rpcChannelRecvQueueLimit     = 10000 // rpc channel接收队列大小限制

	reqSocketReadTimeout  = 90  // req的socket连接阻塞读数据的超时时间。注意：这里的时间会大于心跳超时，也就是用心跳来关闭闲置的连接
	reqMessageMark        = 218 // 设置一个特殊值来过滤识别合法游戏协议包
	reqMessageModeDefault = 80
)

// 简单的区服信息，用于玩家获取区服列表
type BriefRegionInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	State int64  `json:"state"`
}
