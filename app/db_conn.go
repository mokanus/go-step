package app

import (
	"go-step/pkg/github.com/globalsign/mgo"
)

type DbConn struct {
	addr    string
	session *mgo.Session
}
