// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import "github.com/npiganeau/yep/yep/models"

func initBaseMixin() {
	baseMixin := models.Registry.MustGet("BaseMixin")
	baseMixin.AddBooleanField("Active", models.SimpleFieldParams{})
}
