package common

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"crontab/worker/logger"
	"crontab/worker/model"
	"database/sql"
)

var (
	GMsql *MySQLMgr
)

type MySQLMgr struct {
	DB *gorm.DB
}

func InitMySQLConn() (err error) {

	var (
		db    *gorm.DB
		sqlDB *sql.DB
	)

	username := GConfig.MySQL.Username
	password := GConfig.MySQL.Password
	host := GConfig.MySQL.Host
	port := GConfig.MySQL.Port
	database := GConfig.MySQL.Database
	charset := GConfig.MySQL.Charset
	loc := GConfig.MySQL.Loc

	uri := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		username,
		password,
		host,
		port,
		database,
		charset,
		loc,
	)

	//dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	if db, err = gorm.Open(mysql.Open(uri), &gorm.Config{}); err != nil {
		logger.Error.Printf("mysql 连接建立失败: %s ", err)
		return
	}

	if sqlDB, err = db.DB(); err != nil {
		logger.Error.Printf("获取 mysql db 对象失败: %s ", err)
		return
	}

	// 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(GConfig.MySQL.MaxIdleCons)

	// 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(GConfig.MySQL.MaxOpenCons)

	// SetConnMaxLifetime 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Duration(GConfig.MySQL.MaxLifetime) * time.Second)

	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Job{})
	db.AutoMigrate(&model.Log{})

	GMsql = &MySQLMgr{
		DB: db,
	}
	return
}
