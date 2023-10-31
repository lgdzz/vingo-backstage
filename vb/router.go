package vb

import (
	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-utils/vingo"
)

// RegisterAdmin 后台管理接口路由注册
func RegisterAdmin(g *gin.RouterGroup) {
	vingo.RoutesPost(g, "/login", Login)

	// ======注册登录验证中间件======
	g.Use(AdminAuthMiddle())

	vingo.RoutesPost(g, "/change-pwd", ChangePwd)
	vingo.RoutesGet(g, "/accounts", Accounts)
	vingo.RoutesGet(g, "/route-menu/:aid", RouteMenu2)
	vingo.RoutesGet(g, "/route-menu-ant/:aid", RouteMenu)

	// ======注册账户中间件======
	g.Use(AdminAccountMiddle())

	vingo.RoutesGet(g, "/openapi.org.account", SearchOrgAccount)

	vingo.RoutesPost(g, "/components", LoginComponent)

	// ======权限验证中间件======
	g.Use(AdminRightMiddle())

	vingo.RoutesGet(g, "/user", UserList)
	vingo.RoutesPost(g, "/user", UserCreate)
	vingo.RoutesPut(g, "/user/:id", UserUpdate)
	vingo.RoutesPatch(g, "/user/:id", UserPatch)
	vingo.RoutesPut(g, "/user.unlock/:id", UserUnlock)
	vingo.RoutesDelete(g, "/user/:id", UserDelete)

	vingo.RoutesGet(g, "/org-grade", OrgGradeList)
	vingo.RoutesPost(g, "/org-grade", OrgGradeCreate)
	vingo.RoutesPut(g, "/org-grade/:id", OrgGradeUpdate)
	vingo.RoutesDelete(g, "/org-grade/:id", OrgGradeDelete)

	vingo.RoutesGet(g, "/org", OrgList)
	vingo.RoutesGet(g, "/org/:id", OrgDetail)
	vingo.RoutesPost(g, "/org", OrgCreate)
	vingo.RoutesPut(g, "/org/:id", OrgUpdate)
	vingo.RoutesDelete(g, "/org/:id", OrgDelete)

	vingo.RoutesGet(g, "/org.extend/:id", OrgGetExtend)
	vingo.RoutesPut(g, "/org.extend/:id", OrgUpdateExtend)

	vingo.RoutesGet(g, "/dept", DeptList)
	vingo.RoutesPost(g, "/dept", DeptCreate)
	vingo.RoutesPut(g, "/dept/:id", DeptUpdate)
	vingo.RoutesDelete(g, "/dept/:id", DeptDelete)
	vingo.RoutesPost(g, "/dept-member", DeptJoinMember)
	vingo.RoutesDelete(g, "/dept-member", DeptRemoveMember)

	vingo.RoutesGet(g, "/account", AccountList)
	vingo.RoutesPost(g, "/account", AccountCreate)
	vingo.RoutesPut(g, "/account/:id", AccountUpdate)
	vingo.RoutesDelete(g, "/account/:id", AccountDelete)

	vingo.RoutesGet(g, "/role", RoleList)
	vingo.RoutesPost(g, "/role", RoleCreate)
	vingo.RoutesPut(g, "/role/:id", RoleUpdate)
	vingo.RoutesDelete(g, "/role/:id", RoleDelete)
	vingo.RoutesPost(g, "/role.super.half", RoleSuperHalf)

	vingo.RoutesGet(g, "/rule", RuleList)
	vingo.RoutesPost(g, "/rule", RuleCreate)
	vingo.RoutesPut(g, "/rule/:id", RuleUpdate)
	vingo.RoutesDelete(g, "/rule/:id", RuleDelete)

	vingo.RoutesGet(g, "/dictionary", DictionaryTree)
	vingo.RoutesGet(g, "/dictionary/:id", DictionaryDetail)
	vingo.RoutesPost(g, "/dictionary", DictionaryCreate)
	vingo.RoutesPut(g, "/dictionary/:id", DictionaryUpdate)
	vingo.RoutesDelete(g, "/dictionary/:id", DictionaryDelete)

	vingo.RoutesGet(g, "/search-logs-files", GetLogFiles)
	vingo.RoutesGet(g, "/search-logs-contents", FindLogs)
}
