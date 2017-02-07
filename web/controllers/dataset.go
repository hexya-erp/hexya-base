// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/npiganeau/yep/yep/server"
)

func CallKW(c *server.Context) {
	uid := c.Session().Get("uid").(int64)
	var params CallParams
	c.BindRPCParams(&params)
	res, err := Execute(uid, params)
	c.RPC(http.StatusOK, res, err)
}

func SearchRead(c *server.Context) {
	uid := c.Session().Get("uid").(int64)
	var params searchReadParams
	c.BindRPCParams(&params)
	res, err := searchRead(uid, params)
	c.RPC(http.StatusOK, res, err)
}
