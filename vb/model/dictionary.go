package model

import (
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"gorm.io/gorm"
)

// Dictionary 数据字典
type Dictionary struct {
	ID          uint   `gorm:"primaryKey;column:id" json:"id"`
	Pid         uint   `gorm:"column:pid" json:"pid"`
	Name        string `gorm:"column:name" json:"name"`               // 字典索引
	Description string `gorm:"column:description" json:"description"` // 字典名称
	Value       string `gorm:"column:value" json:"value"`             // 值
	Path        string `gorm:"column:path" json:"path"`
	Sort        uint8  `gorm:"column:sort" json:"sort"` // 排序
	Len         uint   `gorm:"column:len" json:"len"`
	OrgID       uint   `gorm:"column:org_id" json:"orgId"`
}

type DictionarySimple struct {
	ID          uint                `json:"id"`
	Description string              `json:"description"`
	Name        string              `json:"name"`
	Value       string              `json:"value"`
	HasChild    bool                `json:"hasChild"`
	Children    []*DictionarySimple `json:"children"`
}

// TableName get sql table name.获取数据库表名
func (m *Dictionary) TableName() string {
	return "dictionary"
}

func (m *Dictionary) AfterSave(tx *gorm.DB) (err error) {
	go RefreshDictionary()
	return nil
}

func (m *Dictionary) AfterDelete(tx *gorm.DB) (err error) {
	go RefreshDictionary()
	return nil
}

var DictionaryMap map[uint]string

// 刷新字典缓存
func RefreshDictionary() {
	var rows []Dictionary
	mysql.Db.Find(&rows)

	var dictionaryMap = map[uint]string{}
	for _, row := range rows {
		dictionaryMap[row.ID] = row.Value
	}

	DictionaryMap = dictionaryMap
}

// 通过字典ID获取字典值
func GetDictionaryValue(id uint) string {
	if value, ok := DictionaryMap[id]; ok {
		return value
	} else {
		return "-"
	}
}
