package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"

	"crontab/master/common"
	"crontab/master/router"
	"crontab/master/service"
)

var (
	err      error
	confFile string
	eng      *gin.Engine
)

// 解析命令行参数
func initArgs() {
	// go run main.go -config config/config.yaml -xxx 123 -yyy ddd
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

	// redis 连接池
	if err = common.InitRedisConn(); err != nil {
		goto ERR
	}

	// mysql 连接池
	if err = common.InitMySQLConn(); err != nil {
		goto ERR
	}

	//// mongodb 连接池
	//if err = common.InitMongoConn(); err != nil {
	//	goto ERR
	//}

	//  JobService 任务管理器
	if err = service.InitJobSer(); err != nil {
		goto ERR
	}

	// WorkerService 集群节点管理器
	if err = service.InitWorkerSer(); err != nil {
		goto ERR
	}

	// 启动 HTTP 服务
	eng = gin.Default()
	eng = router.RegisterRoute(eng)
	eng.Run(common.GConfig.Http.Addr)

ERR:
	fmt.Println(err)
}
