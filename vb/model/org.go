package model

import (
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

// Org 组织架构
type Org struct {
	vingo.DbModel
	Pid            uint      `gorm:"column:pid" json:"pid"`                                                           // 上级ID
	Pids           []uint    `gorm:"column:pids;serializer:json" json:"pids"`                                         // 上级ID集合
	Path           string    `gorm:"column:path" json:"path"`                                                         // 组织结构
	PathName       string    `gorm:"column:path_name" json:"pathName"`                                                // 组织结构名称
	Code           string    `gorm:"column:code" json:"code"`                                                         // 组织编码
	Name           string    `gorm:"column:name" json:"name"`                                                         // 组织名称
	NameEn         string    `gorm:"column:name_en" json:"nameEn"`                                                    // 英文名称
	FullName       string    `gorm:"column:full_name" json:"fullName"`                                                // 完整名称
	GradeID        uint      `gorm:"column:grade_id" json:"gradeId"`                                                  // 组织类型
	Len            uint8     `gorm:"column:len" json:"len"`                                                           // 组织长度
	Status         int8      `gorm:"column:status" json:"status"`                                                     // 1-启用|2禁用
	Sort           uint      `gorm:"column:sort" json:"sort"`                                                         // 排序，升序
	Description    string    `gorm:"column:description" json:"description"`                                           // 描述
	ContactName    string    `gorm:"column:contact_name" json:"contactName"`                                          // 组织联系人
	ContactTel     string    `gorm:"column:contact_tel" json:"contactTel"`                                            // 组织联系电话
	ContactAddress string    `gorm:"column:contact_address" json:"contactAddress"`                                    // 组织联系地址
	ProvinceCode   string    `gorm:"column:province_code" json:"provinceCode"`                                        // 省编码
	CityCode       string    `gorm:"column:city_code" json:"cityCode"`                                                // 市编码
	CountyCode     string    `gorm:"column:county_code" json:"countyCode"`                                            // 区县编码
	Province       string    `gorm:"column:province" json:"province"`                                                 // 省
	City           string    `gorm:"column:city" json:"city"`                                                         // 市
	County         string    `gorm:"column:county" json:"county"`                                                     // 区
	Grade          *OrgGrade `gorm:"joinForeignKey:grade_id;foreignKey:id;references:GradeID" json:"grade,omitempty"` // 组织类型
}

// TableName get sql table name.获取数据库表名
func (m *Org) TableName() string {
	return "org"
}

type OrgQuery struct {
	RootID  uint   `form:"rootId"`
	GradeID uint   `form:"gradeId"`
	Keyword string `form:"keyword"`
}

type OrgTree struct {
	ID        uint   `json:"id"`
	Pid       uint   `json:"pid"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	FullName  string `json:"fullName"`
	GradeId   string `json:"gradeId"`
	GradeName string `json:"gradeName"`
}

func (s *Org) Parent() Org {
	var org Org
	mysql.NotExistsErr(&org, "id=?", s.Pid)
	return org
}

func (s *Org) SetPath(tx *gorm.DB) {
	if s.Pid > 0 {
		parent := s.Parent()
		path := vingo.SliceStringToUint(strings.Split(parent.Path, ","))
		path = append(path, s.ID)
		pathName := strings.Split(parent.PathName, ",")
		pathName = append(pathName, s.Name)

		s.Path = strings.Join(vingo.SliceUintToString(path), ",")
		s.PathName = strings.Join(pathName, ",")
		s.Pids = append(parent.Pids, parent.ID)
		s.Len = parent.Len + 1
	} else {
		s.Path = strconv.Itoa(int(s.ID))
		s.PathName = s.Name
		s.Pids = []uint{}
		s.Len = 1
	}
	tx.Select("pids", "path", "path_name", "len").Updates(&s)
}

func (s *Org) GetExtend() (extend OrgExtend) {
	mysql.NotExistsErr(&extend, "org_id=?", s.ID)
	return
}

// todo 优化建议：可以缓存提高效率
func GetOrgName(id uint) (name string) {
	mysql.Model(&Org{}).Where("id=?", id).Select("name").Scan(&name)
	return
}

type OrgTreeItem struct {
	Id             uint   `gorm:"column:id" json:"id"`
	Pid            uint   `gorm:"column:pid" json:"pid"`
	Name           string `gorm:"column:name" json:"name"`
	FullName       string `gorm:"column:full_name" json:"fullName"`
	Code           string `gorm:"column:code" json:"code"`
	ContactName    string `gorm:"column:contact_name" json:"contactName"`
	ContactTel     string `gorm:"column:contact_tel" json:"contactTel"`
	ContactAddress string `gorm:"column:contact_address" json:"contactAddress"`
	Province       string `gorm:"column:province" json:"province"`
	City           string `gorm:"column:city" json:"city"`
	County         string `gorm:"column:county" json:"county"`
	GradeId        uint   `gorm:"column:grade_id" json:"gradeId"`
	GradeName      string `gorm:"column:grade_name" json:"gradeName"`
	AdminRoleName  string `gorm:"column:admin_role_name" json:"adminRoleName"`
}
