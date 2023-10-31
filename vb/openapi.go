package vb

import (
	"github.com/lgdzz/vingo-backstage/vb/model"
	"github.com/lgdzz/vingo-backstage/vb/service"
	"github.com/lgdzz/vingo-utils/vingo"
)

// 搜索组织内的账户信息
func SearchOrgAccount(c *vingo.Context) {
	var query = vingo.GetRequestQuery[model.AccountQuery](c)
	query.OrgID = c.GetOrgId()
	c.ResponseBody(service.AccountList(&query))
}
