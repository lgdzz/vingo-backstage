package vb

import (
	"encoding/json"
	"github.com/lgdzz/vingo-backstage/vb/model"
	"github.com/lgdzz/vingo-backstage/vb/service"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"time"
)

type LoginComponentMethod struct {
	Method map[string]any `json:"method"`
	ctx    *vingo.Context
}

type LoginComponentResult struct {
	Result map[string]any `json:"result"`
}

// 需登录的组件接口
// demo body
// {"method":{"CallTest":{"page":1,"size":10}}}
func LoginComponent(c *vingo.Context) {
	var body LoginComponentMethod
	c.RequestBody(&body)
	body.ctx = c

	result := LoginComponentResult{Result: make(map[string]any)}
	for k, v := range body.Method {
		tmp, _ := json.Marshal(&v)
		var param map[string]any
		_ = json.Unmarshal(tmp, &param)

		result.Result[k] = vingo.CallStructFunc(&body, k, param)
	}
	c.ResponseBody(result.Result)
}

// ======组件方法======

func (m *LoginComponentMethod) orgIdDefault(inputOrgId float64) uint {
	if uint(inputOrgId) > 0 {
		return uint(inputOrgId)
	} else {
		return m.ctx.GetUint("orgId")
	}
}

// 获取服务器时间
func (m *LoginComponentMethod) GetNowTime() string {
	return time.Now().Format(vingo.DatetimeFormat)
}

// 获取组织类型树
func (m *LoginComponentMethod) GetOrgGradeTree(gradeId float64) []map[string]any {
	return service.GetOrgGradeTree(uint(gradeId))
}

// 获取组织树
func (m *LoginComponentMethod) GetOrgTree(orgId float64) []map[string]any {
	return service.GetOrgTree(&model.OrgQuery{RootID: uint(orgId)})
}

// 获取部门树
func (m *LoginComponentMethod) GetDeptTree(orgId float64) []model.DeptSimple {
	var result []model.DeptSimple
	tree := service.GetOrgDeptTree(&model.DeptQuery{OrgID: uint(orgId), Status: 1})
	vingo.CustomOutput(&tree, &result)
	return result
}

// 获取角色树
func (m *LoginComponentMethod) GetRoleTree(orgId float64) []map[string]any {
	return service.GetRoleTreeByOrgId(uint(orgId), 1, m.ctx)
}

// 获取角色权限树
func (m *LoginComponentMethod) GetRoleRuleTree(roleId float64) []map[string]any {
	return service.GetRuleTreeByRoleId(uint(roleId), "")
}

// 获取字典(系统字典+本组织字典)
func (m *LoginComponentMethod) GetDictionary() map[string][]model.DictionarySimple {
	var result = map[string][]model.DictionarySimple{}
	// 系统字典
	result["system"] = service.GetDictionaryTreeSimple(1)
	// 组织内部字典
	result["self"] = service.GetDictionaryTreeSimple(m.ctx.GetOrgId())
	return result
}

// 组织账户列表
func (m *LoginComponentMethod) GetOrgAccount(orgId float64) []model.AccountList {
	return service.AccountList(&model.AccountQuery{OrgID: m.orgIdDefault(orgId)})
}

// 部门账户列表
func (m *LoginComponentMethod) GetDeptAccount(deptId float64) []model.AccountList {
	var result = make([]model.AccountList, 0)
	var accountIds []uint
	mysql.Table("dept_member").Where("dept_id=?", uint(deptId)).Pluck("account_id", &accountIds)
	if len(accountIds) == 0 {
		return result
	}
	return service.AccountList(&model.AccountQuery{AccountID: &accountIds})
}
