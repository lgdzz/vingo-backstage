package weixin

import (
	"fmt"
	"github.com/lgdzz/vingo-backstage/vb/model"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"strconv"
)

// 同步部门
func SyncDepartment(c *vingo.Context) {
	orgId := c.GetOrgId()
	client := NewClient(Config{
		// todo
	})

	departments := client.GetDepartmentList()
	depts := vingo.ForEach(departments, func(item Department, index int) model.Dept {
		return model.Dept{
			ID:      uint(item.Id),
			Pid:     uint(item.ParentId),
			OrgID:   orgId,
			Name:    item.Name,
			Code:    "",
			Sort:    item.Order,
			Status:  0,
			Channle: 1,
		}
	})

	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	// 删除旧数据
	tx.Where("org_id=? AND channel=1", orgId).Unscoped().Delete(&model.Dept{})
	// 写入新数据
	tx.Create(&depts)
	for _, dept := range depts {
		if dept.Pid > 0 {
			var parent model.Dept
			mysql.TXNotExistsErr(tx, &parent, "id=? AND org_id=?", dept.Pid, dept.OrgID)
			dept.Path = fmt.Sprintf("%v,%d", parent.Path, dept.ID)
			dept.Len = parent.Len + 1
		} else {
			dept.Path = strconv.Itoa(int(dept.ID))
			dept.Len = 1
		}
		tx.Select("path", "len").Updates(&dept)
	}
	c.ResponseSuccess()
}
