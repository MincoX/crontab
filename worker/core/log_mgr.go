package core

import (
	"time"

	"crontab/worker/common"
	"crontab/worker/logger"
	"crontab/worker/model"
)

var (
	GLogMgr *LogMgr
)

// LogMgr mongodb存储日志
type LogMgr struct {
	logChan        chan *model.Log
	autoCommitChan chan *common.LogBatch
}

// Append 发送日志
func (_self *LogMgr) Append(jobLog *model.Log) {
	select {
	case _self.logChan <- jobLog:
	default:
		// 队列满了就丢弃
	}
}

// 日志存储协程
func (_self *LogMgr) writeLoop() {
	var (
		err          error
		log          *model.Log
		logBatch     *common.LogBatch // 当前的批次
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch // 超时批次
	)

	for {
		select {
		case log = <-_self.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				// 让这个批次超时自动提交，对每一个批次只创建一个定时器，
				// 定时器放在 if 语句内，每一个 batch 对应一个定时器。若是在 if 外面，则是对每一个 log 都创建一个定时器
				commitTimer = time.AfterFunc(
					time.Duration(common.GConfig.Worker.LogCommitTimeout)*time.Second,
					func(batch *common.LogBatch) func() {
						return func() {
							_self.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}

			// 把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			// 如果批次满了, 就立即发送
			if len(logBatch.Logs) >= common.GConfig.Worker.LogBatchSize {
				// 批量插入数据库
				if err = common.GMsql.DB.Model(&model.Log{}).Create(&timeoutBatch.Logs).Error; err != nil {
					logger.Error.Printf("批量插入日志失败：%s ", err)
				}
				//_ = common.GMgo.InsertMany("log", logBatch.Logs)
				// 清空logBatch
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}

		case timeoutBatch = <-_self.autoCommitChan: // 过期的批次
			// 判断过期批次是否仍旧是当前的批次，避免超时的同时批次满了，已经自动提交
			if timeoutBatch != logBatch {
				continue // 跳过已经被提交的批次
			}
			// 批量插入数据库
			if err = common.GMsql.DB.Model(&model.Log{}).Create(&timeoutBatch.Logs).Error; err != nil {
				logger.Error.Printf("批量插入日志失败：%s ", err)
			}
			//_ = common.GMgo.InsertMany("log", timeoutBatch.Logs)
			// 清空logBatch
			logBatch = nil
		}
	}
}

func InitLogMgr() (err error) {

	GLogMgr = &LogMgr{
		logChan:        make(chan *model.Log, 2000),      // 存放的是每一个日志
		autoCommitChan: make(chan *common.LogBatch, 200), // 存放的是每 batch 日志
	}

	// 启动一个mongodb处理协程
	go GLogMgr.writeLoop()
	return
}
