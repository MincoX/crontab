package core

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"

	"crontab/worker/common"
	"crontab/worker/logger"
)

var (
	GRegister *Register
)

// Register 注册节点到etcd： /cron/workers/IP地址
type Register struct {
	client       *clientv3.Client
	kv           clientv3.KV
	lease        clientv3.Lease
	registerTime time.Time
	localIP      string // 本机IP
}

// 注册到/cron/workers/IP, 并自动续租
func (_self *Register) keepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		// 注册路径
		regKey = common.JobWorkerDir + _self.localIP

		cancelFunc = nil

		// 创建租约
		if leaseGrantResp, err = _self.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		// 自动续租
		if keepAliveChan, err = _self.lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到etcd
		if _, err = _self.kv.Put(cancelCtx, regKey, time.Now().Format("2006/01/02 15:04:05"), clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		// 处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(3 * time.Second)
		// cancelFunc != nil 说明 key 已经和租约 id 绑定了，此时自动续租失败，需要重新创建租约，取消 key 与之前租约 id 的绑定
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

func InitRegister() (err error) {
	var (
		curTime time.Time
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIp string
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   common.GConfig.Etcd.Endpoints,                                // 集群地址
		DialTimeout: time.Duration(common.GConfig.Etcd.DialTimeout) * time.Second, // 连接超时
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		logger.Error.Printf("etcd 连接建立失败: %s", err)
		return
	}

	curTime = time.Now()

	// 本机IP
	if localIp, err = common.GetLocalIP(); err != nil {
		logger.Error.Printf("获取本机 ip 失败: %s", err)
		return
	} else {
		logger.Info.Printf("worker 上线， ip：%s, time：%s", localIp, curTime)
	}

	// 得到 KV 和 Lease 的 API 子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GRegister = &Register{
		client:       client,
		kv:           kv,
		lease:        lease,
		registerTime: curTime,
		localIP:      localIp,
	}

	// 服务注册，并自动续约；当服务宕机，会停止自动续约，一段时间后 key 就自动过期了（worker 下线）
	go GRegister.keepOnline()
	return
}
