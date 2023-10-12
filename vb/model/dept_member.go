package model

import "github.com/lgdzz/vingo-utils/vingo"

// DeptMember 部门成员
type DeptMember struct {
	DeptID    uint             `gorm:"primaryKey;column:dept_id" json:"deptId"`       // 部门ID
	AccountID uint             `gorm:"primaryKey;column:account_id" json:"accountId"` // 组织用户ID
	OrgID     uint             `gorm:"column:org_id" json:"orgId"`
	IsLeader  uint8            `gorm:"column:is_leader" json:"isLeader"` // 是否是领导
	CreatedAt *vingo.LocalTime `gorm:"column:created_at" json:"createdAt"`
}

// TableName get sql table name.获取数据库表名
func (m *DeptMember) TableName() string {
	return "dept_member"
}

type DeptMemberItem struct {
	DeptID    uint `json:"deptId"`
	AccountID uint `json:"accountId"`
}
