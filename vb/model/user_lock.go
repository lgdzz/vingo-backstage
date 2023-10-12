package model

import (
	"time"
)

// UserLock 用户锁
type UserLock struct {
	Username   string     `gorm:"primaryKey;column:username" json:"username"` // 用户名
	Bad        uint8      `gorm:"column:bad" json:"bad"`                      // 密码错误次数
	LockTime   *time.Time `gorm:"column:lock_time" json:"lockTime"`           // 锁定时间
	UnlockTime *time.Time `gorm:"column:unlock_time" json:"unlockTime"`       // 解锁时间
	Duration   uint       `gorm:"column:duration" json:"duration"`            // 时长，秒
}

// TableName get sql table name.获取数据库表名
func (m *UserLock) TableName() string {
	return "user_lock"
}
