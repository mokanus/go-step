package app

import (
	"encoding/binary"
	"fmt"
)

// ----------------------------------------------------------------------------
// REQ收发消息的打包和解包
// ----------------------------------------------------------------------------
// messageMark   -> 1位，0:1
// messageMode   -> 1位，1:2
// pktSize       -> 4位，2:6（注意：这里的pktSize是指pktCode + pktBody的总大小，也就是pktBody的字节数是pktSize - 2）
// pktCode       -> 2位，6:8
// rpcBody       -> x位，8:8+x
func PackMessage(pktCode uint16, pktBody []byte) []byte {
	bodyLen := len(pktBody)
	size := 8 + bodyLen
	message := make([]byte, size, size)

	message[0] = reqMessageMark
	message[1] = reqMessageModeDefault
	binary.LittleEndian.PutUint32(message[2:6], uint32(2+bodyLen))
	binary.LittleEndian.PutUint16(message[6:8], pktCode)

	if bodyLen > 0 {
		copy(message[8:8+bodyLen], pktBody)
	}

	return message
}

func UnpackMessage(message []byte) (pktCode uint16, pktBody []byte, err error) {
	l := uint64(len(message))

	if l < 8 {
		err = fmt.Errorf("pkt message too small: %d", l)
		return
	}

	if message[0] != reqMessageMark {
		err = fmt.Errorf("unexpected pkt mark: %d", message[0])
		return
	}

	if message[1] != reqMessageModeDefault {
		err = fmt.Errorf("unexpected pkt mode: %d", message[1])
		return
	}

	pktSize := binary.LittleEndian.Uint32(message[2:6])

	expectMessageSize := 6 + uint64(pktSize)

	if l != expectMessageSize {
		err = fmt.Errorf("pkt message size miss match head, expect: %d, in fact: %d", expectMessageSize, l)
		return
	}

	pktCode = binary.LittleEndian.Uint16(message[6:8])
	pktBody = message[8:]

	return
}

// ----------------------------------------------------------------------------
// RPC请求消息的打包和解包
// ----------------------------------------------------------------------------
// messageType   -> 1位，0:1
// callID        -> 2位，1:3
// rpcCode       -> 2位，3:5
// channelUidLen -> 1位，5:6
// rpcBodyLen    -> 4位，6:10
// uid           -> x位，10:10+x
// rpcBody       -> y位，10+x:10+x+y
func PackRpcRequest(callID uint16, channelUid string, rpcCode uint16, rpcBody []byte) []byte {
	channelUidLen := len(channelUid)
	rpcBodyLen := len(rpcBody)
	size := 10 + channelUidLen + rpcBodyLen
	message := make([]byte, size, size)

	message[0] = rpcRequestMessage
	binary.LittleEndian.PutUint16(message[1:3], callID)
	binary.LittleEndian.PutUint16(message[3:5], rpcCode)
	message[5] = uint8(channelUidLen)
	binary.LittleEndian.PutUint32(message[6:10], uint32(rpcBodyLen))
	if channelUidLen > 0 {
		copy(message[10:10+channelUidLen], []byte(channelUid))
	}
	if rpcBodyLen > 0 {
		copy(message[10+channelUidLen:10+channelUidLen+rpcBodyLen], rpcBody)
	}
	return message
}

func UnpackRpcRequest(message []byte) (callID uint16, channelUid string, rpcCode uint16, rpcBody []byte, err error) {
	l := uint64(len(message))

	if l < 10 {
		err = fmt.Errorf("rpc message too small: %d", l)
		return
	}

	callID = binary.LittleEndian.Uint16(message[1:3])
	rpcCode = binary.LittleEndian.Uint16(message[3:5])
	channelUidLen := message[5]
	rpcBodyLen := binary.LittleEndian.Uint32(message[6:10])

	expectMessageSize := 10 + uint64(channelUidLen) + uint64(rpcBodyLen)

	if l != expectMessageSize {
		err = fmt.Errorf("rpc message size miss match head, expect: %d, in fact: %d", expectMessageSize, l)
		return
	}

	if channelUidLen > 0 {
		channelUid = string(message[10 : 10+channelUidLen])
	}
	if rpcBodyLen > 0 {
		rpcBody = message[10+channelUidLen:]
	}

	return
}

// ----------------------------------------------------------------------------
// RPC请求响应的打包和解包
// ----------------------------------------------------------------------------
// messageType -> 1位，0:1
// callID      -> 2位，1:3
// rspCode     -> 1位，3:4
// rspBodyLen  -> 4位，4:8
// rspBody     -> x位，8:8+x
func PackRpcResponse(callID uint16, rspCode uint8, rspBody []byte) []byte {
	rspBodyLen := len(rspBody)
	size := 8 + rspBodyLen
	message := make([]byte, size, size)

	message[0] = rpcResponseMessage
	binary.LittleEndian.PutUint16(message[1:3], callID)
	message[3] = uint8(rspCode)
	binary.LittleEndian.PutUint32(message[4:8], uint32(rspBodyLen))
	if rspBodyLen > 0 {
		copy(message[8:8+rspBodyLen], rspBody)
	}
	return message
}

func UnpackRpcResponseCallID(message []byte) (callID uint16, err error) {
	l := len(message)

	if l < 8 {
		err = fmt.Errorf("rpc message too small: %d", l)
		return
	}

	callID = binary.LittleEndian.Uint16(message[1:3])

	return
}

func UnpackRpcResponseCodeAndBody(message []byte) (rspCode uint8, rspBody []byte, err error) {
	l := uint64(len(message))

	if l < 8 {
		err = fmt.Errorf("rpc message too small: %d", l)
		return
	}

	rspCode = message[3]
	rspBodyLen := binary.LittleEndian.Uint32(message[4:8])

	expectMessageSize := 8 + uint64(rspBodyLen)

	if l != expectMessageSize {
		err = fmt.Errorf("rpc message size miss match head, expect: %d, in fact: %d", expectMessageSize, l)
		return
	}

	if rspBodyLen > 0 {
		rspBody = message[8:]
	}

	return
}
