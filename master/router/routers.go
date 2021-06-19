package router

import (
	"github.com/gin-gonic/gin"

	"crontab/master/controller"
	"crontab/master/middleware"
)

func RegisterRoute(egn *gin.Engine) *gin.Engine {
	egn.Use(middleware.CORSMiddleware())

	egn.POST("/user/register", controller.Register)
	egn.POST("/user/login", controller.Login)
	egn.GET("/user/logout", middleware.AuthMiddleware(), controller.Logout)
	egn.GET("/user/info", middleware.AuthMiddleware(), controller.UserInfo)
	egn.GET("/user/list", middleware.AuthMiddleware(), controller.UserList)

	egn.GET("/job/list", middleware.AuthMiddleware(), controller.JobList)
	egn.POST("/job/add", middleware.AuthMiddleware(), controller.JobAdd)
	egn.POST("/job/delete", middleware.AuthMiddleware(), controller.JobDelete)
	egn.POST("/job/kill", middleware.AuthMiddleware(), controller.JobKill)
	egn.POST("/job/logs", middleware.AuthMiddleware(), controller.JobLogs)

	egn.GET("/worker/list", middleware.AuthMiddleware(), controller.WorkerList)

	return egn

}
