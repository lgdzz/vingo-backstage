package model

import (
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"github.com/lgdzz/vingo-utils/vingo/db/page"
	"gorm.io/gorm"
)

type Account struct {
	ID        uint             `gorm:"primaryKey;column:id" json:"id"`
	OrgID     uint             `gorm:"column:org_id" json:"orgId"`                                                   // 组织ID
	UserID    uint             `gorm:"column:user_id" json:"userId"`                                                 // 用户ID
	RoleID    uint             `gorm:"column:role_id" json:"roleId"`                                                 // 角色ID
	Status    int8             `gorm:"column:status;default:1" json:"status"`                                        // 1-启用|2-禁用
	Extends   any              `gorm:"column:extends;serializer:json" json:"extends"`                                // 附加属性，对象
	User      *User            `gorm:"joinForeignKey:user_id;foreignKey:id;references:UserID" json:"user,omitempty"` // 用户
	Role      *Role            `gorm:"joinForeignKey:role_id;foreignKey:id;references:RoleID" json:"role,omitempty"` // 角色
	Org       *Org             `gorm:"joinForeignKey:org_id;foreignKey:id;references:OrgID" json:"org,omitempty"`    // 组织架构
	CreatedAt *vingo.LocalTime `gorm:"column:created_at;" json:"createdAt"`
	UpdatedAt *vingo.LocalTime `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt   `gorm:"column:deleted_at" json:"deletedAt"`
}

func (m *Account) TableName() string {
	return "account"
}

type AccountQuery struct {
	page.Limit
	OrgID     uint    `form:"orgId"`
	UserID    uint    `form:"userId"`
	RoleID    uint    `form:"roleId"`
	Status    uint    `form:"status"`
	Keyword   string  `form:"keyword"`
	AccountID *[]uint `form:"accountId"`
}

type AccountList struct {
	AccountID    uint   `gorm:"column:id" json:"accountId"`
	OrgID        uint   `gorm:"column:org_id" json:"orgId"`
	OrgName      string `gorm:"column:org_name" json:"orgName"`
	OrgGradeName string `gorm:"column:org_grade_name" json:"orgGradeName"`
	OrgGradeID   uint   `gorm:"column:org_grade_id" json:"orgGradeId"`
	OrgPID       uint   `gorm:"column:org_pid" json:"orgPid"`
	UserID       uint   `gorm:"column:user_id" json:"userId"`
	Username     string `gorm:"column:username" json:"username"`
	Realname     string `gorm:"column:realname" json:"realname"`
	Phone        string `gorm:"column:phone" json:"phone"`
	RoleID       uint   `gorm:"column:role_id" json:"roleId"`
	RoleName     string `gorm:"column:role_name" json:"roleName"`
	FromID       uint   `gorm:"column:from_id" json:"fromId"`
	Status       int    `gorm:"column:status" json:"status"`
}

type AccountBody struct {
	UserRegisterBody
	OrgID  uint   `json:"orgId"`
	RoleID uint   `json:"roleId"`
	Status int8   `json:"status"`
	Type   string `json:"type"`
}

// 获取账户账号
func (s *Account) GetUser() (user User) {
	mysql.NotExistsErr(&user, s.UserID)
	return
}

// 获取账户角色
func (s *Account) GetRole() (role Role) {
	mysql.NotExistsErr(&role, s.RoleID)
	return
}

// 获取账户组织
func (s *Account) GetOrg() (org Org) {
	mysql.NotExistsErr(&org, s.OrgID)
	return
}
