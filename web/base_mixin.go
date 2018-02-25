// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool/h"
)

func init() {
	baseMixin := h.BaseMixin()

	baseMixin.Methods().ToggleActive().DeclareMethod(
		`ToggleActive toggles the Active field of this object`,
		func(rs h.BaseMixinSet) {
			fInfos := rs.FieldsGet(models.FieldsGetArgs{})
			for fName := range fInfos {
				if fName != "active" {
					continue
				}
				if rs.Get("Active").(bool) {
					rs.Set("Active", false)
				} else {
					rs.Set("Active", true)
				}
				return
			}
		})
}
