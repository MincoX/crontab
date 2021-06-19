package common

var (
	JobType = map[string]int{
		"定时任务": 0,
		"单次任务": 1,
	}

	StatusTyp = map[string]int{
		"待调度":  0, // 任务在 etcd 中，还没有通过 watch 放入计划表
		"待执行":  1, // 任务已同步至计划表
		"执行中":  2, // 任务在执行队列中
		"执行异常": 3, // 任务被强杀，或者执行出错
		"已完成":  4, // 任务成功从执行队列中删除（只对单次任务有效，定时任务执行完成后状态成待执行）
		"已删除":  5, // 任务从 etcd 中删除
	}
)

const (
	// JobSaveDir 任务保存目录
	JobSaveDir = "/cron/jobs/"

	// JobKillerDir 任务强杀目录
	JobKillerDir = "/cron/killer/"

	// JobLockDir 任务锁目录
	JobLockDir = "/cron/lock/"

	// JobWorkerDir 服务注册目录
	JobWorkerDir = "/cron/workers/"
)
