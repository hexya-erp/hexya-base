// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import "github.com/npiganeau/yep/yep/models"

type ResCompany struct {
	ID   int64
	Name string
}

func initCompany() {
	models.CreateModel("ResCompany")
	models.ExtendModel("ResCompany", new(ResCompany))
}
