package app

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"go-step/log"
	"go-step/log/ubiquitous/log/field"
	"go-step/util"
	"io"
	"strings"
	"time"
)

type ReqConn struct {
	name      string
	socket    *websocket.Conn
	connected bool
	session   map[string]string // 可在连接上绑定会话数据
}

func NewReqConn(socket *websocket.Conn) *ReqConn {
	return &ReqConn{
		name:      socket.RemoteAddr().String(),
		socket:    socket,
		connected: true,
		session:   make(map[string]string),
	}
}

// 业务逻辑中关闭连接，就不需要回调connLostHandler了；而底层框架中关闭连接，说明连接是被动断开的，此时要回调connLostHandler
// 因此区分一下，内部使用close，外部使用Close
func (self *ReqConn) Close() {
	if !self.connected {
		return
	}
	self.connected = false
	self.socket.Close()
}

func (self *ReqConn) Send(pktType ProtoCode, pktData proto.Message) {
	pktBody, err := util.PbMarshal(pktData)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] -> %v 失败：%v", self.name, pktType, err))
		return
	}

	pktCode := pktType.Value()

	log.GetLogger().Info("notify", field.String("player", self.name), field.Any("method", pktType), field.Any("resp", pktData))

	self.SendMessage(PackMessage(pktCode, pktBody))
}

func (self *ReqConn) SendWithTimeout(pktType ProtoCode, pktData proto.Message) {
	pktBody, err := util.PbMarshal(pktData)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] -> %v 失败：%v", self.name, pktType, err))
		return
	}

	if !self.connected {
		log.GetLogger().Error(fmt.Sprintf("[%s] 发送失败：网络已断开", self.name))
		return
	}

	// 偏临时的代码，目前专用于发顶号消息，防止顶号卡死
	self.socket.SetWriteDeadline(time.Now().Add(time.Second))

	if err := self.socket.WriteMessage(websocket.BinaryMessage, PackMessage(pktType.Value(), pktBody)); err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] 发送失败：网络已断开", self.name))
		// 注意：这里一定不能用self.close()，而要用self.Close()，避免去调用lostHandler，否则玩家业务逻辑层面，会有死锁
		// 而lostHandler，会在work协程中被调用！
		self.Close()
	}
}

func (self *ReqConn) SendMessage(message []byte) {
	if !self.connected {
		log.GetLogger().Error(fmt.Sprintf("[%s] 发送失败：网络已断开", self.name))
		return
	}

	self.socket.SetWriteDeadline(time.Now().Add(time.Second))

	if err := self.socket.WriteMessage(websocket.BinaryMessage, message); err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] 发送失败：%v", self.name, err))
		// 注意：这里一定不能用self.close()，而要用self.Close()，避免去调用lostHandler，否则玩家业务逻辑层面，会有死锁
		// 而lostHandler，会在work协程中被调用！
		self.Close()
	}
}

func (self *ReqConn) SetSession(k, v string) {
	self.session[k] = v
}

func (self *ReqConn) GetSession(k string) string {
	return self.session[k]
}

func (self *ReqConn) GetIP() string {
	ipPort := strings.Split(self.socket.RemoteAddr().String(), ":")
	if len(ipPort) != 2 {
		return ""
	}
	return ipPort[0]
}

func (self *ReqConn) GetName() string {
	return self.name
}

func (self *ReqConn) SetName(newName string) {
	self.name = newName
}

func (self *ReqConn) work() {
	defer self.disconnect()

	log.GetLogger().Info("连接开始工作...", field.Any("addr", self.socket.RemoteAddr()))
	for {
		err := self.socket.SetReadDeadline(time.Now().Add(time.Second * reqSocketReadTimeout))
		if err != nil {
			return
		}

		_, message, err := self.socket.ReadMessage()
		if err != nil {
			self.doError(err)
			return
		}

		pktType, pktBody, err := UnpackMessage(message)
		if err != nil {
			self.doError(err)
			return
		}

		self.doRequest(pktType, pktBody)
	}
}

func (self *ReqConn) disconnect() {
	if !self.connected {
		return
	}
	self.connected = false
	self.socket.Close()
	if cnnlostReqHandler != nil {
		cnnlostReqHandler(NewREQ(self, 0, nil))
	}
}

func (self *ReqConn) doError(err error) {
	if !self.connected {
		log.GetLogger().Info(fmt.Sprintf("[%s] 由服务端断开了！", self.socket.RemoteAddr()))
		return
	}
	if err == io.EOF || websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		log.GetLogger().Info(fmt.Sprintf("[%s] 由客户端断开了！", self.socket.RemoteAddr()))
		return
	}
	log.GetLogger().Info(fmt.Sprintf("[%s] 出错断开！原因：%+v", self.socket.RemoteAddr(), err.Error()))
}

func (self *ReqConn) doRequest(pktType uint16, pktBody []byte) {
	if handler, ok := reqHandlers[pktType]; ok {
		handler(NewREQ(self, pktType, pktBody))
	} else if defaultReqHandler != nil {
		defaultReqHandler(NewREQ(self, pktType, pktBody))
	} else {
		log.GetLogger().Error(fmt.Sprintf("[%s] REQ请求[%d]没有对应的处理函数！", self.socket.RemoteAddr(), pktType))
	}
}
