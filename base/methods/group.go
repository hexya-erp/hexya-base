// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package methods

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/tools/logging"
)

func initGroups() {

	resGroups := pool.ResGroups()
	resGroups.ExtendMethod("Create", "",
		func(rs pool.ResGroupsSet, data *pool.ResGroupsData) pool.ResGroupsSet {
			if rs.Env().Context().Get("GroupForceCreate").(bool) {
				return rs.Super().Create(data)
			}
			logging.LogAndPanic(log, "Trying to create a security group")
			panic("Unreachable")
		})

	resGroups.ExtendMethod("Write", "",
		func(rs pool.ResGroupsSet, data *pool.ResGroupsData, fieldsToUnset ...models.FieldNamer) {
			logging.LogAndPanic(log, "Trying to modify a security group")
		})

	resGroups.AddMethod("ReloadGroups",
		`ReloadGroups populates the ResGroups table with groups from the security.Registry.`,
		func(rs pool.ResGroupsSet) {
			log.Debug("Reloading groups")
			pool.ResGroups().NewSet(rs.Env()).Search(pool.ResGroups().All()).Unlink()
			for _, group := range security.Registry.AllGroups() {
				rs.WithContext("GroupForceCreate", true).Create(&pool.ResGroupsData{
					GroupID: group.ID,
					Name:    group.Name,
				})
			}
		})

}
