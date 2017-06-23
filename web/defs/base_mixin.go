// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	baseMixin := pool.BaseMixin()
	baseMixin.AddBooleanField("Active", models.SimpleFieldParams{})

	baseMixin.Methods().ToggleActive().DeclareMethod(
		`ToggleActive toggles the Active field of this object`,
		func(rs pool.BaseMixinSet) {
			if rs.Active() {
				rs.SetActive(false)
			} else {
				rs.SetActive(true)
			}
		})
}
