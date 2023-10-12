package model

import (
	"fmt"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"gorm.io/gorm"
)

// Rule 权限规则
type Rule struct {
	ID               uint   `gorm:"primaryKey;column:id" json:"id"`
	Pid              uint   `gorm:"column:pid" json:"pid"`
	Path             string `gorm:"column:path" json:"path"`
	Name             string `gorm:"column:name" json:"name"`     // 规则名称
	Type             string `gorm:"column:type" json:"type"`     // 1-页面|2-操作
	Method           string `gorm:"column:method" json:"method"` // 接口请求方式
	PermissionID     string `gorm:"column:permission_id" json:"permissionId"`
	Operation        string `gorm:"column:operation" json:"operation"`                 // 操作标识
	ServiceRouter    string `gorm:"column:service_router" json:"serviceRouter"`        // 接口路由
	ClientRouter     string `gorm:"column:client_router" json:"clientRouter"`          // 客户端路由
	ClientRouteName  string `gorm:"column:client_route_name" json:"clientRouteName"`   // 客户端路由名称
	ClientRouteAlias string `gorm:"column:client_route_alias" json:"clientRouteAlias"` // 客户端路由别名
	Icon             string `gorm:"column:icon" json:"icon"`
	Sort             uint8  `gorm:"column:sort" json:"sort"`
}

// TableName get sql table name.获取数据库表名
func (m *Rule) TableName() string {
	return "rule"
}

type Permission struct {
	ID        string   `json:"id"`
	Operation []string `json:"operation"`
}

func (s *Rule) Parent() Rule {
	var rule Rule
	mysql.NotExistsErr(&rule, "id=?", s.Pid)
	return rule
}

func (s *Rule) SetPath(tx *gorm.DB) {
	if s.Pid > 0 {
		s.Path = fmt.Sprintf("%v,%d", s.Parent().Path, s.ID)
	} else {
		s.Path = fmt.Sprintf("%d", s.ID)
	}
	tx.Select("path").Updates(&s)
}
