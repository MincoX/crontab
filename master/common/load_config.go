package common

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var (
	GConfig *Config
)

type Config struct {
	Http    Http
	Etcd    Etcd
	Redis   Redis
	MySQL   MySQL
	MongoDB MongoDB
	Worker  Worker
}

type Http struct {
	Addr        string `yaml:"addr"`
	LoginExpire int    `yaml:"login_expire"`
}

type Etcd struct {
	DialTimeout int      `yaml:"dial_timeout"`
	Endpoints   []string `yaml:"endpoints"`
}

type Redis struct {
	Addr         string `json:"addr"`
	Password     string `json:"password"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	DB           int    `json:"db"`
}

type MySQL struct {
	DriverName  string `yaml:"driver_name"`
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	Database    string `yaml:"database"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	MaxIdleCons int    `yaml:"max_idle_cons"`
	MaxOpenCons int    `yaml:"max_open_cons"`
	MaxLifetime int    `yaml:"max_lifetime"`
	Charset     string `yaml:"charset"`
	Loc         string `yaml:"loc"`
}

type MongoDB struct {
	Addr           string `yaml:"addr"`
	ConnectTimeout int    `yaml:"connect_timeout"`
	DB             string `yaml:"db"`
	AuthDB         string `yaml:"auth_db"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	PoolLimit      int    `yaml:"pool_limit"`
}

type Worker struct {
	ScheduleSleep    int    `yaml:"schedule_sleep"`
	BashPath         string `yaml:"bash_path"`
	LogBatchSize     int    `yaml:"log_batch_size"`
	LogCommitTimeout int    `yaml:"log_commit_timeout"`
}

// InitConfig 加载配置
func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)

	// 1, 把配置文件读进来
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	// 2, 做JSON反序列化
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return
	}

	// 3, 赋值单例
	GConfig = &conf

	return
}
