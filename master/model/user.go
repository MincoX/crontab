package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey" json:"id"`
	Username  string `gorm:"type:varchar(20);not null" json:"username"`
	Password  string `gorm:"type:varchar(120);not null" json:"password"`
	Email     string `gorm:"type:varchar(32);" json:"email"`
	Telephone string `gorm:"size:16;" json:"telephone"`
	ImageUrl  string `gorm:"type:varchar(32);" json:"image_url"`
	Jobs      []Job  // 一对多关联属性，表示多个任务
}
