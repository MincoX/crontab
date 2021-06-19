package common

import (
	"context"
	"time"

	"github.com/gorhill/cronexpr"

	"crontab/worker/model"
)

type UserDto struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Telephone string    `json:"telephone"`
	CreatedAt time.Time `json:"created_at"`
}

// Job 定时任务
type Job struct {
	Name     string `json:"name"`     //  任务名
	Command  string `json:"command"`  // shell命令
	CronExpr string `json:"cronExpr"` // cron表达式
	Typ      int    `json:"typ"`      // 任务类型
	Num      int    `json:"num"`      // 执行次数
}

// JobEvent 变化事件
type JobEvent struct {
	EventType int //  SAVE, DELETE
	Job       *Job
}

// JobSchedulePlan 任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的 cronexpr 表达式
	NextTime time.Time            // 下次调度时间
}

// JobExecuteInfo 任务执行状态
type JobExecuteInfo struct {
	Job        *Job               // 任务信息
	PlanTime   time.Time          // 理论上的调度时间
	RealTime   time.Time          // 实际的调度时间
	CancelCtx  context.Context    // 任务command的context
	CancelFunc context.CancelFunc //  用于取消command执行的cancel函数
}

// JobExecuteResult 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo // 执行状态
	Output      []byte          // 脚本输出
	Err         error           // 脚本错误原因
	StartTime   time.Time       // 启动时间
	EndTime     time.Time       // 结束时间
}

// JobStatusEvent 任务状态更新
type JobStatusEvent struct {
	StatusTyp int // 任务状态类型
	Job       *Job
	AddNum    bool
	NextTime  time.Time
}

// LogBatch 日志批次
type LogBatch struct {
	Logs []*model.Log // 多条日志
}

// JobLogFilter 任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

// SortLogByStartTime 任务日志排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"` // {startTime: -1}
}

// Response HTTP接口应答
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}
