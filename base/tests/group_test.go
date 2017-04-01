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
		Convey("Testing Group Loading", t, func() {
			pool.ResGroups().NewSet(env).ReloadGroups()
			groups := pool.ResGroups().NewSet(env).Load()
			So(groups.Len(), ShouldEqual, len(security.Registry.AllGroups()))
		})
		Convey("Testing Group ReLoading", t, func() {
			pool.ResGroups().NewSet(env).ReloadGroups()
			groups := pool.ResGroups().NewSet(env).Load()
			So(groups.Len(), ShouldEqual, len(security.Registry.AllGroups()))
		})
	})
}
