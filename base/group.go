// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	group := pool.Group().DeclareModel()
	group.AddCharField("GroupID", models.StringFieldParams{Required: true})
	group.AddCharField("Name", models.StringFieldParams{Required: true, Translate: true})

	group.Methods().Create().Extend("",
		func(rs pool.GroupSet, data models.FieldMapper) pool.GroupSet {
			if rs.Env().Context().HasKey("GroupForceCreate") {
				return rs.Super().Create(data)
			}
			log.Panic(rs.T("Trying to create a security group"))
			panic("Unreachable")
		})

	group.Methods().Write().Extend("",
		func(rs pool.GroupSet, data models.FieldMapper, fieldsToUnset ...models.FieldNamer) bool {
			log.Panic(rs.T("Trying to modify a security group"))
			panic("Unreachable")
		})

	group.Methods().ReloadGroups().DeclareMethod(
		`ReloadGroups populates the Group table with groups from the security.Registry
		and refresh all memberships from the database to the security.Registry.`,
		func(rs pool.GroupSet) {
			log.Debug("Reloading groups")
			// Sync groups: registry => Database
			var existingGroupIds []string
			for _, group := range security.Registry.AllGroups() {
				existingGroupIds = append(existingGroupIds, group.ID)
				if !pool.Group().Search(rs.Env(), pool.Group().GroupID().Equals(group.ID)).IsEmpty() {
					// The group already exists in the database
					continue
				}
				rs.WithContext("GroupForceCreate", true).Create(&pool.GroupData{
					GroupID: group.ID,
					Name:    group.Name,
				})
			}
			// Remove unknown groups from database
			pool.Group().Search(rs.Env(), pool.Group().GroupID().NotIn(existingGroupIds)).Unlink()
			// Sync memberships: DB => Registry
			allUsers := pool.User().NewSet(rs.Env()).FetchAll()
			allUsers.AddMandatoryGroups()
			allUsers.SyncMemberships()
		})
}
