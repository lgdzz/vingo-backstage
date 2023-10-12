package model

import (
	"fmt"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"gorm.io/gorm"
)

// Dept 部门
type Dept struct {
	ID        uint             `gorm:"primaryKey;column:id" json:"id"`
	OrgID     uint             `gorm:"primaryKey;column:org_id" json:"orgId"` // 组织ID
	Pid       uint             `gorm:"column:pid" json:"pid"`                 // 上级部门ID
	Path      string           `gorm:"column:path" json:"path"`               // 部门结构
	Len       uint8            `gorm:"column:len" json:"len"`                 // 部门长度
	Channle   uint8            `gorm:"column:channel" json:"channel"`         // 数据来源，1-同步|2-录入
	Name      string           `gorm:"column:name" json:"name"`               // 部门名称
	Code      string           `gorm:"column:code" json:"code"`               // 部门编码
	Sort      int64            `gorm:"column:sort" json:"sort"`               // 排序，升序
	Status    uint8            `gorm:"column:status" json:"status"`           // 1-启用|2禁用
	CreatedAt *vingo.LocalTime `gorm:"column:created_at;" json:"createdAt"`
	UpdatedAt *vingo.LocalTime `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt   `gorm:"column:deleted_at" json:"deletedAt"`
}

// TableName get sql table name.获取数据库表名
func (m *Dept) TableName() string {
	return "dept"
}

type DeptQuery struct {
	OrgID   uint   `form:"orgId"`
	Status  uint8  `form:"status"`
	Keyword string `form:"keyword"`
}

type DeptSimple struct {
	ID       uint          `json:"id"`
	Name     string        `json:"name"`
	Code     string        `json:"code"`
	OrgID    uint          `json:"orgId"`
	HasChild bool          `json:"hasChild"`
	Children []*DeptSimple `json:"children"`
}

func (s *Dept) Parent() Dept {
	var dept Dept
	mysql.NotExistsErr(&dept, "id=?", s.Pid)
	return dept
}

func (s *Dept) SetPath(tx *gorm.DB) {
	if s.Pid > 0 {
		parent := s.Parent()
		s.Path = fmt.Sprintf("%v,%d", parent.Path, s.ID)
		s.Len = parent.Len + 1
	} else {
		s.Path = fmt.Sprintf("%d", s.ID)
		s.Len = 1
	}
	tx.Select("path", "len").Updates(&s)
}

func (s *Dept) SetPathChild(tx *gorm.DB) {
	var list []Dept
	mysql.Db.Find(&list, "pid=?", s.ID)
	for _, dept := range list {
		dept.SetPath(tx)
		dept.SetPathChild(tx)
	}
}

// dept.id,dept.pid,dept.name,dept.code,dept.status,dept.org_id AS orgId,DATE_FORMAT(dept.created_at,'%Y-%m-%d') AS joinTime,(select count(*) from `dept_member` where dept.id=dept_member.dept_id) AS people
type DeptTreeItem struct {
	Id        uint   `gorm:"column:id" json:"id"`
	Pid       uint   `gorm:"column:pid" json:"pid"`
	Name      string `gorm:"column:name" json:"name"`
	Code      string `gorm:"column:code" json:"code"`
	Status    uint   `gorm:"column:code" json:"status"`
	OrgId     uint   `gorm:"column:org_id" json:"orgId"`
	MemberNum int64  `gorm:"column:memebr_num" json:"memberNum"`
}
