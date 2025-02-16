package app

import (
	"errors"
	"fmt"
	"go-step/log"
	"go-step/pkg/github.com/golang/protobuf/proto"
	"go-step/util"
)

// 每一次rpc请求（Call、Cast）会创建一个RpcAgent来负责执行请求。
type RpcAgent struct {
	err        error
	name       string
	conn       *RpcConn
	channelUid string
}

func (self *RpcAgent) Call(rpcType ProtoCode, rpcData proto.Message, rpcResp proto.Message) error {
	if self.err != nil {
		return self.err
	}

	rpcBody, err := util.PbMarshal(rpcData)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] call(%v) Error: %v", self.name, rpcType, err))
		return ErrServer
	}

	if len(rpcBody) > rpcBodySizeLimit {
		log.GetLogger().Error(fmt.Sprintf("[%s] call(%v) Error: 包体过大", self.name, rpcType))
		return ErrServer
	}

	if self.conn == nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] call(%v) Error: 连接不存在", self.name, rpcType))
		return ErrServer
	}
	log.GetLogger().Info(fmt.Sprintf("[%s] call(%v) %v", self.name, rpcType, rpcData))

	// 阻塞在此处，直到网络对端返回消息
	message, err := self.conn.call(self.channelUid, rpcType.Value(), rpcBody)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] call(%v) Error: %v", self.name, rpcType, err))
		return ErrServer
	}

	rspCode, rspBody, err := UnpackRpcResponseCodeAndBody(message)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] call(%v) Error: %v", self.name, rpcType, err))
		return ErrServer
	}

	switch rspCode {
	// rspCode为0，代表网络对端返回正常结果
	case 0:
		// 不关心返回，则成功
		if rpcResp == nil {
			log.GetLogger().Info(fmt.Sprintf("[%s] call(%v) Resp: <nil>", self.name, rpcType))
			return nil
		}

		// 关心返回，但是rspBody解析出错了，按对端异常处理
		if err := proto.Unmarshal(rspBody, rpcResp); err != nil {
			log.GetLogger().Error(fmt.Sprintf("[%s] call(%v) Error: 对端返回的结果无法解析：%v", self.name, rpcType, err))
			return ErrServer
		}
		log.GetLogger().Info(fmt.Sprintf("[%s] call(%v) Resp: %v", self.name, rpcType, rpcResp))
		return nil

	// rspCode为1，代表网络对端进入了正常逻辑，但主动返回了失败，rspBody为失败信息
	case 1:
		log.GetLogger().Info(fmt.Sprintf("[%s] call(%v) Fail: %v", self.name, rpcType, string(rspBody)))
		return errors.New(string(rspBody))

	// rspCode既不为0，也不为1，代表网络对端没进入正常逻辑，而是出异常报错了（如：参数协议解析失败等），rspBody是报错信息
	default:
		log.GetLogger().Error(fmt.Sprintf("[%s] call(%v) Error: 对端异常了，异常信息为：%v", self.name, rpcType, string(rspBody)))
		return ErrServer
	}
}

func (self *RpcAgent) Cast(rpcType ProtoCode, rpcData proto.Message) error {
	if self.err != nil {
		return self.err
	}

	rpcBody, err := util.PbMarshal(rpcData)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] cast(%v) Error: %v", self.name, rpcType, err))
		return ErrServer
	}

	if len(rpcBody) > rpcBodySizeLimit {
		log.GetLogger().Error(fmt.Sprintf("[%s] cast(%v) Error: 包体过大", self.name, rpcType))
		return ErrServer
	}

	if self.conn == nil {
		log.GetLogger().Error(fmt.Sprintf("[%s] cast(%v) Error: 连接不存在", self.name, rpcType))
		return ErrServer
	}

	log.GetLogger().Info(fmt.Sprintf("[%s] cast(%v) %v", self.name, rpcType, rpcData))

	if err := self.conn.cast(self.channelUid, rpcType.Value(), rpcBody); err != nil {
		log.GetLogger().Info(fmt.Sprintf("[%s] cast(%v) Error: %v", self.name, rpcType, err))
		return ErrServer
	}

	return nil
}
