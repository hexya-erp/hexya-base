// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	pool.Lang().DeclareModel()
	pool.Lang().AddCharField("Name", models.StringFieldParams{Required: true, Unique: true})
	pool.Lang().AddCharField("Code", models.StringFieldParams{String: "Locale Code", Required: true,
		Help: "This field is used to set/get locales for user", Unique: true})
	pool.Lang().AddCharField("ISOCode", models.StringFieldParams{Help: "This ISO code is the name of PO files to use for translations"})
	pool.Lang().AddBooleanField("Translatable", models.SimpleFieldParams{})
	pool.Lang().AddBooleanField("Active", models.SimpleFieldParams{})
	pool.Lang().AddSelectionField("Direction", models.SelectionFieldParams{Selection: types.Selection{"ltr": "Left-to-Right", "rtl": "Right-to-left"},
		Required: true, Default: models.DefaultValue("ltr")})
	pool.Lang().AddCharField("DateFormat", models.StringFieldParams{Required: true, Default: models.DefaultValue("2006-01-02")})
	pool.Lang().AddCharField("TimeFormat", models.StringFieldParams{Required: true, Default: models.DefaultValue("15:04:05")})
	pool.Lang().AddCharField("Grouping", models.StringFieldParams{String: "Separator Format", Required: true,
		Default: models.DefaultValue("[]"), Help: `The Separator Format should be like [,n] where 0 < n :starting from Unit digit."
-1 will end the separation. e.g. [3,2,-1] will represent 106500 to be 1,06,500"
[1,2,-1] will represent it to be 106,50,0;[3] will represent it as 106,500."
Provided ',' as the thousand separator in each case.`})
	pool.Lang().AddCharField("DecimalPoint", models.StringFieldParams{String: "Decimal Separator", Required: true, Default: models.DefaultValue(".")})
	pool.Lang().AddCharField("ThousandsSep", models.StringFieldParams{String: "Thousands Separator", Default: models.DefaultValue(",")})
}
