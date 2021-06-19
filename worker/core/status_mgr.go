package core

import (
	"gorm.io/gorm"

	"crontab/worker/common"
	"crontab/worker/logger"
	"crontab/worker/model"
)

var (
	GStatusMgr *StatusMgr
)

type StatusMgr struct {
	crontabChan chan *common.JobStatusEvent
	OnceChan    chan *common.JobStatusEvent
}

func (_self *StatusMgr) pushStatusEvent(eve *common.JobStatusEvent, jobType int) {
	if jobType == 0 {
		// 定时任务
		_self.crontabChan <- eve
	} else {
		// 单次任务
		_self.OnceChan <- eve
	}
}

func (_self *StatusMgr) updateStatusLoop() {

	var (
		err         error
		statusEvent *common.JobStatusEvent
	)
	for {
		select {
		case statusEvent = <-_self.crontabChan:
			if err = common.GMsql.DB.Model(&model.Job{}).Where("name = ?", statusEvent.Job.Name).
				Updates(map[string]interface{}{
					"status":    statusEvent.StatusTyp,
					"next_time": statusEvent.NextTime,
					"num":       gorm.Expr("num + ?", common.GetNumField(statusEvent.AddNum)), // 对执行过的任务的执行次数进行加 1
				}).Error; err != nil {
				logger.Error.Printf("同步任务状态失败: ", err)
			}

		case statusEvent = <-_self.OnceChan:
			if err = common.GMsql.DB.Model(&model.Job{}).Where("name = ?", statusEvent.Job.Name).
				Updates(map[string]interface{}{
					//"num":       1,
					"status":    statusEvent.StatusTyp,
					"next_time": statusEvent.NextTime,
				}).Error; err != nil {
				logger.Error.Printf("同步任务状态失败: ", err)
			}
		}
	}
}

func InitStatusMgr() (err error) {

	GStatusMgr = &StatusMgr{
		crontabChan: make(chan *common.JobStatusEvent, 1000),
		OnceChan:    make(chan *common.JobStatusEvent, 1000),
	}

	go GStatusMgr.updateStatusLoop()
	return
}
