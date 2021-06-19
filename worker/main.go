package main

import (
	"crontab/worker/common"
	"crontab/worker/core"
	"crontab/worker/logger"
	"flag"
	"runtime"
)

var (
	err      error
	confFile string // 配置文件路径
)

// 解析命令行参数
func initArgs() {
	// go run main.go -config ../config.yaml -username MincoX
	flag.StringVar(&confFile, "config", "config.yaml", "go run main.go -config ../config.yaml")
	flag.Parse()
}

// 初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	// 初始化命令行参数
	initArgs()

	// 初始化线程
	initEnv()

	// 加载配置
	if err = common.InitConfig(confFile); err != nil {
		goto ERR
	}

	// mysql 连接池
	if err = common.InitMySQLConn(); err != nil {
		goto ERR
	}

	//// mongo 连接池
	//if err = common.InitMongoConn(); err != nil {
	//	goto ERR
	//}

	// 服务注册
	if err = core.InitRegister(); err != nil {
		goto ERR
	}

	// 启动任务调度器
	if err = core.InitScheduler(); err != nil {
		goto ERR
	}

	// 启动任务监听器
	if err = core.InitJobMgr(); err != nil {
		goto ERR
	}

	// 启动任务状态管理器
	if err = core.InitStatusMgr(); err != nil {
		goto ERR
	}

	// 启动任务执行器
	if err = core.InitExecutor(); err != nil {
		goto ERR
	}

	// 启动日志管理器
	if err = core.InitLogMgr(); err != nil {
		goto ERR
	}

	// 阻塞主协程
	select {}

ERR:
	logger.Error.Println("服务启动失败：", err)
}
