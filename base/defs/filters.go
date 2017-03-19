// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
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
}
