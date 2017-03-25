// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/server"
)

// A SessionInfo holds data about an application session
type SessionInfo struct {
	SessionID   int64                  `json:"session_id"`
	UID         int64                  `json:"uid"`
	UserContext map[string]interface{} `json:"user_context"`
	DB          string                 `json:"db"`
	Username    string                 `json:"username"`
	CompanyID   int64                  `json:"company_id"`
}

// GetSessionInfo returns session info to the client
func GetSessionInfo(c *server.Context) {
	var (
		userContext *types.Context
		companyID   int64
	)
	sess := c.Session()
	if sess.Get("uid") != nil {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := pool.ResUsers().NewSet(env).Search(pool.ResUsers().ID().Equals(sess.Get("uid").(int64)))
			userContext = user.ContextGet()
			companyID = user.Company().ID()
		})
		c.RPC(http.StatusOK, SessionInfo{
			SessionID:   sess.Get("ID").(int64),
			UID:         sess.Get("uid").(int64),
			UserContext: userContext.ToMap(),
			DB:          "default",
			Username:    sess.Get("login").(string),
			CompanyID:   companyID,
		})
	}
	c.RPC(http.StatusOK, SessionInfo{})
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
