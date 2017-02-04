// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/models"
)

func initFilters() {
	models.NewModel("IrFilters", new(struct {
		ID        int64
		ResModel  string
		Domain    string
		Context   string
		Name      string
		IsDefault bool
		User      pool.ResUsersSet `yep:"type(many2one)"`
		ActionID  actions.ActionRef
	}))
}
