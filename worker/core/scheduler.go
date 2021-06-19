package core

import (
	"time"

	"crontab/worker/common"
	"crontab/worker/logger"
	"crontab/worker/model"
)

var (
	GScheduler *Scheduler
)

// Scheduler 任务调度
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              //  etcd任务事件队列
	jobPlanTable      map[string]*common.JobSchedulePlan // 任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo  // 任务执行表
	jobResultChan     chan *common.JobExecuteResult      // 任务结果队列
}

// PushJobEvent 推送任务变化事件
func (_self *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	_self.jobEventChan <- jobEvent
}

// 处理任务事件
func (_self *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		err             error
		jobExecuting    bool
		jobExisted      bool
		jobSchedulePlan *common.JobSchedulePlan
		jobExecuteInfo  *common.JobExecuteInfo
	)

	switch jobEvent.EventType {
	case common.JobEventSave: // 保存任务事件
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			logger.Error.Printf("构造调度任务失败: ", err)
			return
		}
		_self.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
		GStatusMgr.pushStatusEvent(common.BuildStatusEvent(common.StatusTyp["待执行"], jobEvent.Job, jobSchedulePlan.NextTime, false), jobEvent.Job.Typ)
		logger.Info.Println(jobEvent.Job.Name, ": 已同步至任务调度表！")

	case common.JobEventDelete: // 删除任务事件
		if jobSchedulePlan, jobExisted = _self.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(_self.jobPlanTable, jobEvent.Job.Name)
			GStatusMgr.pushStatusEvent(common.BuildStatusEvent(common.StatusTyp["已删除"], jobEvent.Job, jobSchedulePlan.NextTime, false), jobEvent.Job.Typ)
			logger.Info.Println(jobEvent.Job.Name, ": 任务删除成功！")
			return
		}
		logger.Info.Println(jobEvent.Job.Name, ": 任务不存在，删除失败！")

	case common.JobEventKill: // 强杀任务事件
		// 取消掉 Command 执行, 判断任务是否在执行中
		if jobExecuteInfo, jobExecuting = _self.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() // 触发command杀死shell子进程, 任务得到退出
			GStatusMgr.pushStatusEvent(common.BuildStatusEvent(common.StatusTyp["执行异常"], jobExecuteInfo.Job, jobExecuteInfo.PlanTime, true), jobExecuteInfo.Job.Typ)
			logger.Info.Println(jobEvent.Job.Name, ": 任务强杀成功！")
			return
		}
		logger.Info.Println(jobEvent.Job.Name, ": 任务未运行，强杀失败！")
	}
}

// TrySchedule 重新计算任务调度状态
func (_self *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	// 遍历任务计划表，立即 start 已经过期的任务；
	// 返回下一次执行任务开始时间与当前时间的间隔

	var (
		jobPlan  *common.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)

	// 如果任务计划表为空话，此时设置 60s 后再尝试调度
	if len(_self.jobPlanTable) == 0 {
		scheduleAfter = time.Duration(common.GConfig.Worker.ScheduleSleep) * time.Second
		logger.Info.Printf("暂时没有可调度的任务，下一次调度 %s 后触发！", scheduleAfter.String())
		return
	}

	logger.Debug.Printf("调度表共 %s 个任务待调度 ", len(_self.jobPlanTable))

	// 当前时间
	now = time.Now()
	// 遍历所有任务
	for _, jobPlan = range _self.jobPlanTable {
		// 如果任务下次的执行时间在当前时间之前，说明任务已经过期，立即尝试 start 任务
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			_self.TryStartJob(jobPlan)
			// 更新定时任务下次执行时间，单次任务不需要更新；此时单次任务的下次执行时间还是，此次执行时间
			if jobPlan.Job.Typ == 0 {
				jobPlan.NextTime = jobPlan.Expr.Next(now)
			}
		}

		// 得到任务计划表中最先过期的任务的下次执行时间，只有定时任务的下次执行时间有效；
		// 单次任务的下次执行时间没有更新还是当前次执行时间，所以要求 jobTyp == 0
		if jobPlan.Job.Typ == 0 && (nearTime == nil || jobPlan.NextTime.Before(*nearTime)) {
			nearTime = &jobPlan.NextTime
		}
	}

	if nearTime == nil {
		scheduleAfter = time.Duration(common.GConfig.Worker.ScheduleSleep) * time.Second
	} else {
		// 下次调度间隔（最近要执行的任务调度时间 - 当前时间）
		scheduleAfter = (*nearTime).Sub(now)
	}
	logger.Info.Printf("下一次调度 %s 后触发！", scheduleAfter.String())
	return
}

// TryStartJob 尝试执行任务
func (_self *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
	// 尝试执行任务，因为任务执行时间长短的不确定性，有可能下次执行的时间到了，但是该任务还未执行完成，此时跳过此次调度，不开始新的执行
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)

	// 如果任务正在执行，跳过本次执行
	if jobExecuteInfo, jobExecuting = _self.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		logger.Info.Println(jobPlan.Job.Name, ": 尚未退出，取消本次执行。下次执行时间：", jobPlan.Expr.Next(time.Now()))
		return
	}

	// 将成功执行的任务放入任务执行列表中
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)
	_self.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	GExecutor.ExecuteJob(jobExecuteInfo)
	GStatusMgr.pushStatusEvent(common.BuildStatusEvent(common.StatusTyp["执行中"], jobPlan.Job, jobPlan.NextTime, false), jobPlan.Job.Typ)
	logger.Info.Println(jobPlan.Job.Name, ": 任务执行中！")
}

// PushJobResult 回传任务执行结果
func (_self *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	_self.jobResultChan <- jobResult
}

// 处理任务结果
func (_self *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	var (
		statusTyp int
		err       error
		job       model.Job
		jobLog    *model.Log
	)

	// 从执行表中删除
	delete(_self.jobExecutingTable, result.ExecuteInfo.Job.Name)
	// 单次任务还要从计划表中删除，避免被再次调度到执行表
	if result.ExecuteInfo.Job.Typ == 1 {
		delete(_self.jobPlanTable, result.ExecuteInfo.Job.Name)
		// 从 etcd 中删除，避免新 worker 上线时会将其同步到计划表中去
		if _, err = GJobMgr.DeleteJob(result.ExecuteInfo.Job.Name); err != nil {
			logger.Error.Println(result.ExecuteInfo.Job.Name, ": etcd 中单次任务删除失败！")
		}
	}

	if err = common.GMsql.DB.Where("name=?", result.ExecuteInfo.Job.Name).First(&job).Error; err != nil {
		logger.Error.Println("查询到日对应的任务失败: ", err)
	}
	// 生成执行日志
	if result.Err != common.ErrLockAlreadyRequired {
		jobLog = &model.Log{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.Format("2006/01/02 15:04:05"),
			ScheduleTime: result.ExecuteInfo.RealTime.Format("2006/01/02 15:04:05"),
			StartTime:    result.StartTime.Format("2006/01/02 15:04:05"),
			EndTime:      result.EndTime.Format("2006/01/02 15:04:05"),
			JobID:        int(job.ID),
		}

		if result.Err != nil {
			jobLog.Err = result.Err.Error()
			jobLog.Result = "0"
			statusTyp = common.StatusTyp["执行异常"]
			// TODO 发送邮件
			logger.Error.Println(result.ExecuteInfo.Job.Name, ": 任务执行异常！")
		} else {
			jobLog.Err = ""
			jobLog.Result = "1"
			if result.ExecuteInfo.Job.Typ == 0 {
				// 定时任务
				statusTyp = common.StatusTyp["待执行"]
			} else {
				// 单次任务
				statusTyp = common.StatusTyp["已完成"]
			}
			logger.Info.Println(result.ExecuteInfo.Job.Name, ": 任务执行成功！")
		}
		GStatusMgr.pushStatusEvent(common.BuildStatusEvent(statusTyp, result.ExecuteInfo.Job, result.ExecuteInfo.PlanTime, true), result.ExecuteInfo.Job.Typ)
		GLogMgr.Append(jobLog)
	}
}

// 调度协程
func (_self *Scheduler) scheduleLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult     *common.JobExecuteResult
	)

	// 初始化一次(1秒)
	scheduleAfter = _self.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	// 定时任务common.Job
	for {
		select {
		case jobEvent = <-_self.jobEventChan: //监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			_self.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 最近的任务到期了
		case jobResult = <-_self.jobResultChan: // 监听任务执行结果
			_self.handleJobResult(jobResult)
		}
		// 调度一次任务
		scheduleAfter = _self.TrySchedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

// InitScheduler 初始化调度器
func InitScheduler() (err error) {
	GScheduler = &Scheduler{
		// 用来接收 http 推送过来的 job
		jobEventChan: make(chan *common.JobEvent, 1000),
		// 监听到任务变化时，将任务同步到执行计划表中
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
		// 将开始执行的任务放入执行表中
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		// 接收任务执行完成后的输出等信息
		jobResultChan: make(chan *common.JobExecuteResult, 1000),
	}

	// 启动调度协程
	go GScheduler.scheduleLoop()
	return
}
