// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/npiganeau/yep/yep/server"
)

func CallKW(c *gin.Context) {
	sess := sessions.Default(c)
	uid := sess.Get("uid").(int64)
	var params server.CallParams
	server.BindRPCParams(c, &params)
	res, err := server.Execute(uid, params)
	server.RPC(c, http.StatusOK, res, err)
}

func SearchRead(c *gin.Context) {
	sess := sessions.Default(c)
	uid := sess.Get("uid").(int64)
	var params server.SearchReadParams
	server.BindRPCParams(c, &params)
	res, err := server.SearchRead(uid, params)
	server.RPC(c, http.StatusOK, res, err)
}
