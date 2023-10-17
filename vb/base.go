package vb

import (
	"fmt"
	"github.com/lgdzz/vingo-backstage/vb/index"
	"github.com/lgdzz/vingo-backstage/vb/model"
	"github.com/lgdzz/vingo-backstage/vb/service"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"github.com/lgdzz/vingo-utils/vingo/db/page"
	"strconv"
)

// ======组织架构管理接口======

func UserList(c *vingo.Context) {
	var query = vingo.GetRequestQuery[model.UserQuery](c)
	db := mysql.Table("user u").Joins("left join `account` a on a.user_id=u.id AND a.deleted_at IS NULL").Joins("left join `user_lock` ul on ul.username=u.username AND ul.unlock_time IS NOT NULL AND ul.unlock_time>now()").Where("u.deleted_at IS NULL").Select("u.id,u.type,u.phone,u.realname,u.username,u.status,u.remark,u.last_ip lastIp,u.last_time lastTime,count(a.id) accNumber,if(ul.username IS NULL,false,true) lockStatus").Group("u.id")
	if query.Status != nil {
		db = db.Where("u.status=?", query.Status)
	}
	db = mysql.LikeOr(db, query.Keyword, "u.username", "u.phone", "u.realname")
	c.ResponseBody(page.New[map[string]any](db, page.Option{
		Limit: page.Limit{Page: query.Page, Size: query.Size},
		Order: page.OrderDefault(query.Order),
	}, nil))
}

// 注册账号
func UserCreate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.User](c)
	if body.Password == "" {
		// 设置默认密码
		body.Password = vingo.RandomString(6)
		if body.Remark != "" {
			body.Remark = fmt.Sprintf("%v；初始密码为：%v", body.Remark, body.Password)
		} else {
			body.Remark = fmt.Sprintf("初始密码为：%v", body.Password)
		}
	} else {
		// 验证密码强度
		vingo.PasswordStrength(body.Password, index.Config.System.Auth.Strength)
	}
	body.Salt = vingo.RandomString(5)
	body.Password = vingo.PasswordToCipher(body.Password, body.Salt)
	// 唯一验证
	service.CheckUserUnique(0, "username", body.Username)
	service.CheckUserUnique(0, "phone", body.Phone)
	mysql.Create(&body)
	c.ResponseSuccess()
}

func UserUpdate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.User](c)
	var user = mysql.Get[model.User](c.Param("id"))
	// 唯一验证
	service.CheckUserUnique(user.ID, "username", user.Username)
	service.CheckUserUnique(user.ID, "phone", user.Phone)
	var data = map[string]any{}
	data["phone"] = body.Phone
	data["username"] = body.Username
	data["realname"] = body.Realname
	data["status"] = body.Status
	data["remark"] = body.Remark
	if body.Password != "" {
		// 验证密码强度
		vingo.PasswordStrength(body.Password, index.Config.System.Auth.Strength)
		data["password"] = vingo.PasswordToCipher(body.Password, user.Salt)
	}
	mysql.Model(&user).Updates(data)
	c.ResponseSuccess()
}

func UserPatch(c *vingo.Context) {
	var body = vingo.GetRequestBody[struct {
		Field string `json:"field"`
		Value any    `json:"value"`
	}](c)
	var user = mysql.Get[model.User](c.Param("id"))
	user.CheckPatchWhite(body.Field)
	mysql.Model(&user).UpdateColumn(body.Field, body.Value)
	c.ResponseSuccess()
}

func UserUnlock(c *vingo.Context) {
	var user = mysql.Get[model.User](c.Param("id"))
	// 账号解锁
	user.Unlock()
	c.ResponseSuccess()
}

func UserDelete(c *vingo.Context) {
	var user = mysql.Get[model.User](c.Param("id"))
	if user.HasAccount() {
		panic(fmt.Sprintf("账号[%v]下面有账户信息，如要删除请先删除账户，或直接禁用账号", user.Username))
	}
	mysql.Delete(&user)
	c.ResponseSuccess()
}

// API组织类型列表
func OrgGradeList(c *vingo.Context) {
	c.ResponseBody(service.GetOrgGradeTree(0))
}

// API组织类型创建
func OrgGradeCreate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.OrgGrade](c)
	body.Sort = 255
	body.Extends = model.OrgGradeExtends{}

	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Create(&body)
	mysql.SetPath(tx, &body)
	c.ResponseSuccess()
}

// API组织类型修改
func OrgGradeUpdate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.OrgGrade](c)
	var grade = mysql.Get[model.OrgGrade](c.Param("id"))
	mysql.Model(&grade).Select("name", "name_en", "description", "extends", "sort", "admin_role_id").Updates(&body)
	c.ResponseSuccess()
}

// API组织类型删除
func OrgGradeDelete(c *vingo.Context) {
	var grade = mysql.Get[model.OrgGrade](c.Param("id"))
	mysql.CheckHasChild(&model.OrgGrade{}, grade.Id)
	mysql.Delete(&grade)
	c.ResponseSuccess()
}

// API组织列表
func OrgList(c *vingo.Context) {
	query := model.OrgQuery{RootID: c.GetUint("orgId")}
	c.RequestQuery(&query)
	c.ResponseBody(service.GetOrgTree(&query))
}

// API组织详情
func OrgDetail(c *vingo.Context) {
	c.ResponseBody(mysql.Get[model.Org](c.Param("id")))
}

// API组织创建
func OrgCreate(c *vingo.Context) {
	org := model.Org{Pid: c.GetOrgId(), GradeID: 2, Pids: make([]uint, 0), Status: 1}
	c.RequestBody(&org)
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), org.Pid)
	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Create(&org)
	org.SetPath(tx)
	c.ResponseSuccess()
}

// API组织修改
func OrgUpdate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.Org](c)
	var org = mysql.Get[model.Org](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), org.ID)
	mysql.Model(&org).Select("name", "full_name", "code", "status", "sort", "description", "contact_name", "contact_tel", "contact_address", "province_code", "city_code", "county_code", "province", "city", "county").Updates(&body)
	c.ResponseSuccess()
}

// API组织删除
func OrgDelete(c *vingo.Context) {
	var org = mysql.Get[model.Org](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), org.ID)
	mysql.CheckHasChild(&model.Org{}, org.ID)
	mysql.Delete(&org)
	c.ResponseSuccess()
}

// API组织扩展
func OrgGetExtend(c *vingo.Context) {
	var org = mysql.Get[model.Org](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), org.ID)
	var extend = model.OrgExtend{OrgID: org.ID, ExpireDate: vingo.LocalTime{}.Now()}
	mysql.FirstOrCreate(&extend)
	c.ResponseBody(extend)
}

// API组织扩展
func OrgUpdateExtend(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.OrgExtend](c)
	var org = mysql.Get[model.Org](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), org.ID)
	var extend = mysql.GetByColumn[model.OrgExtend]("org_id=?", org.ID)
	mysql.Model(&extend).Select("max_member", "expire_date").Updates(&body)
	c.ResponseSuccess()
}

// API部门列表
func DeptList(c *vingo.Context) {
	var query = vingo.GetRequestQuery[model.DeptQuery](c)
	if query.OrgID == 0 {
		query.OrgID = c.GetOrgId()
	}
	c.ResponseBody(service.GetOrgDeptTree(&query))
}

// API部门创建
func DeptCreate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.Dept](c)
	if body.OrgID == 0 {
		body.OrgID = c.GetOrgId()
	}
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), body.OrgID)
	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Create(&body)
	mysql.SetPath(tx, &body)
	c.ResponseSuccess()
}

// API部门修改
func DeptUpdate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.Dept](c)
	var dept = mysql.Get[model.Dept](c.Param("id"))
	// 安全验证
	if body.Pid == dept.ID {
		panic("上级部门不能是其自身")
	}
	service.CheckOrgRight(c.GetUint("orgId"), dept.OrgID)

	var diff = vingo.DiffBox{Old: dept, New: body}

	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	// 修改记录
	tx.Model(&dept).Select("pid", "name", "code", "sort", "status").Updates(&body)
	// 修改了上级部门需要重新生成path信息，包括所有子级的path信息
	if diff.IsChange("Pid") {
		mysql.SetPathAndChildPath(tx, &dept)
	}
	c.ResponseSuccess()
}

// API部门删除
func DeptDelete(c *vingo.Context) {
	var dept = mysql.Get[model.Dept](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), dept.OrgID)

	if mysql.Exists(&model.Dept{}, "org_id=? AND pid=?", dept.OrgID, dept.ID) {
		panic("记录有子项，删除失败")
	}

	mysql.Delete(&dept)
	c.ResponseSuccess()
}

// API部门加入成员
func DeptJoinMember(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.DeptMemberItem](c)
	// 获取正常的部门
	dept := service.GetOkDept(body.DeptID)
	// 获取正常的账户
	account := service.GetOkAccount(body.AccountID)
	// 验证部门所属组织是否和账户所属组织为同一个组织
	if dept.OrgID != account.OrgID {
		panic("部门和账户不属于同一个组织")
	}
	// 验证账户是否已经在部门内
	if service.CheckMemberIsInDept(dept.ID, account.ID) {
		panic("该成员已经在部门中")
	}
	// 验证是否对该部门有管理权限
	service.CheckOrgRight(c.GetOrgId(), dept.OrgID)
	mysql.Create(&model.DeptMember{OrgID: dept.OrgID, DeptID: dept.ID, AccountID: account.ID})
	c.ResponseSuccess()
}

// API部门移出成员
func DeptRemoveMember(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.DeptMemberItem](c)
	var member = mysql.GetByColumn[model.DeptMember]("dept_id=? AND account_id=?", body.DeptID, body.AccountID)
	// 验证是否对该部门有管理权限
	service.CheckOrgRight(c.GetOrgId(), member.OrgID)
	mysql.Delete(&member)
	c.ResponseSuccess()
}

// API账户列表
func AccountList(c *vingo.Context) {
	var query = vingo.GetRequestQuery[model.AccountQuery](c)
	c.ResponseBody(service.AccountList(&query))
}

// API账户创建
func AccountCreate(c *vingo.Context) {
	selfOrgId := c.GetOrgId()
	var body = vingo.GetRequestBody[model.AccountBody](c)
	if body.OrgID == 0 {
		body.OrgID = selfOrgId
	}

	// 安全验证
	service.CheckOrgRight(selfOrgId, body.OrgID)
	service.CheckOrgRole(body.OrgID, body.RoleID)
	user := model.User{}
	switch body.Type {
	case "reg":
		// 验证密码强度
		vingo.PasswordStrength(body.Password, index.Config.System.Auth.Strength)
		// 账号不存在，先注册再绑定
		func(body *model.AccountBody) {
			service.CheckUserUnique(0, "username", body.Username) // 验证用户名是否存在
			service.CheckUserUnique(0, "phone", body.Username)    // 验证手机号是否存在
			user.Username = body.Username
			user.Realname = body.Realname
			user.Phone = body.Phone
			user.Remark = body.Remark
			user.FromID = body.FromID
			user.Salt = vingo.RandomString(5)
			user.Password = vingo.PasswordToCipher(body.Password, user.Salt)
			user.Status = vingo.Enable
			tx := mysql.Begin()
			defer mysql.AutoCommit(tx)
			// 注册账号
			tx.Create(&user)
			tx.Create(&model.Account{OrgID: body.OrgID, RoleID: body.RoleID, UserID: user.ID})
		}(&body)
	case "bind":
		if !mysql.Exists(&user, "username=?", body.Username) {
			panic(fmt.Sprintf("账号[%v]未注册", body.Username))
		} else if index.Config.System.Account.Many == false {
			panic("账号已存在，系统未开启多账户绑定")
		}
		service.CheckAccountRoleIsExists(user.ID, body.OrgID, body.RoleID, 0)
		// 账号存在，直接绑定
		mysql.Create(&model.Account{OrgID: body.OrgID, RoleID: body.RoleID, UserID: user.ID})
	default:
		panic("type可选参数[reg|bind]")
	}
	c.ResponseSuccess()
}

// API账户修改
func AccountUpdate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.AccountBody](c)
	var account = mysql.Get[model.Account](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), account.OrgID)
	service.CheckOrgRole(body.OrgID, body.RoleID)
	service.CheckAccountRoleIsExists(account.UserID, account.OrgID, body.RoleID, account.ID)
	if body.Status == vingo.Disable && account.ID == c.GetAccId() {
		panic("不能禁用当前登录账户")
	}
	account.RoleID = body.RoleID
	account.Status = body.Status
	// 多账户模式
	if index.Config.System.Account.Many {
		mysql.Select("role_id", "status").Updates(&account)
	} else {
		user := account.GetUser()
		service.CheckUserUnique(user.ID, "username", body.Username)
		service.CheckUserUnique(user.ID, "phone", body.Phone)
		user.Username = body.Username
		user.Phone = body.Phone
		user.Realname = body.Realname
		tx := mysql.Begin()
		defer mysql.AutoCommit(tx)
		tx.Select("role_id", "status").Updates(&account)
		tx.Select("username", "phone", "realname").Updates(&user)
	}

	c.ResponseSuccess()
}

// API账户删除
func AccountDelete(c *vingo.Context) {
	var account = mysql.Get[model.Account](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), account.OrgID)
	if account.ID == c.GetAccId() {
		panic("禁止删除当前登录用户")
	}
	// 多账户模式
	if index.Config.System.Account.Many {
		mysql.Delete(&account)
	} else {
		user := account.GetUser()
		tx := mysql.Begin()
		defer mysql.AutoCommit(tx)
		tx.Delete(&account)
		tx.Delete(&user)
	}
	c.ResponseSuccess()
}

// API角色列表
func RoleList(c *vingo.Context) {
	data := service.GetRoleTreeByOrgId(vingo.ToUint(c.DefaultQuery("orgId", vingo.ToString(c.GetOrgId()))), 0, c)
	c.ResponseBody(data)
}

// API角色创建
func RoleCreate(c *vingo.Context) {
	body := model.Role{Pid: c.GetOrgId(), Extends: make(map[string]any)}
	c.RequestBody(&body)

	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), body.OrgID)

	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Omit("AllowEdit").Create(&body)
	body.SetPath(tx)
	c.ResponseSuccess()
}

// API角色修改
func RoleUpdate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.Role](c)
	var role = mysql.Get[model.Role](c.Param("id"))
	// 安全验证
	if body.Pid == role.ID {
		panic("上级角色不能是其自身")
	}
	service.CheckOrgRight(c.GetOrgId(), body.OrgID)
	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Model(&role).Select("pid", "name", "status", "remark", "extends", "rules", "half_rules").Updates(&body)
	role.SetPath(tx)
	c.ResponseSuccess()
}

// API角色删除
func RoleDelete(c *vingo.Context) {
	var role = mysql.Get[model.Role](c.Param("id"))
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), role.OrgID)
	if role.HasAccount() {
		panic(fmt.Sprintf("角色[%v]被账户使用，如要删除请先删除使用该角色的账户信息", role.Name))
	}
	mysql.Delete(&role)
	c.ResponseSuccess()
}

// API修改超管的过滤字段
func RoleSuperHalf(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.Role](c)
	var role = mysql.Get[model.Role](1)
	// 安全验证
	service.CheckOrgRight(c.GetOrgId(), role.OrgID)
	mysql.Model(&role).Select("half_rules").Updates(&body)
	c.ResponseSuccess()
}

// API权限规则列表
func RuleList(c *vingo.Context) {
	c.ResponseBody(service.GetRuleTreeByRoleId(0, c.DefaultQuery("keyword", "")))
}

// API权限规则创建
func RuleCreate(c *vingo.Context) {
	body := model.Rule{Pid: c.GetOrgId()}
	c.RequestBody(&body)
	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Create(&body)
	body.SetPath(tx)
	c.ResponseSuccess()
}

// API权限规则修改
func RuleUpdate(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.Rule](c)
	var rule = mysql.Get[model.Rule](c.Param("id"))
	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Model(&rule).Select("pid", "name", "type", "method", "permission_id", "operation", "service_router", "client_router", "client_route_name", "client_route_alias", "icon", "sort").Updates(&body)
	rule.SetPath(tx)
	c.ResponseSuccess()
}

// API权限规则删除
func RuleDelete(c *vingo.Context) {
	var rule = mysql.Get[model.Rule](c.Param("id"))
	// 安全验证，如果有子级数据禁止删除
	mysql.CheckHasChild(&model.Rule{}, rule.ID)
	mysql.Delete(&rule)
	c.ResponseSuccess()
}

// API字典列表
func DictionaryTree(c *vingo.Context) {
	c.Response(&vingo.ResponseData{Data: service.GetDictionaryTree(c.GetOrgId())})
}

// API字典创建
func DictionaryCreate(c *vingo.Context) {
	body := model.Dictionary{OrgID: c.GetOrgId()}
	c.RequestBody(&body)
	if body.Value == "" {
		body.Value = body.Description
	}
	tx := mysql.Begin()
	defer mysql.AutoCommit(tx)
	tx.Create(&body)
	if body.Pid > 0 {
		var parent model.Dictionary
		mysql.TXNotExistsErr(tx, &parent, body.Pid)
		body.Name = fmt.Sprintf("%v_%d", parent.Name, body.ID)
		body.Path = fmt.Sprintf("%v,%d", parent.Path, body.ID)
		body.Len = parent.Len + 1
	} else {
		body.Path = strconv.Itoa(int(body.ID))
		body.Len = 1
	}
	tx.Select("name", "path", "len").Updates(&body)
	c.ResponseSuccess()
}

// API字典详情
func DictionaryDetail(c *vingo.Context) {
	dictionary := model.Dictionary{}
	mysql.NotExistsErr(&dictionary, c.Param("id"))
	c.Response(&vingo.ResponseData{Data: &dictionary})
}

// API字典修改
func DictionaryUpdate(c *vingo.Context) {
	body := model.Dictionary{}
	c.RequestBody(&body)
	if body.Value == "" {
		body.Value = body.Description
	}
	dictionary := model.Dictionary{}
	mysql.NotExistsErr(&dictionary, c.Param("id"))
	mysql.Model(&dictionary).Select("description", "value", "sort").Updates(&body)
	c.ResponseSuccess()
}

// API字典删除
func DictionaryDelete(c *vingo.Context) {
	dictionary := model.Dictionary{}
	mysql.NotExistsErr(&dictionary, c.Param("id"))
	mysql.CheckHasChild(&model.Dictionary{}, dictionary.ID)
	mysql.Delete(&dictionary)
	c.ResponseSuccess()
}

// API查询日志文件列表
func GetLogFiles(c *vingo.Context) {
	c.Response(&vingo.ResponseData{Data: vingo.GetLogFiles(), NoLog: true})
}

// API查询日志
func FindLogs(c *vingo.Context) {
	source, _ := c.GetQuery("source")
	keyword, _ := c.GetQuery("keyword")
	c.Response(&vingo.ResponseData{Data: vingo.FindLogs(source, keyword), NoLog: true})
}
