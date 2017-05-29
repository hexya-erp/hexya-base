// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func initCompany() {
	models.NewModel("Company")
	company := pool.Company()
	company.AddCharField("Name", models.StringFieldParams{})
}
