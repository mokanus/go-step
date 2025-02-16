package app

import (
	"fmt"
	"github.com/mokanus/go-step/log"
	"net/http"
)

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		log.GetLogger().Error(fmt.Sprintf("非法的请求方式：%s", r.Method))
		http.Error(w, "not supported method", 500)
		return
	}

	// 先判断是否是public请求
	if handler, ok := qryPublicHandlers[r.URL.Path]; ok {
		handler(NewQRY(r, w, ""))
		return
	}
	if defaultQryPublicHandler != nil {
		defaultQryPublicHandler(NewQRY(r, w, ""))
		return
	}

	// 再判断是否是private请求。注意：所有的private请求都来自AdminApp，都带有X-My-Qry-Admin-Token。没这个字段、字段值错误的，都是非法的请求！
	if len(r.Header["X-My-Qry-Admin-Token"]) == 0 {
		log.GetLogger().Error(fmt.Sprintf("非法的请求路径：%s", r.URL.Path))
		http.Error(w, "not found", 404)
		return
	}
	if r.Header["X-My-Qry-Admin-Token"][0] != qryAdminToken {
		log.GetLogger().Error(fmt.Sprintf("错误的AdminToken：%s", r.Header["X-My-Qry-Admin-Token"][0]))
		http.Error(w, "not found", 404)
		return
	}

	channelUid := ""
	if len(r.Header["X-My-Qry-Channel-Uid"]) > 0 {
		channelUid = r.Header["X-My-Qry-Channel-Uid"][0]
	}

	if handler, ok := qryPrivateHandlers[r.URL.Path]; ok {
		handler(NewQRY(r, w, channelUid))
		return
	}
	if defaultQryPrivateHandler != nil {
		defaultQryPrivateHandler(NewQRY(r, w, channelUid))
		return
	}

	// 是合法的private请求，但还没有对应的处理函数
	log.GetLogger().Error(fmt.Sprintf("QRY请求[%s]没有对应的处理函数！", r.URL.Path))
	http.Error(w, "no handler", 500)
}
