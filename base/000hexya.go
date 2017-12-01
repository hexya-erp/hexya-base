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

package base

import (
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/hexya-erp/hexya/pool"
)

const (
	// MODULE_NAME is the name of this module
	MODULE_NAME string = "base"
)

var log *logging.Logger

func init() {
	log = logging.GetLogger("base")
	server.RegisterModule(&server.Module{
		Name: MODULE_NAME,
		PostInit: func() {
			err := models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {

				mainCompanyPartner := pool.Partner().Search(env, pool.Partner().ID().Equals(1))
				if mainCompanyPartner.IsEmpty() {
					log.Debug(mainCompanyPartner.T("Creating main company partner"))
					mainCompanyPartner = pool.Partner().Create(env, &pool.PartnerData{
						ID:        1,
						Name:      "Your Company",
						IsCompany: true,
						Customer:  false,
					})
					env.Cr().Execute("SELECT nextval('partner_id_seq')")
				}

				mainCompany := pool.Company().Search(env, pool.Company().ID().Equals(1))
				if mainCompany.IsEmpty() {
					log.Debug(mainCompany.T("Creating main company"))
					euro := pool.Currency().Search(env, pool.Currency().HexyaExternalID().Equals("base_EUR"))
					mainCompany = pool.Company().Create(env, &pool.CompanyData{
						ID:              1,
						Name:            mainCompanyPartner.Name(),
						Partner:         mainCompanyPartner,
						Currency:        euro,
						HexyaExternalID: "base_main_company",
					})
					env.Cr().Execute("SELECT nextval('company_id_seq')")
				}

				adminPartner := pool.Partner().Search(env, pool.Partner().ID().Equals(2))
				if adminPartner.IsEmpty() {
					log.Debug(adminPartner.T("Creating admin partner"))
					adminPartner = pool.Partner().Create(env, &pool.PartnerData{
						ID:       2,
						Lang:     "en_US",
						Name:     "Administrator",
						Function: "IT Manager",
						Customer: false,
					})
					env.Cr().Execute("SELECT nextval('partner_id_seq')")
				}

				avatarImg, _ := ioutil.ReadFile(filepath.Join(generate.HexyaDir, "hexya", "server", "static", "base", "src", "img", "avatar.png"))

				adminUser := pool.User().Search(env, pool.User().ID().Equals(security.SuperUserID))
				ActionID := actions.MakeActionRef("base_action_res_users")
				if adminUser.IsEmpty() {
					log.Debug(adminUser.T("Creating admin user"))
					pool.User().Create(env, &pool.UserData{
						ID:          security.SuperUserID,
						Name:        "Administrator",
						Active:      true,
						Company:     mainCompany,
						Companies:   mainCompany,
						Login:       "admin",
						LoginDate:   dates.DateTime{},
						Password:    "admin",
						Partner:     adminPartner,
						ActionID:    ActionID,
						ImageSmall:  base64.StdEncoding.EncodeToString(avatarImg),
						ImageMedium: base64.StdEncoding.EncodeToString(avatarImg),
						Image:       base64.StdEncoding.EncodeToString(avatarImg),
					})
					env.Cr().Execute("SELECT nextval('user_id_seq')")
				}

				pool.Group().NewSet(env).ReloadGroups()

				pool.ConfigParameter().NewSet(env).Init()
			})
			if err != nil {
				log.Panic("Error while initializing", "error", err)
			}
		},
	})
}
