package app

import (
	"fmt"
	"github.com/mokanus/go-step/log"
	"runtime/debug"
	"time"
)

type Channel struct {
	uid      string
	rpcQueue chan *RPC
	lapQueue chan *LAP
	idle     int
}

func NewChannel(uid string) *Channel {
	var queueSize int
	if uid == "" {
		queueSize = 100000
	} else {
		queueSize = 1024
	}
	w := &Channel{
		uid:      uid,
		rpcQueue: make(chan *RPC, queueSize),
		lapQueue: make(chan *LAP, queueSize),
	}
	w.run()
	return w
}

func (self *Channel) run() {
	go func() {
		ticker := time.NewTicker(time.Minute)

		defer func() {
			if err := recover(); err != nil {
				delChannel(self.uid)
				ticker.Stop()
				log.GetLogger().Error(fmt.Sprintf("[RpcChannel:%s]崩溃了：%v\n%s", self.uid, err, debug.Stack()))
			}
		}()

		rpcDone := false
		lapDone := false
		for {
			select {
			case <-ticker.C:
				self.idle++
				if self.idle >= rpcChannelIdleRecycleMinutes {
					delChannel(self.uid)
					ticker.Stop()
					self.rpcQueue <- nil
					self.lapQueue <- nil
				}
			case rpc := <-self.rpcQueue:
				if rpc != nil {
					self.idle = 0
					self.doRpcRequest(rpc)
				} else {
					rpcDone = true
					if lapDone {
						return
					}
				}
			case lap := <-self.lapQueue:
				if lap != nil {
					self.idle = 0
					self.doLapRequest(lap)
				} else {
					lapDone = true
					if rpcDone {
						return
					}
				}
			}
		}
	}()
}

func (self *Channel) doRpcRequest(rpc *RPC) {
	if handler, ok := rpcHandlers[rpc.rpcType]; ok {
		handler(rpc)
	} else if defaultRpcHandler != nil {
		defaultRpcHandler(rpc)
	} else {
		log.GetLogger().Error(fmt.Sprintf("RPC请求[%d]没有对应的处理函数！", rpc.rpcType))
	}
}

func (self *Channel) doLapRequest(lap *LAP) {
	if lapHandler != nil {
		lapHandler(lap)
	} else {
		log.GetLogger().Error(fmt.Sprintf("LAP请求[%v]没有对应的处理函数！", lap.Fun))
	}
}
