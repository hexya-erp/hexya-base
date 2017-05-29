// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/pool"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGroupLoading(t *testing.T) {
	models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		var (
			adminUser, user                    pool.UserSet
			adminGrp, someGroup, everyoneGroup pool.GroupSet
		)
		Convey("Testing Group Loading", t, func() {
			pool.Group().NewSet(env).ReloadGroups()
			groups := pool.Group().NewSet(env).FetchAll()
			So(groups.Len(), ShouldEqual, len(security.Registry.AllGroups()))
			adminUser = pool.User().Search(env, pool.User().ID().Equals(security.SuperUserID))
			adminGrp = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupAdminID))
			everyoneGroup = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupEveryoneID))
			So(adminUser.Groups().Ids(), ShouldHaveLength, 2)
			So(adminUser.Groups().Ids(), ShouldContain, adminGrp.ID())
			So(adminUser.Groups().Ids(), ShouldContain, everyoneGroup.ID())
		})
		Convey("Testing Group ReLoading with a new group", t, func() {
			security.Registry.NewGroup("some_group", "Some Group")
			pool.Group().NewSet(env).ReloadGroups()
			groups := pool.Group().NewSet(env).FetchAll()
			So(groups.Len(), ShouldEqual, len(security.Registry.AllGroups()))
		})
		Convey("Creating a new user with a new group", t, func() {
			adminUser = pool.User().Search(env, pool.User().ID().Equals(security.SuperUserID))
			adminGrp = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupAdminID))
			everyoneGroup = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupEveryoneID))
			someGroup = pool.Group().Search(env, pool.Group().GroupID().Equals("some_group"))
			user = pool.User().Create(env, &pool.UserData{
				Name:   "Test User",
				Login:  "test_user",
				Groups: someGroup,
			})
			So(user.Groups().Ids(), ShouldHaveLength, 1)
			So(user.Groups().Ids(), ShouldContain, someGroup.ID())
			So(adminUser.Groups().Ids(), ShouldHaveLength, 2)
			So(adminUser.Groups().Ids(), ShouldContain, adminGrp.ID())
			So(adminUser.Groups().Ids(), ShouldContain, everyoneGroup.ID())
		})
		Convey("Giving the new user admin rights", t, func() {
			groups := someGroup.Union(adminGrp)
			user.SetGroups(groups)
			pool.Group().NewSet(env).ReloadGroups()
			adminGrp = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupAdminID))
			everyoneGroup = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupEveryoneID))
			someGroup = pool.Group().Search(env, pool.Group().GroupID().Equals("some_group"))
			So(user.Groups().Ids(), ShouldHaveLength, 3)
			So(user.Groups().Ids(), ShouldContain, someGroup.ID())
			So(user.Groups().Ids(), ShouldContain, adminGrp.ID())
			So(user.Groups().Ids(), ShouldContain, everyoneGroup.ID())
			So(adminUser.Groups().Ids(), ShouldHaveLength, 2)
			So(adminUser.Groups().Ids(), ShouldContain, adminGrp.ID())
			So(adminUser.Groups().Ids(), ShouldContain, everyoneGroup.ID())
		})
		Convey("Removing rights and checking that after reload, we get admin right", t, func() {
			user = pool.User().Search(env, pool.User().Login().Equals("test_user"))
			user.SetGroups(pool.Group().NewSet(env))
			adminUser = pool.User().Search(env, pool.User().ID().Equals(security.SuperUserID))
			adminUser.SetGroups(pool.Group().NewSet(env))
			So(user.Groups().Ids(), ShouldBeEmpty)
			So(adminUser.Groups().Ids(), ShouldBeEmpty)
			pool.Group().NewSet(env).ReloadGroups()
			// We need to reload admin and everyone groups because they changed id
			adminGrp = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupAdminID))
			everyoneGroup = pool.Group().Search(env, pool.Group().GroupID().Equals(security.GroupEveryoneID))
			So(user.Groups().Ids(), ShouldHaveLength, 1)
			So(user.Groups().Ids(), ShouldContain, everyoneGroup.ID())
			So(adminUser.Groups().Ids(), ShouldHaveLength, 2)
			So(adminUser.Groups().Ids(), ShouldContain, adminGrp.ID())
			So(adminUser.Groups().Ids(), ShouldContain, everyoneGroup.ID())
		})
	})
}
