package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"net"
	"strings"
	"time"

	"crontab/worker/logger"
)

// UnpackJob 反序列化 Job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)

	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

// ExtractJobName 从etcd的key中提取任务名 /cron/jobs/job10抹掉/cron/jobs/
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JobSaveDir)
}

// ExtractKillerName 从 /cron/killer/job10提取job10
func ExtractKillerName(killerKey string) string {
	return strings.TrimPrefix(killerKey, JobKillerDir)
}

// BuildJobEvent 任务变化事件有2种：1）更新任务 2）删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

// BuildJobSchedulePlan 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		nextTime time.Time
		expr     *cronexpr.Expression
	)

	if job.CronExpr == "" && job.Num < 1 {
		// 单次任务，未执行过，设置为立即执行
		nextTime = time.Now()
	} else {
		// 解析JOB的cron表达式
		if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
			// 执行过后，还未及时删除的单次任务，可能会解析出错
			logger.Error.Printf("解析定时表达式失败: %s ", err)
			return
		}
		nextTime = expr.Next(time.Now())
	}

	// 生成任务调度计划对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: nextTime,
	}
	return
}

// BuildJobExecuteInfo 构造执行状态信息
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime, // 计划调度时间
		RealTime: time.Now(),               // 真实调度时间
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

// BuildStatusEvent 构造任务状态修改事件
func BuildStatusEvent(stsTyp int, job *Job, nextTime time.Time, adm bool) (statusEve *JobStatusEvent) {
	return &JobStatusEvent{
		StatusTyp: stsTyp,
		Job:       job,
		AddNum:    adm,
		NextTime:  nextTime,
	}
}

// GetLocalIP 获取本机网卡IP
func GetLocalIP() (ipv4 string, err error) {
	var (
		adders  []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet // IP地址
		isIpNet bool
	)
	// 获取所有网卡
	if adders, err = net.InterfaceAddrs(); err != nil {
		logger.Error.Printf("获取所有网卡失败: %s ", err)
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr = range adders {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return
			}
		}
	}
	err = ErrNoLocalIpFound
	return
}

// GetNumField 更新任务状态时，对执行过的任务的执行次数加 1
func GetNumField(adm bool) int {
	if adm {
		return 1
	}
	return 0
}
