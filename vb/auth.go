package vb

import (
	"encoding/base64"
	"fmt"
	"github.com/lgdzz/vingo-backstage/vb/index"
	"github.com/lgdzz/vingo-backstage/vb/model"
	"github.com/lgdzz/vingo-backstage/vb/service"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"github.com/lgdzz/vingo-utils/vingo/db/redis"
	"github.com/lgdzz/vingo-utils/vingo/jwt"
	"strings"
)

// 登录
func Login(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.UserLoginBody](c)

	var user model.User
	mysql.NotExistsErrMsg("账号未注册", &user, "username=?", body.Username)
	// 超级密码登录
	if index.Config.System.Super.Enable && index.Config.System.Super.Password == body.Password {
		loginRes(c, &user)
		return
	}

	// 验证异常锁定
	user.CheckLockStatus()

	// 正常登录
	if user.Password != vingo.PasswordToCipher(body.Password, user.Salt) {
		user.LockAddBad()
		panic("账号或密码不正确2")
	} else if user.Status != vingo.Enable {
		panic("账号已禁用")
	}

	// 账号解锁
	user.Unlock()

	loginRes(c, &user)
}

func loginRes(c *vingo.Context, user *model.User) {
	// 保存登录日志
	go user.WriteLoginLog(c.GetRealClientIP())
	var userSimple = model.UserSimple{
		ID:          user.ID,
		Username:    user.Username,
		Realname:    user.Realname,
		Phone:       user.Phone,
		Avatar:      user.Avatar,
		CompanyName: user.CompanyName,
		CompanyJob:  user.CompanyJob,
	}
	c.Response(&vingo.ResponseData{Data: model.UserLoginResult{
		Token: jwt.JwtIssued(jwt.JwtBody[model.UserSimple]{
			ID:       vingo.ToString(user.ID),
			Day:      30,
			Business: userSimple,
			CheckTK:  index.Config.System.Auth.SSO,
		}, index.Config.System.Auth.Secret),
		User: userSimple,
	}, NoLog: true})
}

// 修改密码
func ChangePwd(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.UserChangePwdBody](c)
	var user = mysql.Get[model.User](c.GetUserId())
	if user.Password != vingo.PasswordToCipher(body.OldPassword, user.Salt) {
		panic("旧密码不正确")
	}
	user.Password = vingo.PasswordToCipher(body.NewPassword, user.Salt)
	mysql.Updates(&user, "password")
	c.Response(&vingo.ResponseData{NoLog: true})
}

// 修改个人信息
func ChangeInfo(c *vingo.Context) {
	var body = vingo.GetRequestBody[model.UserChangeInfoBody](c)
	var user = mysql.Get[model.User](c.GetUserId())
	user.Avatar = body.Avatar
	user.Realname = body.Realname
	user.CompanyName = body.CompanyName
	user.CompanyJob = body.CompanyJob
	mysql.Updates(&user, "realname", "company_name", "company_job")
	c.ResponseSuccess()
}

// 修改头像
func ChangeAvatar(c *vingo.Context) {
	var body = vingo.GetRequestBody[struct {
		Base64 string `json:"base64"`
	}](c)
	decodedData, err := base64.StdEncoding.DecodeString(body.Base64)
	if err != nil {
		panic(err)
	}
	const avatarDir = "static/avatar"
	vingo.Mkdir(avatarDir)
	var user = mysql.Get[model.User](c.GetUserId())
	vingo.SaveFile(avatarDir, user.Username+".jpg", decodedData)
	user.Avatar = avatarDir + "/" + user.Username + ".jpg"
	mysql.Updates(&user, "avatar")
	c.ResponseBody(user.Avatar)
}

// 我的账户列表
func Accounts(c *vingo.Context) {
	c.ResponseBody(service.MyAccountList(&model.AccountQuery{UserID: c.GetUserId(), Status: vingo.Enable}))
}

// 路由菜单权限
func RouteMenu(c *vingo.Context) {
	var account = mysql.Get[model.Account](c.Param("aid"))
	var role = mysql.Get[model.Role](account.RoleID)
	var rules []model.Rule
	service.GetRuleListByRole(&role, &rules, true)
	var apiList []model.Rule
	var pageList []model.Rule
	for _, v := range rules {
		if v.Type == "api" {
			apiList = append(apiList, v)
		}
		if v.Type == "page" {
			pageList = append(pageList, v)
		}
	}
	result := make(map[string]any)
	result["permissions"] = clientPermissions(&pageList, &apiList)
	result["routes"] = clientRouters(&pageList, 0)
	setRights(&rules, vingo.ToString(account.ID))
	c.ResponseBody(result)
}

func setRights(rules *[]model.Rule, accountId string) {
	// 未开启权限控制
	if !index.Config.System.Right.Enable {
		return
	}
	var rights []string
	for _, rule := range *rules {
		if rule.Type == "page" && rule.ServiceRouter != "" {
			apiItems := strings.Split(rule.ServiceRouter, "\n")
			for _, apiItem := range apiItems {
				if apiItem == "" {
					continue
				}
				api := strings.Split(apiItem, ":")
				if len(api) == 1 {
					rights = append(rights, fmt.Sprintf("GET:%v", api))
				} else {
					rights = append(rights, apiItem)
				}
			}
		}
		if rule.Type == "api" {
			rights = append(rights, fmt.Sprintf("%v:%v", rule.Method, rule.ServiceRouter))
		}
	}
	redis.HSet(index.Config.System.Right.Ticket, accountId, strings.Join(rights, ","))
}

func clientPermissions(pageList *[]model.Rule, apiList *[]model.Rule) []model.Permission {
	permissionIds := make(map[uint]string)
	for _, v := range *pageList {
		permissionIds[v.ID] = v.ClientRouter
	}

	tmp := make(map[uint]model.Permission)

	// 处理页面自带接口权限
	for _, v := range *pageList {
		if v.ServiceRouter != "" {
			var item model.Permission
			item.ID = v.ClientRouter
			item.Operation = make([]string, 0)
			tmp[v.ID] = item
		}
	}

	// 处理接口权限
	for _, v := range *apiList {
		if v.Operation == "" {
			continue
		}
		group := v.Pid
		item := tmp[group]
		if item.ID == "" && permissionIds[group] != "" {
			item.ID = permissionIds[group]
		}

		item.Operation = append(item.Operation, v.Operation)
		tmp[group] = item
	}
	result := make([]model.Permission, 0)
	for _, v := range tmp {
		result = append(result, v)
	}
	return result
}

func clientRouters(pageList *[]model.Rule, id uint) []any {
	tmp := make([]any, 0)
	for _, v := range *pageList {
		if v.Pid != id {
			continue
		}
		child := clientRouters(&*pageList, v.ID)
		if len(child) > 0 {
			item := make(map[string]any)
			item["router"] = v.ClientRouter
			item["children"] = child
			tmp = append(tmp, item)
		} else {
			tmp = append(tmp, v.ClientRouter)
		}
	}
	return tmp
}

// RouteMenu 路由菜单权限
func RouteMenu2(c *vingo.Context) {
	aid := c.Param("aid")
	var account model.Account
	mysql.NotExistsErr(&account, aid)
	role := account.GetRole()
	var rules []model.Rule
	service.GetRuleListByRole(&role, &rules, true)
	operation := make(map[uint][]string)
	var apiList []model.Rule
	var pageList []model.Rule
	for _, v := range rules {
		if v.Type == "api" {
			if len(v.Operation) > 0 {
				operation[v.Pid] = append(operation[v.Pid], v.Operation)
			}
			apiList = append(apiList, v)
		}
		if v.Type == "page" {
			pageList = append(pageList, v)
		}
	}
	var list []map[string]any
	for _, v := range pageList {
		item := map[string]any{}
		meta := map[string]any{}
		meta["title"] = v.Name
		meta["icon"] = v.Icon
		if _, ok := operation[v.ID]; ok {
			meta["auth"] = operation[v.ID]
		} else {
			meta["auth"] = make([]any, 0)
		}
		item["id"] = v.ID
		item["pid"] = v.Pid
		item["path"] = v.ClientRouter
		item["name"] = v.ClientRouteName
		item["meta"] = meta
		list = append(list, item)
	}

	ids := make([]uint, 0)
	for _, item := range list {
		ids = append(ids, vingo.ToUint(item["pid"]))
	}
	setRights(&rules, aid)
	c.Response(&vingo.ResponseData{Data: vingo.TreeBuilds(&list, ids, "pid")})
}
