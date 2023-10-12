package model

import (
	"fmt"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"gorm.io/gorm"
)

// Role 角色
type Role struct {
	vingo.DbModel
	Pid       uint          `gorm:"column:pid" json:"pid"`
	OrgID     uint          `gorm:"column:org_id" json:"orgId"`
	Path      string        `gorm:"column:path" json:"path"`
	Name      string        `gorm:"column:name" json:"name"`
	Master    int8          `gorm:"column:master" json:"master"`
	Status    int8          `gorm:"column:status" json:"status"`                   // 1-启用|2-禁用
	IsSystem  int8          `gorm:"column:is_system" json:"isSystem"`              // 1-系统默认
	Remark    string        `gorm:"column:remark" json:"remark"`                   // 角色描述
	Extends   any           `gorm:"column:extends;serializer:json" json:"extends"` // 扩展对象
	Rules     vingo.UintIds `gorm:"column:rules" json:"rules"`                     // 所有权限
	HalfRules vingo.UintIds `gorm:"column:half_rules" json:"halfRules"`            // 能授权不能使用
	AllowEdit bool          `json:"allowEdit"`
}

// TableName get sql table name.获取数据库表名
func (m *Role) TableName() string {
	return "role"
}

func (s *Role) Parent() Role {
	var role Role
	mysql.NotExistsErr(&role, "id=?", s.Pid)
	return role
}

func (s *Role) SetPath(tx *gorm.DB) {
	if s.Pid > 0 {
		s.Path = fmt.Sprintf("%v,%d", s.Parent().Path, s.ID)
	} else {
		s.Path = fmt.Sprintf("%d", s.ID)
	}
	tx.Select("path").Updates(&s)
}

// 验证是否被账户使用
func (s *Role) HasAccount() bool {
	return mysql.Exists(&Account{}, "role_id=?", s.ID)
}
