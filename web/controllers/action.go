// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/server"
)

// ActionLoad returns the action with the given id
func ActionLoad(c *server.Context) {
	params := struct {
		ActionID          string         `json:"action_id"`
		AdditionalContext *types.Context `json:"additional_context"`
	}{}
	c.BindRPCParams(&params)
	action := actions.Registry.MustGetById(params.ActionID)
	c.RPC(http.StatusOK, action)
}

// ActionRun runs the given server action
func ActionRun(c *server.Context) {
	params := struct {
		ActionID string         `json:"action_id"`
		Context  *types.Context `json:"context"`
	}{}
	c.BindRPCParams(&params)
	action := actions.Registry.MustGetById(params.ActionID)

	// Process context ids into args
	var ids []int64
	if params.Context.Get("active_ids") != nil {
		ids = params.Context.Get("active_ids").([]int64)
	} else if params.Context.Get("active_id") != nil {
		ids = []int64{params.Context.Get("active_id").(int64)}
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		log.Panic("Unable to marshal ids")
	}

	// Process context into kwargs
	contextJSON, _ := json.Marshal(params.Context)
	kwargs := make(map[string]json.RawMessage)
	kwargs["context"] = contextJSON

	// Execute the function
	resAction, _ := Execute(c.Session().Get("uid").(int64), CallParams{
		Model:  action.Model,
		Method: action.Method,
		Args:   []json.RawMessage{idsJSON},
		KWArgs: kwargs,
	})

	if _, ok := resAction.(*actions.BaseAction); ok {
		c.RPC(http.StatusOK, resAction)
	} else {
		c.RPC(http.StatusOK, false)
	}
}
