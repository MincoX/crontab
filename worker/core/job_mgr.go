package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"crontab/worker/common"
	"crontab/worker/logger"
)

var (
	GJobMgr *JobMgr
)

// JobMgr 任务管理器
type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

// 监听任务变化
func (_self *JobMgr) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		keypair            *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
	)

	// 1, get一下/cron/jobs/目录下的所有任务，并且获知当前集群的revision
	if getResp, err = _self.kv.Get(context.TODO(), common.JobSaveDir, clientv3.WithPrefix()); err != nil {
		logger.Error.Printf("读取 etcd 中任务失败: %s ", err)
		return
	}

	logger.Debug.Printf("etcd 中共 %s 个任务待同步 ", len(getResp.Kvs))
	// 当前有哪些任务
	for _, keypair = range getResp.Kvs {
		// 反序列化 json 得到 Job
		if job, err = common.UnpackJob(keypair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JobEventSave, job)
			// 同步给scheduler(调度协程)
			GScheduler.PushJobEvent(jobEvent)
		}
	}

	// 2, 从该revision向后监听变化事件
	go func() {
		// 从 GET 时刻的后续版本开始监听变化
		watchStartRevision = getResp.Header.Revision + 1
		// 监听/cron/jobs/目录的后续变化
		watchChan = _self.watcher.Watch(context.TODO(), common.JobSaveDir, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 任务保存事件
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构建一个更新Event
					jobEvent = common.BuildJobEvent(common.JobEventSave, job)
				case mvccpb.DELETE: // 任务被删除了
					// Delete /cron/jobs/job10
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					// 构建一个删除Event
					jobEvent = common.BuildJobEvent(common.JobEventDelete, job)
				}
				// 变化推给scheduler
				GScheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

// 监听强杀任务通知
func (_self *JobMgr) watchKiller() {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *common.JobEvent
		jobName    string
		job        *common.Job
	)

	// 监听/cron/killer目录
	go func() { // 监听协程
		// 监听/cron/killer/目录的变化
		watchChan = _self.watcher.Watch(context.TODO(), common.JobKillerDir, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 杀死任务事件
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JobEventKill, job)
					// 事件推给scheduler
					GScheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: // killer 标记过期, 被自动删除
					logger.Warn.Println("监听 kill， delete 类型事件")
				}
			}
		}
	}()
}

// DeleteJob 删除任务
func (_self *JobMgr) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	// etcd中保存任务的key
	jobKey = common.JobSaveDir + name

	// 从etcd中删除它
	if delResp, err = _self.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		logger.Error.Printf("刪除 etcd 中任务失败: %s ", err)
		return
	}

	// 返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		// 解析一下旧值, 返回它
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			logger.Error.Printf("刪除 etcd 中任务失败: %s ", err)
			return
		}
		oldJob = &oldJobObj
	}
	return
}

// CreateJobLock 创建任务执行锁
func (_self *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobName, _self.kv, _self.lease)
	return
}

// InitJobMgr 初始化管理器
func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   common.GConfig.Etcd.Endpoints,                                // 集群地址
		DialTimeout: time.Duration(common.GConfig.Etcd.DialTimeout) * time.Second, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		logger.Error.Printf("etcd 连接建立失败: %s ", err)
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	// 赋值单例
	GJobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	// 启动任务监听
	_ = GJobMgr.watchJobs()

	// 启动监听killer
	GJobMgr.watchKiller()

	return
}
