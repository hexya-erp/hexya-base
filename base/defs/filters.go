// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/ir"
	"github.com/npiganeau/yep/yep/models"
)

type IrFilters struct {
	ID        int64
	Model     string
	Domain    string
	Context   string
	Name      string
	IsDefault bool
	User      pool.ResUsersSet `yep:"type(many2one)"`
	ActionID  ir.ActionRef
}

func initFilters() {
	models.CreateModel("IrFilters")
	models.ExtendModel("IrFilters", new(IrFilters))
}
