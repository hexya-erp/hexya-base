// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package web

import (
	"fmt"
	"strings"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	filterModel := pool.Filter().DeclareModel()
	filterModel.AddCharField("Name", models.StringFieldParams{String: "Filter Name", Required: true, Translate: true})
	filterModel.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: pool.User(), OnDelete: models.Cascade,
		Default: func(env models.Environment, maps models.FieldMap) interface{} {
			return pool.User().Search(env, pool.User().ID().Equals(env.Uid()))
		}, Help: `The user this filter is private to. When left empty the filter is public and available to all users.`})
	filterModel.AddTextField("Domain", models.StringFieldParams{Required: true, Default: models.DefaultValue("[]")})
	filterModel.AddTextField("Context", models.StringFieldParams{Required: true, Default: models.DefaultValue("{}")})
	filterModel.AddTextField("Sort", models.StringFieldParams{Required: true, Default: models.DefaultValue("[]")})
	filterModel.AddCharField("ResModel", models.StringFieldParams{String: "Model", Required: true, JSON: "model_id"})
	filterModel.AddBooleanField("IsDefault", models.SimpleFieldParams{String: "Default filter"})
	filterModel.AddCharField("Action", models.StringFieldParams{
		Help: `The menu action this filter applies to. When left empty the filter applies to all menus for this model.`,
		JSON: "action_id"})
	filterModel.AddBooleanField("Active", models.SimpleFieldParams{Default: models.DefaultValue(true)})

	filterModel.AddSQLConstraint("name_model_uid_unique", "unique (name, model_id, user_id, action_id)", "Filter names must be unique")

	filterModel.Methods().GetFilters().DeclareMethod(
		`GetFilters returns the filters for the given model and actionID for the current user`,
		func(rs pool.FilterSet, modelName, actionID string) []models.FieldMap {
			condition := pool.Filter().ResModel().Equals(modelName).
				And().Action().Equals(actionID).
				And().UserFilteredOn(pool.User().ID().Equals(rs.Env().Uid())).Or().User().IsNull()
			userContext := pool.User().Browse(rs.Env(), []int64{rs.Env().Uid()}).ContextGet()
			filterRS := pool.Filter().NewSet(rs.Env())
			res := filterRS.WithNewContext(userContext).Search(condition).Read([]string{"Name", "IsDefault", "Domain", "Context", "User", "Sort"})
			return res
		})

	filterModel.Methods().Copy().Extend("",
		func(rs pool.FilterSet, overrides models.FieldMapper, fieldsToUnset ...models.FieldNamer) pool.FilterSet {
			rs.EnsureOne()
			vals := rs.DataStruct(overrides.FieldMap(fieldsToUnset...))
			vals.Name = fmt.Sprintf("%s (copy)", rs.Name())
			return rs.Super().Copy(overrides, fieldsToUnset...)
		})

	filterModel.Methods().CreateOrReplace().DeclareMethod(
		`CreateOrReplace creates or updates the filter with the given parameters.
		Filter is considered the same if it has the same name (case insensitive) and the same user (if it has one).`,
		func(rs pool.FilterSet, vals models.FieldMapper) pool.FilterSet {
			fMap := vals.FieldMap()
			fMap["domain"] = strutils.MarshalToJSONString(fMap["domain"])
			fMap["domain"] = strings.Replace(fMap["domain"].(string), "false", "False", -1)
			fMap["domain"] = strings.Replace(fMap["domain"].(string), "true", "True", -1)
			fMap["context"] = strutils.MarshalToJSONString(fMap["context"])
			values := rs.DataStruct(fMap)
			currentFilters := rs.GetFilters(values.ResModel, values.Action)
			var matchingFilters []pool.FilterData
			for _, f := range currentFilters {
				filter := rs.DataStruct(f)
				if strings.ToLower(filter.Name) != strings.ToLower(values.Name) {
					continue
				}
				if !filter.User.Equals(values.User) {
					continue
				}
				matchingFilters = append(matchingFilters, *filter)
			}

			if values.IsDefault {
				if !values.User.IsEmpty() {
					// Setting new default: any other default that belongs to the user
					// should be turned off
					actionCondition := rs.GetActionCondition(values.Action)
					defaults := pool.Filter().Search(rs.Env(), actionCondition.
						And().ResModel().Equals(values.ResModel).
						And().User().Equals(values.User).
						And().IsDefault().Equals(true))
					if !defaults.IsEmpty() {
						defaults.SetIsDefault(false)
					}
				} else {
					rs.CheckGlobalDefault(vals, matchingFilters)
				}
			}
			if len(matchingFilters) > 0 {
				// When a filter exists for the same (name, model, user) triple, we simply
				// replace its definition (considering action_id irrelevant here)
				matchingFilter := pool.Filter().Browse(rs.Env(), []int64{matchingFilters[0].ID})
				matchingFilter.Write(values)
				return matchingFilter
			}
			return rs.Create(values)
		})

	filterModel.Methods().CheckGlobalDefault().DeclareMethod(
		`CheckGlobalDefault checks if there is a global default for the ResModel requested.

	       If there is, and the default is different than the record being written
	       (-> we're not updating the current global default), raise an error
	       to avoid users unknowingly overwriting existing global defaults (they
	       have to explicitly remove the current default before setting a new one)

	       This method should only be called if 'vals' is trying to set 'IsDefault'`,
		func(rs pool.FilterSet, vals models.FieldMapper, matchingFilters []pool.FilterData) {
			values := rs.DataStruct(vals.FieldMap())
			actionCondition := rs.GetActionCondition(values.Action)
			defaults := pool.Filter().Search(rs.Env(), actionCondition.
				And().ResModel().Equals(values.ResModel).
				And().User().IsNull().
				And().IsDefault().Equals(true))
			if defaults.IsEmpty() {
				return
			}
			if len(matchingFilters) > 0 && matchingFilters[0].ID == defaults.ID() {
				return
			}
			log.Panic("There is already a shared filter set as default for this model, delete or change it before setting a new default", "model", values.ResModel)
		})

	filterModel.Methods().GetActionCondition().DeclareMethod(
		`GetActionCondition returns a condition for matching filters that are visible in the
		same context (menu/view) as the given action.`,
		func(rs pool.FilterSet, action string) pool.FilterCondition {
			if action != "" {
				// filters specific to this menu + global ones
				return pool.Filter().Action().Equals(action).Or().Action().IsNull()
			}
			return pool.Filter().Action().IsNull()
		})
}
