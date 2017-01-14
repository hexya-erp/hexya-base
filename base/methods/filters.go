// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package methods

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
)

func initFilters() {
	models.CreateMethod("IrFilters", "GetFilters",
		`GetFilters returns the filters for the given model and actionID for the current user`,
		func(rs pool.IrFiltersSet, modelName, actionID string) []pool.IrFilters {
			res := rs.Filter("Model", "=", modelName).Filter("ActionID", "=", actionID).
				Filter("User.ID", "=", rs.Env().Uid()).All()
			return res
		})
}
