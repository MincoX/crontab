package service

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"crontab/master/common"
)

var (
	GWorkerSer *WorkerSer
)

// WorkerSer /cron/workers/
type WorkerSer struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

// ListWorkers 获取在线worker列表
func (_self *WorkerSer) ListWorkers() (workerArr []map[string]string, err error) {
	var (
		getResp *clientv3.GetResponse
		kv      *mvccpb.KeyValue
		ip      string
	)

	// 获取目录下所有Kv
	if getResp, err = _self.kv.Get(context.TODO(), common.JobWorkerDir, clientv3.WithPrefix()); err != nil {
		return
	}

	// 解析每个节点的IP
	for _, kv = range getResp.Kvs {
		// kv.Key : /cron/workers/192.168.2.1
		ip = common.ExtractWorkerIP(string(kv.Key))
		worker := map[string]string{"ip": ip, "activeTime": string(kv.Value)}
		workerArr = append(workerArr, worker)
	}
	return
}

func InitWorkerSer() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   common.GConfig.Etcd.Endpoints,                                // 集群地址
		DialTimeout: time.Duration(common.GConfig.Etcd.DialTimeout) * time.Second, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GWorkerSer = &WorkerSer{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}
