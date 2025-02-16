package app

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/mokanus/go-step/log"
	"github.com/mokanus/go-step/log/ubiquitous/log/field"
	"reflect"
	"strings"
)

// 区服数据库的存储辅助设施！可用于收集一批数据变动，再一次性落地到数据库。

type Store struct {
	documentName   string
	collectionName string
	toSave         map[string]interface{}
	toWipe         map[string]interface{}
}

func NewStore(documentName string) *Store {
	return &Store{
		documentName:   documentName,
		collectionName: strings.ToLower(documentName),
	}
}

func (self *Store) Save(docUid string, obj interface{}, fields ...string) {
	if self.toSave == nil {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] Save对象%+v时，toSave为空！", docUid, self.documentName, obj))
		return
	}

	val := reflect.ValueOf(obj)

	if val.Kind() != reflect.Ptr {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的存储对象(%+v)：存储对象只能为结构体指针！", docUid, self.documentName, obj))
		return
	}

	if val.IsNil() {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的存储对象(%+v)：存储对象不能为nil！", docUid, self.documentName, obj))
		return
	}

	val = val.Elem()

	if val.Kind() != reflect.Struct {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的存储对象(%+v)：存储对象只能为结构体指针！", docUid, self.documentName, obj))
		return
	}

	n := val.NumField()
	if n == 0 {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的存储对象(%+v)：存储对象字段数为0！", docUid, self.documentName, obj))
		return
	}

	objName := val.Type().Name()
	if !strings.HasPrefix(objName, self.documentName) {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的存储对象(%+v)：存储对象名未以%s为前缀！", docUid, self.documentName, obj, self.documentName))
		return
	}
	objName = strings.ToLower(strings.TrimPrefix(objName, self.documentName))

	var keyPath string

	tag := val.Type().Field(0).Tag.Get("key")
	switch tag {
	case "0":
		// 单层结构体存储，存储路径为objName.FiledName
		if len(fields) == 0 {
			// 未指定要存储的字段，则整个结构体入库
			self.toSave[objName] = obj
		} else {
			// 指定了要存储的字段，则入库指定的字段
			for _, f := range fields {
				v := val.FieldByName(f)
				if !v.IsValid() {
					log.GetLogger().Error(fmt.Sprintf("[%s-%s] 保存对象(%+v)时，指定的字段[%s]不存在！", docUid, self.documentName, obj, f))
					continue
				}
				self.toSave[objName+"."+strings.ToLower(f)] = v.Interface()
			}
		}
		return
	case "1":
		keyPath = fmt.Sprintf("%s.%v", objName, val.Field(0))
	case "2":
		if n < 2 {
			log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的存储对象(%+v)：key指定了两个字段，但存储对象只有一个字段！", docUid, self.documentName, obj))
			return
		}
		keyPath = fmt.Sprintf("%s.%v.%v", objName, val.Field(0), val.Field(1))
	default:
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的存储对象(%+v)：存储对象首字段要添加tag，格式为：key:0/1/2！", docUid, self.documentName, obj))
		return
	}

	// 多层结构的存储，是不允许指定字段的
	if len(fields) > 0 {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 错误的存储调用(%+v)：多层结构的存储不允许指定字段！", docUid, self.documentName, obj))
		return
	}

	// 多层结构体存储，存储路径为objName.key1Value或objName.key1Value.key2Value
	if self.toWipe != nil && self.toWipe[keyPath] != nil {
		log.GetLogger().Debug(fmt.Sprintf("[%s-%s] 保存%s时，该key已经在toWipe中，则该key从toWipe移除！", docUid, self.documentName, keyPath))
		delete(self.toWipe, keyPath)
	}
	self.toSave[keyPath] = obj
}

func (self *Store) Wipe(docUid string, obj interface{}) {
	if self.toWipe == nil {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] Wipe对象%+v时，wipe为空！", docUid, self.documentName, obj))
		return
	}

	val := reflect.ValueOf(obj)

	if val.Kind() != reflect.Ptr {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：擦除对象只能为结构体指针！", docUid, self.documentName, obj))
		return
	}

	if val.IsNil() {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：擦除对象不能为nil！", docUid, self.documentName, obj))
		return
	}

	val = val.Elem()

	if val.Kind() != reflect.Struct {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：擦除对象只能为结构体指针！", docUid, self.documentName, obj))
		return
	}

	n := val.NumField()
	if n == 0 {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：擦除对象字段数为0！", docUid, self.documentName, obj))
		return
	}

	objName := val.Type().Name()
	if !strings.HasPrefix(objName, self.documentName) {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：擦除对象名未以%s为前缀！", docUid, self.documentName, obj, self.documentName))
		return
	}
	objName = strings.ToLower(strings.TrimPrefix(objName, self.documentName))

	var keyPath string

	tag := val.Type().Field(0).Tag.Get("key")
	switch tag {
	case "0":
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：单层结构体不能执行擦除！", docUid, self.documentName, obj))
		return
	case "1":
		keyPath = fmt.Sprintf("%s.%v", objName, val.Field(0))
	case "2":
		if n < 2 {
			log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：key指定了两个字段，但擦除对象只有一个字段！", docUid, self.documentName, obj))
			return
		}
		keyPath = fmt.Sprintf("%s.%v.%v", objName, val.Field(0), val.Field(1))
	default:
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 非法的擦除对象(%+v)：擦除对象首字段要添加tag，格式为：key:0/1/2！", docUid, self.documentName, obj))
		return
	}

	if self.toSave != nil && self.toSave[keyPath] != nil {
		log.GetLogger().Debug(fmt.Sprintf("[%s-%s] 擦除%s时，该key已经在toWave中，则该key从toSave移除！", docUid, self.documentName, keyPath))
		delete(self.toSave, keyPath)
	}

	self.toWipe[keyPath] = 1
}

func (self *Store) ClearChanges(docUid string) {
	if self.toSave != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 开始新的数据变更记录时，toSave不为空！", docUid, self.documentName))
	}
	if self.toWipe != nil {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 开始新的数据变更记录时，toWipe不为空！", docUid, self.documentName))
	}

	self.toSave = make(map[string]interface{})
	self.toWipe = make(map[string]interface{})
}

func (self *Store) StoreChanges(regionID int32, docUid string) {
	if self.toSave == nil {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 执行部落数据变更落地时，toSave为空！", docUid, self.documentName))
	}
	if self.toWipe == nil {
		log.GetLogger().Error(fmt.Sprintf("[%s-%s] 执行部落数据变更落地时，toWipe为空！", docUid, self.documentName))
	}

	if len(self.toSave) > 0 && len(self.toWipe) > 0 {
		log.GetLogger().Info("db数据更新和清除", field.String("player_uid", docUid), field.String("doc_name", self.documentName), field.Any("save_data", self.toSave), field.Any("wipe_data", self.toWipe))

		if err := RegionDB(regionID, self.collectionName).UpdateId(docUid, bson.M{"$set": self.toSave, "$unset": self.toWipe}); err != nil {
			log.GetLogger().Error(fmt.Sprintf("[%s-%s] 数据库更新失败！原因：%s", docUid, self.documentName, err))
		}
	} else if len(self.toSave) > 0 {
		log.GetLogger().Info("db数据更新", field.String("player_uid", docUid), field.String("doc_name", self.documentName), field.Any("save_data", self.toSave))
		if err := RegionDB(regionID, self.collectionName).UpdateId(docUid, bson.M{"$set": self.toSave}); err != nil {
			log.GetLogger().Error(fmt.Sprintf("[%s-%s] 数据库更新失败！原因：%s", docUid, self.documentName, err))
		}
	} else if len(self.toWipe) > 0 {
		log.GetLogger().Info("db数据清除", field.String("player_uid", docUid), field.String("doc_name", self.documentName), field.Any("save_data", self.toSave))
		if err := RegionDB(regionID, self.collectionName).UpdateId(docUid, bson.M{"$unset": self.toWipe}); err != nil {
			log.GetLogger().Error(fmt.Sprintf("[%s-%s] 数据库更新失败！原因：%s", docUid, self.documentName, err))
		}
	}

	// 无论保存到数据库是成功还是失败，toSave/toWipe都要清除，因为toSave/toWipe是一次性的数据
	self.toSave = nil
	self.toWipe = nil
}
