package common

import (
	"crontab/master/model"
	"strings"
)

// ToUserDto 登录用户的响应信息
func ToUserDto(user model.User) UserDto {
	return UserDto{
		Username:  user.Username,
		Telephone: user.Telephone,
		CreatedAt: user.CreatedAt,
	}
}

// ExtractWorkerIP 提取worker的IP
func ExtractWorkerIP(regKey string) string {
	return strings.TrimPrefix(regKey, JobWorkerDir)
}
