// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	models.NewModel("Bank")
	pool.Bank().AddCharField("Name", models.StringFieldParams{Required: true})
	pool.Bank().AddCharField("Street", models.StringFieldParams{})
	pool.Bank().AddCharField("Street2", models.StringFieldParams{})
	pool.Bank().AddCharField("Zip", models.StringFieldParams{})
	pool.Bank().AddCharField("City", models.StringFieldParams{})
	pool.Bank().AddMany2OneField("State", models.ForeignKeyFieldParams{RelationModel: "CountryState", String: "Fed. State"}) // domain="[('country_id', '=', country)]"
	pool.Bank().AddMany2OneField("Country", models.ForeignKeyFieldParams{RelationModel: "Country"})
	pool.Bank().AddCharField("Email", models.StringFieldParams{})
	pool.Bank().AddCharField("Phone", models.StringFieldParams{})
	pool.Bank().AddCharField("Fax", models.StringFieldParams{})
	pool.Bank().AddBooleanField("Active", models.SimpleFieldParams{Default: models.DefaultValue(true)})
	pool.Bank().AddCharField("BIC", models.StringFieldParams{String: "Bank Identifier Cord", Index: true, Help: "Sometimes called BIC or Swift."})

	pool.Bank().Methods().NameGet().Extend("",
		func(rs pool.BankSet) string {
			res := rs.Name()
			if rs.BIC() != "" {
				res = fmt.Sprintf("%s - %s", res, rs.BIC())
			}
			return res
		})

}
