package app

import (
	"github.com/globalsign/mgo"
)

type DbConn struct {
	addr    string
	session *mgo.Session
}
