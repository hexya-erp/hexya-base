// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/server"
)

func SessionInfo(sess sessions.Session) gin.H {
	var (
		userContext *types.Context
		companyID   int64
	)
	if sess.Get("uid") != nil {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := pool.ResUsers().NewSet(env).Search(pool.ResUsers().ID().Equals(sess.Get("uid").(int64)))
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

func GetSessionInfo(c *server.Context) {
	c.RPC(http.StatusOK, SessionInfo(c.Session()))
}

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
