package app

import (
	"github.com/mokanus/go-step/pkg/github.com/globalsign/mgo"
)

type DbConn struct {
	addr    string
	session *mgo.Session
}
