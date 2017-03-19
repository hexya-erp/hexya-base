// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import "github.com/npiganeau/yep/yep/models"

func initCompany() {
	resCompany := models.NewModel("ResCompany")
	resCompany.AddCharField("Name", models.StringFieldParams{})
}
