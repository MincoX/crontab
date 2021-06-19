package core

import (
	"math/rand"
	"os/exec"
	"time"

	"crontab/worker/common"
)

var (
	GExecutor *Executor
)

// Executor 任务执行器
type Executor struct {
}

// ExecuteJob 执行一个任务
func (_self *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			err     error
			output  []byte
			jobLock *JobLock
			cmd     *exec.Cmd
			result  *common.JobExecuteResult
		)

		// 任务结果
		result = &common.JobExecuteResult{
			ExecuteInfo: info,            // 执行任务
			Output:      make([]byte, 0), // 任务的输出
		}

		// 初始化分布式锁
		jobLock = GJobMgr.CreateJobLock(info.Job.Name)

		// 记录任务开始时间
		result.StartTime = time.Now()

		// 上锁 随机睡眠(0~1s)，避免因为 cpu 时间分片导致同一台机器上多个节点只会有一个节点一直抢到锁，其他节点一直抢不到锁
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		err = jobLock.TryLock()
		defer jobLock.Unlock()

		if err != nil { // 上锁失败
			result.Err = err
			result.EndTime = time.Now()
		} else {
			// 上锁成功后，重置任务启动时间
			result.StartTime = time.Now()

			// 执行shell命令
			cmd = exec.CommandContext(info.CancelCtx, common.GConfig.Worker.BashPath, "-c", info.Job.Command)

			// 执行并捕获输出
			output, err = cmd.CombinedOutput()

			// 记录任务结束时间
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}
		// 任务执行完成后，把执行的结果返回给Scheduler，Scheduler会从executingTable中删除掉执行记录
		GScheduler.PushJobResult(result)
	}()
}

// InitExecutor 初始化执行器
func InitExecutor() (err error) {
	GExecutor = &Executor{}
	return
}
