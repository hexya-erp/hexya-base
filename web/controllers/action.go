// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/npiganeau/yep/yep/ir"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/server"
)

func ActionLoad(c *gin.Context) {
	params := struct {
		ActionID          string         `json:"action_id"`
		AdditionalContext *types.Context `json:"additional_context"`
	}{}
	server.BindRPCParams(c, &params)
	action := ir.ActionsRegistry.GetActionById(params.ActionID)
	server.RPC(c, http.StatusOK, action)
}
