// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	filter := pool.Filter().DeclareModel()
	filter.AddCharField("ResModel", models.StringFieldParams{})
	filter.AddCharField("Domain", models.StringFieldParams{})
	filter.AddCharField("Context", models.StringFieldParams{})
	filter.AddCharField("Name", models.StringFieldParams{})
	filter.AddCharField("Sort", models.StringFieldParams{})
	filter.AddBooleanField("IsDefault", models.SimpleFieldParams{})
	filter.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: pool.User()})
	filter.AddCharField("ActionID", models.StringFieldParams{GoType: new(actions.ActionRef)})

	filter.Methods().GetFilters().DeclareMethod(
		`GetFilters returns the filters for the given model and actionID for the current user`,
		func(rs pool.FilterSet, modelName, actionID string) []pool.FilterData {
			condition := pool.Filter().ResModel().Equals(modelName).
				And().ActionID().Equals(actions.MakeActionRef(actionID)).
				And().UserFilteredOn(pool.User().ID().Equals(rs.Env().Uid()))
			res := rs.Search(condition).All()
			return res
		})
}
