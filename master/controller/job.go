package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"crontab/master/common"
	"crontab/master/logger"
	"crontab/master/model"
	"crontab/master/response"
	"crontab/master/service"
)

// JobList 列举所有crontab任务
func JobList(ctx *gin.Context) {
	var (
		pageSize    int
		currentPage int
		totalCount  int64
		err         error
		reqTyp      string
		whereFilter map[string]interface{}
		jobs        []model.Job
	)

	reqTyp = ctx.Query("reqTyp")
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("pageSize", "8"))
	currentPage, _ = strconv.Atoi(ctx.DefaultQuery("currentPage", "1"))

	// reqTyp：0 表示查询任务状态为已完成或已删除的所有任务；1 表示查询所有可以调度执行的任务
	if reqTyp == "1" {
		// 所有可以调度执行的任务
		whereFilter = map[string]interface{}{"status": []int64{0, 1, 2, 3}}
	} else {
		// 所有已经删除或者单次任务状态为已完成的任务
		whereFilter = map[string]interface{}{"status": []int64{4, 5}}
	}

	// 获取任务列表
	jobDB := common.GMsql.DB.Model(&model.Job{}).Where(whereFilter)
	if jobDB.Error != nil {
		logger.Error.Printf("查询任务列表失败: %s ", err)
		response.Fail(ctx, fmt.Sprintf("任务查询失败： %s", err), nil)
	}
	jobDB.Count(&totalCount)
	// 排序: 即将执行 > 新增加 > 修改 > 单次 > 定时 > 执行次数
	jobDB.Order("next_time desc").Order("id desc").Order("updated_at desc").
		Order("typ").Order("num desc").Offset((currentPage - 1) * pageSize).Limit(pageSize).Find(&jobs)

	response.Success(ctx, gin.H{"totalCount": totalCount, "jobs": jobs}, nil)
	return
}

// JobAdd 保存任务接口 POST job={"name": "job1", "command": "echo hello", "cronExpr": "* * * * *"}
func JobAdd(ctx *gin.Context) {
	var (
		err      error
		sqlRes   *gorm.DB
		postData *common.Job
	)

	name := ctx.PostForm("name")
	command := ctx.PostForm("command")
	cronExpr := ctx.PostForm("cronExpr")
	jobType, _ := strconv.Atoi(ctx.PostForm("typ"))
	user, _ := ctx.Get("user")

	// 保存到 mysql
	if sqlRes = common.GMsql.DB.Create(&model.Job{
		Name:     name,
		Command:  command,
		CronExpr: cronExpr,
		Status:   0, // 待调度
		Typ:      jobType,
		Num:      0,
		UserID:   int(user.(model.User).ID),
	}); sqlRes.Error != nil {
		logger.Error.Printf("新增任务插入 mysql 出错: ", sqlRes.Error)
		response.Fail(ctx, fmt.Sprintf("任务保存失败： %s", err), nil)
		return
	}

	// 保存到etcd
	if _, err = service.GJobSer.AddJob(&common.Job{
		Name:     name,
		Command:  command,
		CronExpr: cronExpr,
		Typ:      jobType,
		Num:      0,
	}); err != nil {
		logger.Error.Printf("新增任务插入 etcd 出错: ", err)
		response.Fail(ctx, fmt.Sprintf("任务保存失败： %s", err), nil)
		// TODO 将 mysql 中的此任务标记删除
		return
	}

	response.Success(ctx, gin.H{"job": postData}, nil)
	return

}

// JobDelete 删除任务接口 POST /job/delete   name=job1
func JobDelete(ctx *gin.Context) {
	var (
		err    error // interface{}
		oldJob *common.Job
	)
	name := ctx.PostForm("name")

	// 删除 etcd 中任务
	if oldJob, err = service.GJobSer.DeleteJob(name); err != nil {
		// TODO error log
		response.Fail(ctx, fmt.Sprintf("etcd 任务删除失败： %s", err), nil)
		return
	}

	if oldJob == nil {
		response.Response(ctx, http.StatusOK, 0, nil, "任务不存在")
		return
	}

	// 删除 mysql 中任务
	if err = common.GMsql.DB.Model(model.Job{}).Where("name = ?", name).Updates(map[string]interface{}{
		"status": common.StatusTyp["已删除"],
	}).Error; err != nil {
		response.Fail(ctx, fmt.Sprintf("mysql 任务删除失败： %s", err), nil)
		return
	}

	// TODO error log
	response.Success(ctx, gin.H{"jobName": oldJob}, nil)
	return

}

// JobKill 强制杀死某个任务 POST /job/kill  name=job1
func JobKill(ctx *gin.Context) {
	var (
		err error
	)

	name := ctx.PostForm("name")

	// 杀死任务
	if err = service.GJobSer.KillJob(name); err != nil {
		// TODO error log
		response.Fail(ctx, fmt.Sprintf("任务强杀失败： %s", err), nil)
		return
	}

	// TODO insert mysql

	// TODO error log
	response.Success(ctx, gin.H{"job": name}, nil)
	return
}

// JobLogs 查询任务日志
func JobLogs(ctx *gin.Context) {
	var (
		pageSize    int
		currentPage int
		totalCount  int64
		err         error
		jobName     string // 任务名字
		logs        []model.Log
	)

	jobName = ctx.PostForm("jobName")
	pageSize, _ = strconv.Atoi(ctx.DefaultPostForm("pageSize", "8"))
	currentPage, _ = strconv.Atoi(ctx.DefaultPostForm("currentPage", "1"))

	logDB := common.GMsql.DB.Model(&model.Log{}).Where("job_name = ?", jobName)
	if logDB.Error != nil {
		logger.Error.Printf("查询日志失败: ", err)
		response.Fail(ctx, fmt.Sprintf("查询日志失败： %s", err), nil)
		return
	}
	logDB.Count(&totalCount)
	logDB.Order("id desc").Order("result").Offset((currentPage - 1) * pageSize).Limit(pageSize).Find(&logs)

	response.Success(ctx, gin.H{"totalCount": totalCount, "logs": logs}, nil)
	return
}
