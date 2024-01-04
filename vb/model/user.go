package model

import (
	"fmt"
	"github.com/lgdzz/vingo-backstage/vb/index"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"github.com/lgdzz/vingo-utils/vingo/db/page"
	"time"
)

// User 用户
type User struct {
	vingo.DbModel
	Type        string `gorm:"column:type;default:branch" json:"type"` // 账号类型
	Phone       string `gorm:"column:phone" json:"phone"`              // 手机号
	Realname    string `gorm:"column:realname" json:"realname"`        // 真实姓名
	Username    string `gorm:"column:username" json:"username"`        // 用户名
	Password    string `gorm:"column:password" json:"password"`        // 密码
	Salt        string `gorm:"column:salt" json:"-"`
	Status      int8   `gorm:"column:status" json:"status"`                       // 1-启用|2-禁用
	Remark      string `gorm:"column:remark" json:"remark"`                       // 备注
	Avatar      string `gorm:"column:avatar" json:"avatar"`                       // 头像
	CompanyName string `gorm:"column:company_name" json:"companyName"`            // 所在公司
	CompanyJob  string `gorm:"column:company_job" json:"companyJob"`              // 所在公司职务
	LastIP      string `gorm:"column:last_ip" json:"lastIp"`                      // 最后登录IP
	LastTime    string `gorm:"column:last_time" json:"lastTime"`                  // 最后登录时间
	FromChannel string `gorm:"column:from_channel;default:组织" json:"fromChannel"` // 来源渠道
	FromID      uint   `gorm:"column:from_id" json:"fromId"`                      // 来源ID
	WxOpenid    string `gorm:"column:wx_openid" json:"wxOpenid"`                  // 绑定微信
}

// TableName get sql table name.获取数据库表名
func (m *User) TableName() string {
	return "user"
}

type UserLoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserChangePwdBody struct {
	Username    string
	OldPassword string
	NewPassword string
}

type UserChangeInfoBody struct {
	Avatar      string `json:"avatar"`
	Realname    string `json:"realname"`
	CompanyName string `json:"companyName"`
	CompanyJob  string `json:"companyJob"`
}

type UserSimple struct {
	ID          uint   `gorm:"column:id" json:"userId"`
	Username    string `gorm:"column:username" json:"username"`
	Realname    string `gorm:"column:realname" json:"realname"`
	Phone       string `gorm:"column:phone" json:"phone"`
	Avatar      string `gorm:"column:avatar" json:"avatar"`
	CompanyName string `gorm:"column:company_name" json:"companyName"`
	CompanyJob  string `gorm:"column:company_job" json:"companyJob"`
	Password    string `gorm:"column:password" json:"-"`
	Salt        string `gorm:"column:salt" json:"-"`
	Status      int8   `gorm:"column:status" json:"-"`
}

func (s *User) CheckPatchWhite(field string) {
	if !vingo.IsInSlice(field, []string{"realname", "status", "remark"}) {
		panic(fmt.Sprintf("字段%v禁止修改", field))
	}
}

func (s *User) WriteLoginLog(clientIP string) {
	defer func() {
		recover()
	}()
	loginTime := time.Now().Format(vingo.DatetimeFormat)
	mysql.Where("id=?", s.ID).Updates(&User{LastIP: clientIP, LastTime: loginTime})
	mysql.Create(&LoginLog{UserID: s.ID, LoginIP: clientIP, LoginTime: loginTime})
}

// 验证账号是否被锁定，如锁定则抛出错误
func (s *User) CheckLockStatus() {
	if index.Config.System.Auth.Lock.Enable {
		var userLock UserLock
		if mysql.Exists(&userLock, "username=?", s.Username) && userLock.UnlockTime != nil && userLock.UnlockTime.Unix() > time.Now().Unix() {
			panic(fmt.Sprintf("由于您的账号多次输入密码错误，已被系统锁定，解锁时间：%v", userLock.UnlockTime.Format(vingo.DatetimeFormat)))
		}
	}
}

// 登录时密码错误，bad+1
func (s *User) LockAddBad() {
	if !index.Config.System.Auth.Lock.Enable {
		return
	}
	userLock := UserLock{Username: s.Username}
	mysql.FirstOrCreate(&userLock)
	userLock.Bad++
	if userLock.Bad >= index.Config.System.Auth.Lock.Bad {
		now := time.Now()
		later := now.Add(time.Duration(index.Config.System.Auth.Lock.Time) * 60 * time.Second)
		userLock.LockTime = &now
		userLock.UnlockTime = &later
		userLock.Duration = index.Config.System.Auth.Lock.Time * 60
	}
	mysql.Save(&userLock)
}

// 用户解锁
func (s *User) Unlock() {
	if index.Config.System.Auth.Lock.Enable {
		mysql.Delete(&UserLock{}, "username=?", s.Username)
	}
}

type UserLoginResult struct {
	Token string     `json:"token"`
	User  UserSimple `json:"user"`
}

type UserRegisterBody struct {
	Username string `json:"username"`
	Realname string `json:"realname"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Remark   string `json:"remark"`
	FromID   uint   `json:"fromId"`
}

type UserQuery struct {
	page.Limit
	*page.Order
	Keyword string `form:"keyword"`
	Status  *uint  `form:"status"`
}

// 验证是否存在身份信息
func (s *User) HasAccount() bool {
	return mysql.Exists(&Account{}, "user_id=?", s.ID)
}
