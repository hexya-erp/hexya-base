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
	var params CallParams
	server.BindRPCParams(c, &params)
	res, err := Execute(uid, params)
	server.RPC(c, http.StatusOK, res, err)
}

func SearchRead(c *gin.Context) {
	sess := sessions.Default(c)
	uid := sess.Get("uid").(int64)
	var params searchReadParams
	server.BindRPCParams(c, &params)
	res, err := searchRead(uid, params)
	server.RPC(c, http.StatusOK, res, err)
}
