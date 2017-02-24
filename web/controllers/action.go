// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"

	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/server"
)

func ActionLoad(c *server.Context) {
	params := struct {
		ActionID          string         `json:"action_id"`
		AdditionalContext *types.Context `json:"additional_context"`
	}{}
	c.BindRPCParams(&params)
	action := actions.Registry.GetById(params.ActionID)
	c.RPC(http.StatusOK, action)
}
