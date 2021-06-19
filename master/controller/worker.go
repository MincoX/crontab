package controller

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"crontab/master/response"
	"crontab/master/service"
)

// WorkerList 获取所有 worker 节点
func WorkerList(ctx *gin.Context) {
	var (
		err         error
		pageSize    int
		currentPage int
		totalCount  int
	)

	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("pageSize", "8"))
	currentPage, _ = strconv.Atoi(ctx.DefaultQuery("currentPage", "1"))

	workerArr := make([]map[string]string, 10)
	if workerArr, err = service.GWorkerSer.ListWorkers(); err != nil {
		response.Fail(ctx, fmt.Sprintf("查询 worker 节点失败: %s", err), nil)
		return
	}

	totalCount = len(workerArr)
	if len(workerArr) >= (currentPage-1)*pageSize+pageSize {
		workerArr = workerArr[(currentPage-1)*pageSize : pageSize]
	} else {
		workerArr = workerArr[(currentPage-1)*pageSize : totalCount]
	}

	response.Success(ctx, gin.H{"workers": workerArr, "totalCount": totalCount}, nil)
	return
}
