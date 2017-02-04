// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/models"
)

func initUsers() {
	models.NewModel("ResUsers", new(struct {
		ID          int64
		LoginDate   models.DateTime
		Partner     pool.ResPartnerSet `yep:"type(many2one);embed"`
		Name        string
		Login       string
		Password    string
		NewPassword string
		Signature   string
		Active      bool
		ActionID    actions.ActionRef
		//GroupIds []*ir.Group `yep:"json(groups_id)"`
		Company    pool.ResCompanySet `yep:"type(many2one)"`
		CompanyIds pool.ResCompanySet `yep:"json(company_ids);type(many2many)"`
		ImageSmall string
	}))
}
