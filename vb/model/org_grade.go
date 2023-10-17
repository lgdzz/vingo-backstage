package model

import (
	"github.com/lgdzz/vingo-utils/vingo"
	"gorm.io/gorm"
)

// OrgGrade 组织类型
type OrgGrade struct {
	Id          uint             `gorm:"primaryKey;column:id" json:"id"`
	Pid         uint             `gorm:"column:pid" json:"pid"`
	Path        string           `gorm:"column:path" json:"path"`
	Len         uint             `gorm:"column:len" json:"len"`
	Code        string           `gorm:"column:code" json:"code"`                       // 组织类型编码
	Name        string           `gorm:"column:name" json:"name"`                       // 组织类型名称
	NameEn      string           `gorm:"column:name_en" json:"nameEn"`                  // 英文名称
	Description string           `gorm:"column:description" json:"description"`         // 描述
	Extends     OrgGradeExtends  `gorm:"column:extends;serializer:json" json:"extends"` // 扩展对象
	Status      uint             `gorm:"column:status" json:"status"`                   // 1-启用|2-禁用
	Sort        uint             `gorm:"column:sort" json:"sort"`                       // 排序，升序
	AdminRoleID uint             `gorm:"column:admin_role_id" json:"adminRoleId"`
	CreatedAt   *vingo.LocalTime `gorm:"column:created_at;" json:"createdAt"`
	UpdatedAt   *vingo.LocalTime `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt   `gorm:"column:deleted_at" json:"deletedAt"`
}

// TableName get sql table name.获取数据库表名
func (m *OrgGrade) TableName() string {
	return "org_grade"
}

type OrgGradeExtends struct {
	ShareRoleIds []uint `json:"shareRoleIds"`
}

type OrgGradeTreeItem struct {
	Id          uint            `gorm:"column:id" json:"id"`
	Pid         uint            `gorm:"column:pid" json:"pid"`
	Name        string          `gorm:"column:name" json:"name"`
	NameEn      string          `gorm:"column:name_en" json:"nameEn"`
	Description string          `gorm:"column:description" json:"description"`
	Extends     OrgGradeExtends `gorm:"column:extends;serializer:json" json:"extends"`
	Status      uint            `gorm:"column:status" json:"status"`
	Sort        uint            `gorm:"column:sort" json:"sort"`
	AdminRoleId uint            `gorm:"column:admin_role_id" json:"adminRoleId"`
	RoleName    string          `gorm:"column:role_name" json:"roleName"`
	OrgCount    int64           `gorm:"column:org_count" json:"orgCount"`
}
