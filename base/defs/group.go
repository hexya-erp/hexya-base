// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
)

func initGroups() {
	models.NewModel("ResGroups")
	resGroups := pool.ResGroups()
	resGroups.AddCharField("GroupID", models.StringFieldParams{Required: true})
	resGroups.AddCharField("Name", models.StringFieldParams{Required: true, Translate: true})

	resGroups.Methods().Create().Extend("",
		func(rs pool.ResGroupsSet, data *pool.ResGroupsData) pool.ResGroupsSet {
			if rs.Env().Context().HasKey("GroupForceCreate") {
				return rs.Super().Create(data)
			}
			log.Panic("Trying to create a security group")
			panic("Unreachable")
		})

	resGroups.Methods().Write().Extend("",
		func(rs pool.ResGroupsSet, data *pool.ResGroupsData, fieldsToUnset ...models.FieldNamer) {
			log.Panic("Trying to modify a security group")
		})

	resGroups.AddMethod("ReloadGroups",
		`ReloadGroups populates the ResGroups table with groups from the security.Registry
		and refresh all memberships.`,
		func(rs pool.ResGroupsSet) {
			log.Debug("Reloading groups")
			// Sync groups
			pool.ResGroups().NewSet(rs.Env()).FetchAll().Unlink()
			for _, group := range security.Registry.AllGroups() {
				rs.WithContext("GroupForceCreate", true).Create(&pool.ResGroupsData{
					GroupID: group.ID,
					Name:    group.Name,
				})
			}
			// Sync memberships
			for _, user := range pool.ResUsers().NewSet(rs.Env()).FetchAll().Records() {
				secGroups := security.Registry.UserGroups(user.ID())
				grpIds := make([]string, len(secGroups))
				i := 0
				for grp := range secGroups {
					grpIds[i] = grp.ID
					i++
				}
				groups := pool.ResGroups().Search(rs.Env(), pool.ResGroups().GroupID().In(grpIds))
				user.SetGroups(groups)
			}
		})
}
