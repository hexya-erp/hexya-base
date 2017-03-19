// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package methods

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/actions"
)

func initFilters() {
	pool.IrFilters().AddMethod("GetFilters",
		`GetFilters returns the filters for the given model and actionID for the current user`,
		func(rs pool.IrFiltersSet, modelName, actionID string) []pool.IrFiltersData {
			condition := pool.IrFilters().ResModel().Equals(modelName).
				And().ActionID().Equals(actions.MakeActionRef(actionID)).
				And().UserFilteredOn(pool.ResUsers().ID().Equals(rs.Env().Uid()))
			res := rs.Search(condition).All()
			return res
		})
}
