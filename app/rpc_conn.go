package app

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"go-step/log"
	"go-step/log/ubiquitous/log/field"
	"io"
	"runtime/debug"
	"sync"
	"time"
)

type RpcConn struct {
	name         string
	socket       *websocket.Conn
	socketLocker *sync.Mutex
	connected    bool

	rpcSerial        uint16
	rpcWaitMap       map[uint16]chan []byte
	rpcWaitMapLocker *sync.Mutex
}

func NewRpcConn(socket *websocket.Conn, connected bool) *RpcConn {
	return &RpcConn{
		name:             "waiting",
		socket:           socket,
		socketLocker:     new(sync.Mutex),
		connected:        connected,
		rpcSerial:        1,
		rpcWaitMap:       make(map[uint16]chan []byte),
		rpcWaitMapLocker: new(sync.Mutex),
	}
}

func (self *RpcConn) close() {
	if !self.connected {
		return
	}
	self.connected = false
	self.socket.Close()
}

func (self *RpcConn) dial(addr string) {
	zoneName := fmt.Sprintf("%s%d", Type, ID)

	var serviceName string
	if Conf.ServiceName == ServiceSeason {
		serviceName = fmt.Sprintf("%s%d", ServiceSeason, Conf.Season)
	} else {
		serviceName = Conf.ServiceName
	}

	header := map[string][]string{
		"X-My-Rpc-Service-Name": {zoneName, serviceName},
	}
	socket, _, err := websocket.DefaultDialer.Dial(addr, header)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("定时检查时，新建区服RPC[%s]连接失败！原因：%v", addr, err))
		return
	}

	self.socket = socket
	self.name = socket.RemoteAddr().String()
	self.connected = true
	log.GetLogger().Info(fmt.Sprintf("定时检查时，连接区服RPC[%s]成功！", addr))

	go func() {
		if err := recover(); err != nil {
			log.GetLogger().Error(fmt.Sprintf(fmt.Sprintf("区服RPC[%s]连接崩溃了: %v\n%s", addr, err, debug.Stack())))
		}
		self.work()
	}()
}

func (self *RpcConn) work() {
	defer func() {
		self.connected = false
		self.socket.Close()
	}()

	for {
		// 疑问1：如果没有读取到消息，是否会一直在这里阻塞？
		// 疑问2：是否会读到空的message？去看底层，如果一定不会返回空message，下面的判空就可以去掉
		_, message, err := self.socket.ReadMessage()
		if err != nil {
			self.doError(err)
			return
		}

		if len(message) == 0 {
			log.GetLogger().Debug("websocket读取到空的message!")
			continue
		}

		switch message[0] {
		case rpcRequestMessage:
			rpcID, channelUid, pktType, pktBody, err := UnpackRpcRequest(message)
			if err != nil {
				self.doError(err)
				return
			}
			self.doRequest(rpcID, string(channelUid), pktType, pktBody)
		case rpcResponseMessage:
			callID, err := UnpackRpcResponseCallID(message)
			if err != nil {
				self.doError(err)
				return
			}
			self.rspRpcWait(callID, message)
		default:
			log.GetLogger().Error(fmt.Sprintf("出错断开！原因：读取到无效的packetType->%d", message[0]))
			return
		}
	}
}

func (self *RpcConn) call(channelUid string, rpcCode uint16, rpcBody []byte) ([]byte, error) {
	if !self.connected {
		return nil, errors.New("连接断开了")
	}

	wait := make(chan []byte, 1)

	callID, err := self.addRpcWait(wait)
	if err != nil {
		return nil, err
	}

	// 当网络写入失败，不管是什么原因，都把socket给关掉。如果是gameApp关掉了serviceConn，对应的serviceApp会尝试再来连接；如果是serviceApp关掉了regionConn，serviceApp会尝试重新建立连接。
	if err := self.writeMessage(PackRpcRequest(callID, channelUid, rpcCode, rpcBody)); err != nil {
		self.socket.Close()
		self.connected = false
		self.delRpcWait(callID)
		return nil, errors.New("网络写入失败了")
	}

	select {
	case <-time.After(time.Second * rpcCallTimeout):
		self.delRpcWait(callID)
		return nil, errors.New("请求超时了")
	case message := <-wait:
		if message == nil {
			return nil, errors.New("网络出错了")
		}
		return message, nil
	}
}

func (self *RpcConn) cast(channelUid string, rpcCode uint16, rpcBody []byte) error {
	if !self.connected {
		return errors.New("连接断开了")
	}

	// 当网络写入失败，不管是什么原因，都把socket给关掉。如果是gameApp关掉了serviceConn，对应的serviceApp会尝试再来连接；如果是serviceApp关掉了regionConn，serviceApp会尝试重新建立连接。
	if err := self.writeMessage(PackRpcRequest(0, channelUid, rpcCode, rpcBody)); err != nil {
		self.socket.Close()
		self.connected = false
		return errors.New("网络写入失败了")
	}

	return nil
}

func (self *RpcConn) resp(callID uint16, rspCode uint8, rspBody []byte) error {
	if !self.connected {
		return errors.New("连接断开了")
	}

	if err := self.writeMessage(PackRpcResponse(callID, rspCode, rspBody)); err != nil {
		self.socket.Close()
		self.connected = false
		return errors.New("网络写入失败了")
	}

	return nil
}

func (self *RpcConn) writeMessage(message []byte) error {
	self.socketLocker.Lock()
	defer self.socketLocker.Unlock()
	return self.socket.WriteMessage(websocket.BinaryMessage, message)
}

func (self *RpcConn) doError(err error) {
	if err == io.EOF {
		log.GetLogger().Info("由对方断开了！")
		return
	}
	log.GetLogger().Error("出错断开！", field.Error(err))
}

func (self *RpcConn) doRequest(callID uint16, channelUid string, rpcType uint16, rpcBody []byte) {
	channel := getChannel(channelUid)

	select {
	case channel.rpcQueue <- NewRPC(self, callID, channelUid, rpcType, rpcBody):
		return
	default:
		log.GetLogger().Error("rpc channel full!", field.String("channelUid", channelUid))
	}
}

func (self *RpcConn) addRpcWait(wait chan []byte) (id uint16, err error) {
	self.rpcWaitMapLocker.Lock()
	defer self.rpcWaitMapLocker.Unlock()

	if self.rpcWaitMap[self.rpcSerial] != nil {
		err = errors.New("请求队列过长")
		return
	}

	id = self.rpcSerial
	self.rpcWaitMap[id] = wait

	self.rpcSerial++
	if self.rpcSerial > rpcCallMaxConcurrent {
		self.rpcSerial = 1
	}

	return
}

func (self *RpcConn) delRpcWait(id uint16) {
	self.rpcWaitMapLocker.Lock()
	defer self.rpcWaitMapLocker.Unlock()

	wait := self.rpcWaitMap[id]
	if wait == nil {
		return
	}

	close(wait)
	delete(self.rpcWaitMap, id)
}

func (self *RpcConn) rspRpcWait(id uint16, message []byte) {
	self.rpcWaitMapLocker.Lock()
	defer self.rpcWaitMapLocker.Unlock()

	wait := self.rpcWaitMap[id]
	if wait == nil {
		return
	}

	wait <- message

	close(wait)
	delete(self.rpcWaitMap, id)
}
