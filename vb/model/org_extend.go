package model

import "github.com/lgdzz/vingo-utils/vingo"

// OrgExtend 组织扩展
type OrgExtend struct {
	OrgID      uint            `gorm:"primaryKey;column:org_id" json:"orgId"` // 组织ID
	MaxMember  uint            `gorm:"column:max_member" json:"maxMember"`    // 最大成员数
	ExpireDate vingo.LocalTime `gorm:"column:expire_date" json:"expireDate"`  // 到期时间
}

// TableName get sql table name.获取数据库表名
func (m *OrgExtend) TableName() string {
	return "org_extend"
}
