// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGroupLoading(t *testing.T) {
	models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		var (
			adminUser, user     pool.ResUsersSet
			adminGrp, someGroup pool.ResGroupsSet
		)
		Convey("Testing Group Loading", t, func() {
			pool.ResGroups().NewSet(env).ReloadGroups()
			groups := pool.ResGroups().NewSet(env).FetchAll()
			So(groups.Len(), ShouldEqual, len(security.Registry.AllGroups()))
		})
		Convey("Testing Group ReLoading with a new group", t, func() {
			security.Registry.NewGroup("some_group", "Some Group")
			pool.ResGroups().NewSet(env).ReloadGroups()
			groups := pool.ResGroups().NewSet(env).FetchAll()
			So(groups.Len(), ShouldEqual, len(security.Registry.AllGroups()))
		})
		Convey("Creating a new user with a new group", t, func() {
			adminUser = pool.ResUsers().NewSet(env).Search(pool.ResUsers().ID().Equals(security.SuperUserID))
			adminGrp = pool.ResGroups().NewSet(env).Search(pool.ResGroups().GroupID().Equals(security.AdminGroupID))
			someGroup = pool.ResGroups().NewSet(env).Search(pool.ResGroups().GroupID().Equals("some_group"))
			user = pool.ResUsers().NewSet(env).Create(&pool.ResUsersData{
				Name:   "Test User",
				Groups: someGroup,
			})
			So(user.Groups().Ids(), ShouldHaveLength, 1)
			So(user.Groups().Ids(), ShouldContain, someGroup.ID())
			So(adminUser.Groups().Ids(), ShouldHaveLength, 1)
			So(adminUser.Groups().Ids(), ShouldContain, adminGrp.ID())
		})
		Convey("Giving the new user admin rights", t, func() {
			groups := someGroup.Union(adminGrp)
			user.SetGroups(groups)
			So(user.Groups().Ids(), ShouldHaveLength, 2)
			So(user.Groups().Ids(), ShouldContain, someGroup.ID())
			So(user.Groups().Ids(), ShouldContain, adminGrp.ID())
			So(adminUser.Groups().Ids(), ShouldHaveLength, 1)
			So(adminUser.Groups().Ids(), ShouldContain, adminGrp.ID())
		})
		Convey("Removing rights and checking that after reload, we get admin right", t, func() {
			user.SetGroups(pool.ResGroups().NewSet(env))
			adminUser = pool.ResUsers().NewSet(env).Search(pool.ResUsers().ID().Equals(security.SuperUserID))
			adminUser.SetGroups(pool.ResGroups().NewSet(env))
			So(user.Groups().Ids(), ShouldBeEmpty)
			So(adminUser.Groups().Ids(), ShouldBeEmpty)
			pool.ResGroups().NewSet(env).ReloadGroups()
			// We need to reload admin group because it changed id
			adminGrp = pool.ResGroups().NewSet(env).Search(pool.ResGroups().GroupID().Equals(security.AdminGroupID))
			So(user.Groups().Ids(), ShouldBeEmpty)
			So(adminUser.Groups().Ids(), ShouldHaveLength, 1)
			So(adminUser.Groups().Ids(), ShouldContain, adminGrp.ID())
		})
	})
}
