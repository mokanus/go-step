package app

import (
	"github.com/mokanus/go-step/pkg/github.com/globalsign/mgo"
)

type DbAgent struct {
	err        error
	collection *mgo.Collection
}

func (self *DbAgent) FindAll(query interface{}, result interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.Find(query).All(result)
}

func (self *DbAgent) FindSelectAll(query interface{}, selector interface{}, result interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.Find(query).Select(selector).All(result)
}

func (self *DbAgent) FindSortAll(result interface{}, fields ...string) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.Find(nil).Sort(fields...).All(result)
}

func (self *DbAgent) FindLimitAll(query interface{}, result interface{}, limit int) error {
	if self.err != nil {
		return self.err
	}

	return self.collection.Find(query).Limit(limit).All(result)
}

func (self *DbAgent) FindIdOne(id interface{}, result interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.FindId(id).One(result)
}

func (self *DbAgent) FindOne(query interface{}, result interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.Find(query).One(result)
}

func (self *DbAgent) FindSelectOne(query interface{}, selector interface{}, result interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.Find(query).Select(selector).One(result)
}

func (self *DbAgent) Insert(docs ...interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.Insert(docs...)
}

func (self *DbAgent) UpsertId(id interface{}, update interface{}) (*mgo.ChangeInfo, error) {
	if self.err != nil {
		return nil, self.err
	}
	return self.collection.UpsertId(id, update)
}

func (self *DbAgent) UpdateId(id interface{}, update interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.UpdateId(id, update)
}

func (self *DbAgent) RemoveId(id interface{}) error {
	if self.err != nil {
		return self.err
	}
	return self.collection.RemoveId(id)
}

func (self *DbAgent) RemoveAll(selector interface{}) (*mgo.ChangeInfo, error) {
	if self.err != nil {
		return nil, self.err
	}
	return self.collection.RemoveAll(selector)
}
