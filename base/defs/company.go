// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
)

func initCompany() {
	models.NewModel("Company")
	company := pool.Company()
	company.AddCharField("Name", models.StringFieldParams{})
}
