// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/pool"
)

// SessionInfo returns a map with information about the given session
func SessionInfo(sess sessions.Session) gin.H {
	var (
		userContext *types.Context
		companyID   int64
	)
	if sess.Get("uid") != nil {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := pool.User().Search(env, pool.User().ID().Equals(sess.Get("uid").(int64)))
			userContext = user.ContextGet()
			companyID = user.Company().ID()
		})
		return gin.H{
			"session_id":   sess.Get("ID"),
			"uid":          sess.Get("uid"),
			"user_context": userContext.ToMap(),
			"db":           "default",
			"username":     sess.Get("login"),
			"company_id":   companyID,
		}
	}
	return gin.H{}
}

// GetSessionInfo returns the information fo the current session
// to the client
func GetSessionInfo(c *server.Context) {
	c.RPC(http.StatusOK, SessionInfo(c.Session()))
}

// Modules returns the list of installed modules to the client
func Modules(c *server.Context) {
	mods := make([]string, len(server.Modules))
	for i, m := range server.Modules {
		mods[i] = m.Name
	}
	c.RPC(http.StatusOK, mods)
}

// Logout the current user and redirect to login page
func Logout(c *server.Context) {
	sess := c.Session()
	sess.Delete("uid")
	sess.Delete("ID")
	sess.Delete("login")
	sess.Save()
	redirect := c.DefaultQuery("redirect", "/web/login")
	c.Redirect(http.StatusSeeOther, redirect)
}

type ChangePasswordData struct {
	Fields []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"fields"`
}

// ChangePassword is called by the client to change the current user password
func ChangePassword(c *server.Context) {
	uid := c.Session().Get("uid").(int64)
	var params ChangePasswordData
	c.BindRPCParams(&params)
	var oldPassword, newPassword, confirmPassword string
	for _, d := range params.Fields {
		switch d.Name {
		case "old_pwd":
			oldPassword = d.Value
		case "new_password":
			newPassword = d.Value
		case "confirm_pwd":
			confirmPassword = d.Value
		}
	}
	res := make(gin.H)
	err := models.ExecuteInNewEnvironment(uid, func(env models.Environment) {
		rs := pool.User().NewSet(env)
		if strings.TrimSpace(oldPassword) == "" ||
			strings.TrimSpace(newPassword) == "" ||
			strings.TrimSpace(confirmPassword) == "" {
			log.Panic(rs.T("You cannot leave any password empty."))
		}
		if newPassword != confirmPassword {
			log.Panic(rs.T("The new password and its confirmation must be identical."))
		}
		if rs.ChangePassword(oldPassword, newPassword) {
			res["new_password"] = newPassword
			return
		}
		log.Panic(rs.T("Error, password not changed !"))
	})
	c.RPC(http.StatusOK, res, err)
}
