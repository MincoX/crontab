package model

import "gorm.io/gorm"

type Log struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey" json:"id"`
	JobName      string `json:"job_name"`      // 任务名字
	Command      string `json:"command"`       // 脚本命令
	Output       string `json:"output"`        // 命令输出
	Err          string `json:"err" `          // 错误输出
	PlanTime     string `json:"plan_time"`     // 计划开始时间
	ScheduleTime string `json:"schedule_time"` // 实际调度时间
	StartTime    string `json:"start_time"`    // 任务执行开始时间
	EndTime      string `json:"end_time"`      // 任务执行结束时间
	Result       string `json:"result"`        // 任务执行结果，根据是否有错误输出进行标记；0 表示执行出错；1 表示执行成功
	JobID        int    `json:"job_id"`        // 默认外键，任务 id
}
