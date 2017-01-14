// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/npiganeau/yep/yep/server"
)

func MenuLoadNeedaction(c *gin.Context) {
	type lnaParams struct {
		MenuIds []string `json:"menu_ids"`
	}
	var params lnaParams
	server.BindRPCParams(c, &params)

	// TODO: update with real needaction support
	type lnaResponse struct {
		NeedactionEnabled bool `json:"needaction_enabled"`
		NeedactionCounter int  `json:"needaction_counter"`
	}
	res := make(map[string]lnaResponse)
	for _, menu := range params.MenuIds {
		res[menu] = lnaResponse{}
	}
	server.RPC(c, http.StatusOK, res)
}
