package service

import (
	"fmt"
	"github.com/lgdzz/vingo-backstage/vb/model"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"github.com/lgdzz/vingo-utils/vingo/db/page"
	"strings"
)

// 账号唯一验证
func CheckUserUnique(id uint, column string, value string) {
	if value == "" {
		return
	}
	var user model.User
	db := mysql.Where(fmt.Sprintf("`%v`='%v'", column, value))
	if id > 0 {
		db = db.Where("`id`<>?", id)
	}
	if err := db.First(&user).Error; err != nil {
		return
	}
	panic(fmt.Sprintf("%v已存在", value))
}

// 验证当前组织是否有管理目标组织的权限
func CheckOrgRight(seniorOrgId uint, targetOrgId uint) {
	var targetOrg model.Org
	mysql.NotExistsErrMsg("目标组织不存在", &targetOrg, "id=?", targetOrgId)
	if seniorOrgId != targetOrgId && !vingo.IsInSlice(seniorOrgId, targetOrg.Pids) {
		panic("无权管理此组织内容")
	}
}

// 验证角色是否属于组织
func CheckOrgRole(orgId uint, roleId uint) {
	var role model.Role
	if !mysql.Exists(&role, "id=? AND org_id=?", roleId, orgId) && !vingo.IsInSlice(roleId, GetOrgDefaultRoleID(orgId)) {
		panic("角色不存在")
	}
}

// 验证账户是否存在相同角色
func CheckAccountRoleIsExists(userId uint, orgId uint, roleId uint, accountId uint) {
	var count int64
	db := mysql.Table("account").Where("user_id=? AND org_id=? AND role_id=? AND deleted_at IS NULL", userId, orgId, roleId)
	if accountId > 0 {
		db = db.Where("id <> ?", accountId)
	}
	db.Count(&count)
	if count > 0 {
		panic("该账号在此组织内已存在相同角色")
	}
}

// 获取组织的管理结构
func GetOrgTree(query *model.OrgQuery) []map[string]any {
	var orgList []model.OrgTreeItem
	db := mysql.Table("org AS o").Joins("left join `org_grade` AS og on og.id=o.grade_id").Joins("left join `role` on og.admin_role_id=role.id").Where("o.deleted_at IS NULL")
	if query.RootID > 0 {
		db = db.Where("find_in_set(?,o.path)", query.RootID)
	}
	if query.GradeID > 0 {
		db = db.Where("grade_id=?", query.GradeID)
	}
	db = mysql.LikeOr(db, query.Keyword, "o.name", "o.full_name")
	db.Order("o.len asc,o.sort asc,o.id asc").Select("o.id,o.pid,o.name,o.full_name,o.code,o.contact_name,o.contact_tel,o.contact_address,o.province,o.city,o.county,o.grade_id,og.name grade_name,role.name admin_role_name").Scan(&orgList)

	return vingo.Tree(&orgList, "pid", false)
}

// 获取组织类型树
func GetOrgGradeTree(gradeId uint) []map[string]any {
	// 查出符合条件的组织类型列表
	var grades []model.OrgGradeTreeItem
	db := mysql.Table("org_grade AS og").Joins("left join `role` on role.id=og.admin_role_id").Where("og.deleted_at IS NULL")
	if gradeId > 0 {
		db = db.Where("find_in_set(?,og.path)", gradeId)
	}
	db.Order("og.len asc,og.id asc").Select("og.id,og.pid,og.name,og.name_en,og.description,og.extends,og.status,og.sort,og.admin_role_id,role.name role_name,(select count(*) from org o where o.grade_id=og.id and o.deleted_at IS NULL) org_count").Find(&grades)

	return vingo.Tree(&grades, "pid", false)
}

// 获取组织部门树
func GetOrgDeptTree(query *model.DeptQuery) []map[string]any {
	var deptList []model.DeptTreeItem
	db := mysql.Table("dept").Joins("left join `org` AS o on o.id=dept.org_id").Where("dept.deleted_at IS NULL AND o.deleted_at IS NULL")
	if query.OrgID > 0 {
		db = db.Where("dept.org_id=?", query.OrgID)
	}
	db = mysql.LikeOr(db, query.Keyword, "dept.name")
	db.Order("dept.len asc,dept.sort desc,dept.id asc").Select("dept.id,dept.pid,dept.name,dept.code,dept.status,dept.org_id,(select count(*) from `dept_member` where dept.id=dept_member.dept_id) member_num").Scan(&deptList)

	return vingo.Tree(&deptList, "pid", false)
}

// 获取组织默认角色ID（默认管理员角色+共享角色）
func GetOrgDefaultRoleID(orgId uint) (ids []uint) {
	var grade model.OrgGrade
	mysql.Table("org_grade AS og").Joins("left join `org` on org.grade_id=og.id").Where("org.id=?", orgId).First(&grade)
	if grade.AdminRoleID > 0 {
		ids = append(ids, grade.AdminRoleID)
	}
	if len(grade.Extends.ShareRoleIds) > 0 {
		ids = append(ids, grade.Extends.ShareRoleIds...)
	}
	return
}

// 获取组织的角色树
func GetRoleTreeByOrgId(orgId uint, status int8, ctx *vingo.Context) (result []map[string]any) {
	var roles []model.Role
	db := mysql.Where("org_id=? OR id IN(?)", orgId, GetOrgDefaultRoleID(orgId))
	// 如果角色单位等于当前账户单位，则增加path条件，判定为本单位账户操作，只查询自身及以下的角色，防止越权操作
	if orgId == ctx.GetOrgId() {
		db = db.Where("FIND_IN_SET(?,path)", ctx.GetRoleId())
	}
	if status > 0 {
		db = db.Where("status=?", status)
	}
	db.Find(&roles)
	if len(roles) == 0 {
		result = []map[string]any{}
		return
	}

	for i, role := range roles {
		if i > 0 && role.OrgID == orgId {
			roles[i].AllowEdit = true
		}
	}
	result = vingo.Tree(&roles, "pid", false)
	return
}

// 获取所有权限规则的树结构数据
func GetRuleTreeByRoleId(roleId uint, keyword string) []map[string]any {
	var rules []model.Rule
	if roleId > 0 {
		var role model.Role
		mysql.Where("id=?", roleId).First(&role)
		GetRuleListByRole(&role, &rules, false)
	} else {
		db := mysql.Db
		db = mysql.LikeOr(db, keyword, "name")
		db.Find(&rules)
	}
	return vingo.Tree(&rules, "pid", false)
}

// 通过角色获取权限规则列表
func GetRuleListByRole(role *model.Role, rules *[]model.Rule, half bool) {
	if role.Master == vingo.True {
		if half {
			var halfRules = role.HalfRules.Uints()
			if len(halfRules) == 0 {
				halfRules = append(halfRules, 0)
			}
			mysql.Where("id NOT IN(?)", halfRules).Order("sort asc,id asc").Find(&rules)
		} else {
			mysql.Db.Order("sort asc,id asc").Find(&rules)
		}
	} else {
		if half {
			role.Rules = vingo.UintSliceDiff(role.Rules, role.HalfRules)
		}

		mysql.Where("id in(?)", role.Rules.Uints()).Find(&rules)

		var paths []string
		for _, v := range *rules {
			paths = append(paths, strings.Split(strings.Trim(v.Path, ","), ",")...)
		}
		mysql.Where("id in(?)", vingo.SliceStringToInt(vingo.SliceUniqueString(paths))).Order("sort asc,id asc").Find(&rules)
	}
	return
}

// 获取账户列表
func AccountList(query *model.AccountQuery) page.Result {
	db := mysql.Table("account AS acc").Joins("inner join `user` on acc.user_id=user.id").Joins("inner join `org` on acc.org_id=org.id").Joins("inner join `org_grade` AS og on org.grade_id=og.id").Joins("inner join `role` on acc.role_id=role.id").Select("acc.id,acc.status,acc.user_id,acc.org_id,acc.role_id,user.realname,user.username,user.phone,user.avatar,user.company_name,user.company_job,user.from_id,org.pid AS org_pid,org.name AS org_name,role.name AS role_name,og.name AS org_grade_name,og.id org_grade_id").Where("acc.deleted_at IS NULL AND user.deleted_at IS NULL AND org.deleted_at IS NULL AND og.deleted_at IS NULL AND role.deleted_at IS NULL")
	if query.OrgID > 0 {
		db = db.Where("acc.org_id=?", query.OrgID)
	}
	db = mysql.LikeOr(db, query.Keyword, "username", "phone", "realname", "company_name")
	if query.UserID > 0 {
		db = db.Where("acc.user_id=?", query.UserID)
	}
	if query.Status > 0 {
		db = db.Where("acc.status=?", query.Status)
	}
	if query.AccountID != nil {
		db = db.Where("acc.id IN (?)", *query.AccountID)
	}
	return page.New[model.AccountList](db, page.Option{
		Limit: page.Limit{Page: query.Page, Size: query.Size},
		Order: page.OrderDefault(query.Order),
	}, nil)
}

// 获取我的账户列表
func MyAccountList(query *model.AccountQuery) (result []model.AccountList) {
	db := mysql.Table("account AS acc").Joins("inner join `user` on acc.user_id=user.id").Joins("inner join `org` on acc.org_id=org.id").Joins("inner join `org_grade` AS og on org.grade_id=og.id").Joins("inner join `role` on acc.role_id=role.id").Select("acc.id,acc.status,acc.user_id,acc.org_id,acc.role_id,user.realname,user.username,user.phone,user.from_id,org.pid AS org_pid,org.name AS org_name,role.name AS role_name,og.name AS org_grade_name,og.id org_grade_id").Where("acc.deleted_at IS NULL AND user.deleted_at IS NULL AND org.deleted_at IS NULL AND og.deleted_at IS NULL AND role.deleted_at IS NULL")
	if query.UserID > 0 {
		db = db.Where("acc.user_id=?", query.UserID)
	}
	if query.Status > 0 {
		db = db.Where("acc.status=?", query.Status)
	}
	db.Order("id desc").Scan(&result)
	return
}

// 获取字典列表
func GetDictionaryTree(orgId uint) []map[string]any {
	var dictionary []model.Dictionary
	mysql.Model(&model.Dictionary{}).Where("org_id=?", orgId).Order("len asc,sort asc,id asc").Find(&dictionary)
	return vingo.Tree(&dictionary, "pid", false)
}

// 获取字典列表(简单结构)
func GetDictionaryTreeSimple(orgId uint) (result []model.DictionarySimple) {
	tree := GetDictionaryTree(orgId)
	vingo.CustomOutput(&tree, &result)
	return
}

// 获取正常的账户模型
func GetOkAccount(accountId uint) model.Account {
	var account model.Account
	mysql.NotExistsErr(&account, "id=?", accountId)
	if account.Status == vingo.Disable {
		panic("账户已被禁用")
	}
	return account
}

// 获取正常的部门模型
func GetOkDept(deptId uint) model.Dept {
	var dept model.Dept
	mysql.NotExistsErr(&dept, "id=?", deptId)
	if dept.Status == vingo.Disable {
		panic("部门已被禁用")
	}
	return dept
}

// 验证成员是否在部门中
func CheckMemberIsInDept(deptId uint, accountId uint) bool {
	return mysql.Exists(&model.DeptMember{}, "dept_id=? AND account_id=?", deptId, accountId)
}
