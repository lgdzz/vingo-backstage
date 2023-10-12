package model

// LoginLog 登录日志
type LoginLog struct {
	ID        uint   `gorm:"primaryKey;column:id" json:"-"`
	UserID    uint   `gorm:"column:user_id" json:"userId"`
	IPIsp     string `gorm:"column:ip_isp" json:"ipIsp"`
	LoginIP   string `gorm:"column:login_ip" json:"loginIp"`
	LoginTime string `gorm:"column:login_time" json:"loginTime"`
	Channel   string `gorm:"column:channel" json:"channel"`
}

// TableName get sql table name.获取数据库表名
func (m *LoginLog) TableName() string {
	return "login_log"
}
