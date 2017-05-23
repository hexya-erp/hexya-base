// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/models"
)

func initFilters() {
	irFilters := models.NewModel("IrFilters")
	irFilters.AddCharField("ResModel", models.StringFieldParams{})
	irFilters.AddCharField("Domain", models.StringFieldParams{})
	irFilters.AddCharField("Context", models.StringFieldParams{})
	irFilters.AddCharField("Name", models.StringFieldParams{})
	irFilters.AddBooleanField("IsDefault", models.SimpleFieldParams{})
	irFilters.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: "ResUsers"})
	irFilters.AddCharField("ActionID", models.StringFieldParams{GoType: new(actions.ActionRef)})

	irFilters.AddMethod("GetFilters",
		`GetFilters returns the filters for the given model and actionID for the current user`,
		func(rs pool.IrFiltersSet, modelName, actionID string) []pool.IrFiltersData {
			condition := pool.IrFilters().ResModel().Equals(modelName).
				And().ActionID().Equals(actions.MakeActionRef(actionID)).
				And().UserFilteredOn(pool.ResUsers().ID().Equals(rs.Env().Uid()))
			res := rs.Search(condition).All()
			return res
		})
}
