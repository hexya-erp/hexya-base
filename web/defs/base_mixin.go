// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
)

func initBaseMixin() {
	baseMixin := models.Registry.MustGet("BaseMixin")
	baseMixin.AddBooleanField("Active", models.SimpleFieldParams{})

	baseMixin.AddMethod("ToggleActive",
		`ToggleActive toggles the Active field of this object`,
		func(rs pool.BaseMixinSet) {
			if rs.Active() {
				rs.SetActive(false)
			} else {
				rs.SetActive(true)
			}
		})

}
