// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	_ "github.com/npiganeau/yep-base/base"
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/tests"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	tests.RunTests(m, "base")
}

func TestUserAuthentication(t *testing.T) {
	Convey("Testing User Authentication", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			userJohn := pool.ResUsers().NewSet(env).Create(&pool.ResUsersData{
				Name:     "John Smith",
				Login:    "jsmith",
				Password: "secret",
			})
			Convey("Correct user authentication", func() {
				uid, err := pool.ResUsers().NewSet(env).Authenticate("jsmith", "secret")
				So(uid, ShouldEqual, userJohn.ID())
				So(err, ShouldBeNil)
			})
			Convey("Invalid credentials authentication", func() {
				uid, err := pool.ResUsers().NewSet(env).Authenticate("jsmith", "wrong-secret")
				So(uid, ShouldEqual, 0)
				So(err, ShouldHaveSameTypeAs, security.InvalidCredentialsError(""))
			})
			Convey("Unknown user authentication", func() {
				uid, err := pool.ResUsers().NewSet(env).Authenticate("jsmith2", "wrong-secret")
				So(uid, ShouldEqual, 0)
				So(err, ShouldHaveSameTypeAs, security.UserNotFoundError(""))
			})
		})
	})
}
