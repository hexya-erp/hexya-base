// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package domains

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/tests"
	_ "github.com/hexya-erp/hexya/hexya/tests/testllmodule"
)

func TestMain(m *testing.M) {
	tests.RunTests(m, "domains")
}

func TestDomains(t *testing.T) {
	Convey("Testing Domains", t, func() {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("Creating users", func() {
				userJohnData := models.FieldMap{
					"Name":    "John Smith",
					"Email":   "jsmith@example.com",
					"IsStaff": true,
					"Nums":    1,
				}
				env.Pool("User").Call("Create", userJohnData)

				userJaneProfileData := models.FieldMap{
					"Age":     24,
					"Money":   12345,
					"Street":  "165 5th Avenue",
					"City":    "New York",
					"Zip":     "0305",
					"Country": "USA",
				}
				profile := env.Pool("Profile").Call("Create", userJaneProfileData).(models.RecordCollection)
				userJaneData := models.FieldMap{
					"Name":    "Jane Smith",
					"Email":   "jane.smith@example.com",
					"Profile": profile,
					"Nums":    2,
				}
				env.Pool("User").Call("Create", userJaneData)

				userWillData := models.FieldMap{
					"Name":    "Will Smith",
					"Email":   "will.smith@example.com",
					"IsStaff": true,
					"Nums":    3,
				}
				env.Pool("User").Call("Create", userWillData)

				martinProfile := env.Pool("Profile").Call("Create", models.FieldMap{"Age": 45})
				userData := models.FieldMap{
					"Name":    "Martin Weston",
					"Email":   "mweston@example.com",
					"Profile": martinProfile,
				}
				user := env.Pool("User").Call("Create", userData).(models.RecordCollection)
				So(user.Get("Profile").(models.RecordCollection).Get("Age"), ShouldEqual, 45)
			})
			Convey("Testing simple [(A), (B)] domain", func() {
				dom1 := []interface{}{
					0: []interface{}{"Name", "like", "Smith"},
					1: []interface{}{"Age", "=", 24},
				}
				dom1Users := env.Pool("User").Search(ParseDomain(dom1))
				So(dom1Users.Len(), ShouldEqual, 1)
				So(dom1Users.Get("Name"), ShouldEqual, "Jane Smith")
			})
			Convey("Testing ['|', (A), (B)] domain", func() {
				dom2 := []interface{}{
					0: "|",
					1: []interface{}{"Name", "like", "Will"},
					2: []interface{}{"Email", "ilike", "Jane.Smith"},
				}
				dom2Users := env.Pool("User").Search(ParseDomain(dom2)).OrderBy("Name")
				So(dom2Users.Len(), ShouldEqual, 2)
				userRecs := dom2Users.Records()
				So(userRecs[0].Get("Name"), ShouldEqual, "Jane Smith")
				So(userRecs[1].Get("Name"), ShouldEqual, "Will Smith")
			})
			Convey("Testing ['|', (A), '&' , (B), (C), (D)] domain", func() {
				dom3 := []interface{}{
					0: "|",
					1: []interface{}{"Name", "like", "Will"},
					2: "&",
					3: []interface{}{"Age", ">", 0},
					4: []interface{}{"Age", "<", 25},
					5: []interface{}{"Email", "not like", "will.smith"},
				}
				dom3Users := env.Pool("User").Search(ParseDomain(dom3)).OrderBy("Name")
				So(dom3Users.Len(), ShouldEqual, 1)
				So(dom3Users.Get("Name"), ShouldEqual, "Jane Smith")
			})
		})
	})
}
