package app

import (
	"github.com/golang/protobuf/proto"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/log/ubiquitous/log/field"
	"strconv"
)

type REQ struct {
	conn    *ReqConn
	pktType uint16
	pktName string
	pktBody []byte
}

func NewREQ(conn *ReqConn, pktType uint16, pktBody []byte) *REQ {
	pktName, ok := protoNames[pktType]
	if !ok {
		pktName = strconv.Itoa(int(pktType))
	}

	return &REQ{
		conn:    conn,
		pktType: pktType,
		pktName: pktName,
		pktBody: pktBody,
	}
}

func (self *REQ) Conn() *ReqConn {
	return self.conn
}

func (self *REQ) Type() uint16 {
	return self.pktType
}

func (self *REQ) Body() []byte {
	return self.pktBody
}

func (self *REQ) Read(msg proto.Message) bool {
	if msg != nil {
		if err := proto.Unmarshal(self.pktBody, msg); err != nil {
			log.GetLogger().Error("request", field.String("player", self.conn.name), field.String("type", self.pktName), field.Any("params", msg), field.Error(err))
			return false
		}
	}

	// 调试日志
	log.GetLogger().Info("request", field.String("player", self.conn.name), field.String("method", self.pktName), field.Any("params", msg))

	return true
}
