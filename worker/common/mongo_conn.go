package common

import (
	"time"

	"gopkg.in/mgo.v2"

	"crontab/worker/logger"
)

var (
	GMgo *MongoMgr
)

// MongoMgr mongodb日志管理
type MongoMgr struct {
	S *mgo.Session
}

func (_self *MongoMgr) connect(collection string) (*mgo.Session, *mgo.Collection) {
	ms := _self.S.Copy()
	c := ms.DB(GConfig.MongoDB.DB).C(collection)
	ms.SetMode(mgo.Monotonic, true)
	return ms, c
}

func (_self *MongoMgr) Find(collection string, filter interface{}, limit int, skip int, result interface{}) (err error) {
	ms, c := _self.connect(collection)
	defer ms.Close()

	if err = c.Find(filter).Limit(limit).Skip(skip).All(result); err != nil {
		logger.Error.Println("数据查询失败: ", err)
		return
	}
	return
}

func (_self *MongoMgr) Insert(collection string, doc interface{}) (err error) {
	ms, c := _self.connect(collection)
	defer ms.Close()

	if err = c.Insert(doc); err != nil {
		logger.Error.Println("执行日志插入失败: ", err)
		return
	}
	return
}

func (_self *MongoMgr) InsertMany(collection string, docs []interface{}) (err error) {
	ms, c := _self.connect(collection)
	defer ms.Close()

	if err = c.Insert(docs...); err != nil {
		logger.Error.Println("执行日志插入失败: ", err, docs)
		return
	}
	return
}

func InitMongoConn() (err error) {

	var (
		s *mgo.Session
	)

	dialInfo := &mgo.DialInfo{
		Addrs:     []string{GConfig.MongoDB.Addr},                              //数据库地址 dbhost: mongodb://user@123456:127.0.0.1:27017
		Timeout:   time.Duration(GConfig.MongoDB.ConnectTimeout) * time.Second, // 连接超时时间 Millisecond 毫秒
		Source:    GConfig.MongoDB.AuthDB,                                      // 设置权限的数据库 authdb: admin
		Username:  GConfig.MongoDB.Username,                                    // 设置的用户名 authuser: user
		Password:  GConfig.MongoDB.Password,                                    // 设置的密码 authpass: 123456
		PoolLimit: GConfig.MongoDB.PoolLimit,                                   // 连接池的数量 poollimit: 100
	}
	if s, err = mgo.DialWithInfo(dialInfo); err != nil {
		logger.Error.Printf("mongodb 连接建立失败: %s ", err)
		return
	}

	GMgo = &MongoMgr{
		S: s,
	}
	return
}
