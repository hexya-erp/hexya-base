// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	. "github.com/smartystreets/goconvey/convey"
)

func Test2ManyRelations(t *testing.T) {
	Convey("Testing 2many relations modification with client triplets", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Testing many2many '6' triplet", func() {
				adminGroup := pool.ResGroups().NewSet(env).Search(pool.ResGroups().GroupID().Equals(security.AdminGroupID))
				jsonGroupsData := fmt.Sprintf("[[6, 0, [%d]]]", adminGroup.ID())
				var groupsData interface{}
				json.Unmarshal([]byte(jsonGroupsData), &groupsData)

				env.Pool("ResUsers").Call("Create", models.FieldMap{
					"Name":   "Test User",
					"Login":  "test_user",
					"Groups": groupsData,
				})
				user := pool.ResUsers().NewSet(env).Search(pool.ResUsers().Login().Equals("test_user"))
				So(user.Groups().Len(), ShouldEqual, 1)
				So(user.Groups().ID(), ShouldEqual, adminGroup.ID())
			})
		})
	})
}
