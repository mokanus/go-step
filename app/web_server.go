package app

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go-step/log"
	"go-step/log/ubiquitous/log/field"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"time"
)

func startWebServer() {
	log.GetLogger().Info("监听网络端口", field.Int32("port", Conf.Port))

	svr := &http.Server{
		Addr: fmt.Sprintf(":%d", Conf.Port),
		Handler: &Acceptor{
			// TODO: 去理解Upgrader的配置
			upgrader: &websocket.Upgrader{
				HandshakeTimeout: time.Second * acceptorHandshakeTimeout,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		},
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.GetLogger().Error(fmt.Sprintf("监听协程崩溃了：%v\n%s", err, debug.Stack()))
			}
		}()
		if err := svr.ListenAndServe(); err != nil {
			panic("监听失败: " + err.Error())
		}
	}()
}

type Acceptor struct {
	upgrader *websocket.Upgrader
}

func (self *Acceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 收到连接，连接未提供Upgrade字段，则肯定不是来自客户端的req连接，也不是来自ZoneApp的rpc连接，那就以http短连接进行处理
	if r.Header["Upgrade"] == nil {
		switch r.URL.Path {
		case "/debug/pprof/":
			pprof.Index(w, r)
		case "/debug/pprof/trace":
			pprof.Trace(w, r)
		case "/debug/pprof/profile":
			pprof.Handler("profile").ServeHTTP(w, r)
		case "/debug/pprof/heap":
			pprof.Handler("heap").ServeHTTP(w, r)
		case "/debug/pprof/block":
			pprof.Handler("block").ServeHTTP(w, r)
		case "/debug/pprof/goroutine":
			pprof.Handler("goroutine").ServeHTTP(w, r)
		case "/debug/pprof/allocs":
			pprof.Handler("allocs").ServeHTTP(w, r)
		case "/debug/pprof/cmdline":
			pprof.Handler("cmdline").ServeHTTP(w, r)
		case "/debug/pprof/threadcreate":
			pprof.Handler("threadcreate").ServeHTTP(w, r)
		case "/debug/pprof/mutex":
			pprof.Handler("mutex").ServeHTTP(w, r)
		default:
			handleQuery(w, r)
		}
		return
	}

	// 2. 收到连接，且连接的Header信息里带有X-My-Rpc-Service-Name-Season，则必定是来自ZoneApp的rpc连接，将连接升级为websocket
	if len(r.Header["X-My-Rpc-Service-Name"]) >= 2 {
		zoneName := r.Header["X-My-Rpc-Service-Name"][0]
		serviceName := r.Header["X-My-Rpc-Service-Name"][1]
		socket, err := self.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.GetLogger().Error(fmt.Sprintf("来自[%s %s]的rpc连接升级websocket失败（Header=%v），原因：%v", r.RemoteAddr, zoneName, r.Header, err))
			return
		}
		log.GetLogger().Info(fmt.Sprintf("收到来自[%s %s]的rpc连接！", r.RemoteAddr, zoneName))
		conn := NewRpcConn(socket, true)
		conn.name = fmt.Sprintf("%s %s", r.RemoteAddr, zoneName)
		addServiceRpcConn(conn, serviceName)
		conn.work()
		return
	}

	// 3. 收到连接，连接带有Upgrade但又没有X-My-Rpc-Service-Name-List，则认为是来自客户端的req连接，将连接升级为websocket
	socket, err := self.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("来自客户端的req连接升级websocket失败（Header=%v），原因：%v", r.Header, err))
		return
	}
	conn := NewReqConn(socket)
	conn.work()
}
