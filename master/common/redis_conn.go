package common

import (
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	GRdb *RedisMgr
)

type RedisMgr struct {
	RDB *redis.Client
}

func InitRedisConn() (err error) {

	addr := GConfig.Redis.Addr
	password := GConfig.Redis.Password
	db := GConfig.Redis.DB

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password, // no password set
		DB:           db,       // use default DB
		ReadTimeout:  time.Duration(GConfig.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(GConfig.Redis.WriteTimeout) * time.Second,
	})

	GRdb = &RedisMgr{RDB: rdb}

	return
}
