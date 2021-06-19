package common

import "errors"

var (
	ErrNoLocalIpFound      = errors.New("没有找到网卡IP")
)
