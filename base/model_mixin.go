// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"encoding/json"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
)

func init() {

	h.ModelMixin().Methods().ToggleActive().DeclareMethod(
		`ToggleActive toggles the Active field of this object if it exists.`,
		func(rs h.BaseMixinSet) {
			_, exists := rs.Model().Fields().Get("active")
			if !exists {
				return
			}
			if rs.Get("Active").(bool) {
				rs.Set("Active", false)
			} else {
				rs.Set("Active", true)
			}
		})

	h.ModelMixin().Methods().Search().Extend("",
		func(rs h.ModelMixinSet, cond q.ModelMixinCondition) h.ModelMixinSet {
			activeField, exists := rs.Model().Fields().Get("active")
			activeTest := !rs.Env().Context().HasKey("active_test") || rs.Env().Context().GetBool("active_test")
			if !exists || !activeTest || cond.HasField(activeField) {
				return rs.Super().Search(cond)
			}
			activeCond := q.ModelMixinCondition{
				Condition: models.Registry.MustGet(rs.ModelName()).Field("active").Equals(true),
			}
			cond = cond.AndCond(activeCond)
			return rs.Super().Search(cond)
		})

	h.ModelMixin().Methods().SearchAll().Extend("",
		func(rs h.ModelMixinSet) h.ModelMixinSet {
			_, exists := rs.Model().Fields().Get("active")
			activeTest := !rs.Env().Context().HasKey("active_test") || !rs.Env().Context().GetBool("active_test")
			if !exists || !activeTest {
				return rs.Super().SearchAll()
			}
			activeCond := q.ModelMixinCondition{
				Condition: models.Registry.MustGet(rs.ModelName()).Field("active").Equals(true),
			}
			return rs.Search(activeCond)
		})

	h.CommonMixin().Methods().WithWorker().DeclareMethod(
		`WithWorker sets the JobsArgs' Worker to a worker corresponding to the given name. ('Main' by default)`,
		func(rs h.CommonMixinSet, workerName string) h.JobArgsSet {
			out := h.JobArgs().NewSet(rs.Env()).Create(&h.JobArgsData{})
			out.SetModelName(rs.Collection().ModelName())
			out.SetWorkerName(workerName)
			return out
		})

	h.CommonMixin().Methods().WithParams().DeclareMethod(
		`WithParams gives the current JobArg some parameters. (no parameters by default)`,
		func(rs h.CommonMixinSet, params ...interface{}) h.JobArgsSet {
			out := h.JobArgs().NewSet(rs.Env()).Create(&h.JobArgsData{})
			out.SetModelName(rs.Collection().ModelName())
			json, _ := json.Marshal(params)
			out.SetParams(string(json))
			return out
		})

	h.CommonMixin().Methods().Enqueue().DeclareMethod(
		`Enqueue creates and pushes to queue the given JobArg created using WithParams and WithWorker or a default one`,
		func(rs h.CommonMixinSet, method models.Methoder) {
			out := h.JobArgs().NewSet(rs.Env()).Create(&h.JobArgsData{})
			out.SetModelName(rs.Collection().ModelName())
			out.Enqueue(method)
		})

	h.CommonMixin().Methods().PollCancel().DeclareMethod(
		`PollCancel panics if rs was called to be canceled. Meant to be used in methods called by a Worker`,
		func(rs h.CommonMixinSet) {
			h.Worker().NewSet(rs.Env()).PollCancel()
		})
}
