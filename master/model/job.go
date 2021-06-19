package model

import (
	"time"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	ID       uint       `gorm:"primaryKey" json:"id"`
	Name     string     `gorm:"type:varchar(20);not null" json:"name"`      //  任务名
	Command  string     `gorm:"type:varchar(255);not null" json:"command"`  // shell命令
	CronExpr string     `gorm:"type:varchar(20);not null" json:"cron_expr"` // cron表达式
	Status   int        `json:"status"`                                     // 执行状态
	NextTime *time.Time `json:"next_time"`                                  // 下次调度时间
	Typ      int        `json:"typ"`                                        // 任务类型(0: 定时任务；1: 单次任务)
	Num      int        `json:"num"`                                        // 执行次数
	UserID   int        `json:"user_id"`                                    // 默认外键，用户 id
	Logs     []Log      // 一对多关联属性，表示多条日志
}
