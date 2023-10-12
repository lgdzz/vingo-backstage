package vb

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lgdzz/vingo-backstage/vb/index"
	"github.com/lgdzz/vingo-backstage/vb/model"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/db/mysql"
	"github.com/lgdzz/vingo-utils/vingo/db/redis"
	"github.com/lgdzz/vingo-utils/vingo/jwt"
	"strconv"
	"strings"
)

// AdminAuthMiddle 验证账号
func AdminAuthMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtBody := jwt.JwtCheck[model.UserSimple](c.GetHeader("Authorization"), index.Config.System.Auth.Secret)
		var realName = jwtBody.Business.Realname
		if realName == "" {
			realName = jwtBody.Business.Username
		}
		c.Set("realName", realName)
		c.Set("userId", vingo.ToUint(jwtBody.ID))
	}
}

// AdminAccountMiddle 验证账户
func AdminAccountMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		aId, _ := strconv.Atoi(c.GetHeader("AccountId"))

		var account model.Account
		if !mysql.Exists(&account, aId) {
			panic("账户不存在")
		} else if account.UserID != c.GetUint("userId") {
			// 操作的账户不属于当前登录的账号
			panic(&vingo.AuthException{Message: "无效的账户参数"})
		} else if account.Status == vingo.Disable {
			panic(&vingo.AuthException{Message: "账户已被禁用"})
		}

		c.Set("accountId", uint(aId))   // 账户ID
		c.Set("roleId", account.RoleID) // 角色ID
		c.Set("orgId", account.OrgID)   // 组织ID

		c.Set("user", fmt.Sprintf("AccId=%d;OrgId=%d", c.GetUint("accountId"), c.GetUint("orgId")))
	}
}

// AdminRightMiddle 验证权限
func AdminRightMiddle() gin.HandlerFunc {
	return func(c *gin.Context) {
		debug := c.GetHeader("Debug")
		if debug == "1" {
			return
		}
		if index.Config.System.Right.Enable {
			right := fmt.Sprintf("%v:%v", c.Request.Method, c.FullPath())
			rights := make([]string, 0)
			rightsPointer := redis.HGet[string](index.Config.System.Right.Ticket, vingo.ToString(c.GetUint("accountId")))
			if rightsPointer != nil {
				rights = strings.Split(*rightsPointer, ",")
			}
			if !vingo.IsInSlice(right, rights) {
				panic("无接口使用权限")
			}
		}
	}
}
