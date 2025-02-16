package app

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"go-step/log"
	"strconv"
)

type RPC struct {
	method     string
	conn       *RpcConn
	callID     uint16
	channelUid string
	rpcType    uint16
	rpcName    string
	rpcBody    []byte
}

func NewRPC(conn *RpcConn, callID uint16, channelUid string, rpcType uint16, rpcBody []byte) *RPC {
	rpcName, ok := protoNames[rpcType]
	if !ok {
		rpcName = strconv.Itoa(int(rpcType))
	}

	rpc := &RPC{
		conn:       conn,
		callID:     callID,
		channelUid: channelUid,
		rpcType:    rpcType,
		rpcName:    rpcName,
		rpcBody:    rpcBody,
	}
	if callID > 0 {
		rpc.method = "CALL"
	} else {
		rpc.method = "CAST"
	}
	return rpc
}

func (self *RPC) ChannelUid() string {
	return self.channelUid
}

func (self *RPC) Name() string {
	if self.conn == nil {
		return "localhost"
	} else {
		return self.conn.name
	}
}

func (self *RPC) Type() uint16 {
	return self.rpcType
}

func (self *RPC) Body() []byte {
	return self.rpcBody
}

func (self *RPC) Read(msg proto.Message) bool {
	if msg != nil {
		if err := proto.Unmarshal(self.rpcBody, msg); err != nil {
			log.GetLogger().Error(fmt.Sprintf("[%s:%s] Recv: %v(%s)：%s", self.Name(), self.channelUid, self.rpcName, self.method, err))
			return false
		}
	}

	if self.rpcName != "ARENA_SYNC_ANSWER" && self.rpcName != "MINING_SYNC_ANSWER" &&
		self.rpcName != "ACTIVITY_SYNC" && self.rpcName != "CHAMP_SYNC_CAST" && self.rpcName != "SLG_PLAYER_ONLINE_BEAT_CAST" {
		if self.rpcName == "MINING_SLOT_NOTIFY" || self.rpcName == "TRANSMIT" {
			log.GetLogger().Info(fmt.Sprintf("[%s:%s] Recv: %v(%s)", self.Name(), self.channelUid, self.rpcName, self.method))
		} else {
			// 调试日志
			log.GetLogger().Info(fmt.Sprintf("[%s:%s] Recv: %v(%s) %v", self.Name(), self.channelUid, self.rpcName, self.method, msg))
		}
	}

	return true
}

func (self *RPC) RspError(msg string) {
	if self.conn == nil || self.callID == 0 {
		return
	}

	log.GetLogger().Error(fmt.Sprintf("[%s:%s] Resp: %v Error: %s", self.Name(), self.channelUid, self.rpcName, msg))

	if err := self.conn.resp(self.callID, 2, []byte(msg)); err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s:%s] 响应异常 %v: %v", self.Name(), self.channelUid, self.rpcName, err))
	}
}

func (self *RPC) RspFail(msg string) {
	if self.conn == nil || self.callID == 0 {
		return
	}

	log.GetLogger().Info(fmt.Sprintf("[%s:%s] Resp: %v Fail: %s", self.Name(), self.channelUid, self.rpcName, msg))

	if err := self.conn.resp(self.callID, 1, []byte(msg)); err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s:%s] 响应异常 %v: %v", self.Name(), self.channelUid, self.rpcName, err))
	}
}

func (self *RPC) RspResult(rspData proto.Message) {
	if self.conn == nil || self.callID == 0 {
		return
	}

	if rspData == nil {
		if err := self.conn.resp(self.callID, 0, nil); err != nil {
			log.GetLogger().Error(fmt.Sprintf("[%s:%s] 响应异常 %v: %v", self.Name(), self.channelUid, self.rpcName, err))
		}
		return
	}

	rspBody, err := proto.Marshal(rspData)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s:%s] 响应异常 %v: %v", self.Name(), self.channelUid, self.rpcName, err))
		return
	}

	if len(rspBody) > rpcBodySizeLimit {
		log.GetLogger().Error(fmt.Sprintf("[%s:%s] 响应异常 %v: 包体过大", self.Name(), self.channelUid, self.rpcName))
		return
	}

	log.GetLogger().Info(fmt.Sprintf("[%s:%s] Resp %v: %v", self.Name(), self.channelUid, self.rpcName, rspData))

	if err := self.conn.resp(self.callID, 0, rspBody); err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s:%s] 响应异常 %v: %v", self.Name(), self.channelUid, self.rpcName, err))
		return
	}
}
