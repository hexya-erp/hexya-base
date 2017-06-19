// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	countryGroup := pool.CountryGroup().DeclareModel()
	countryGroup.AddCharField("Name", models.StringFieldParams{Required: true})
	countryGroup.AddMany2ManyField("Countries", models.Many2ManyFieldParams{RelationModel: pool.Country()})

	countryState := pool.CountryState().DeclareModel()
	countryState.AddCharField("Name", models.StringFieldParams{String: "State Name", Required: true,
		Help: "Administrative divisions of a country. E.g. Fed. State, Departement, Canton"})
	countryState.AddMany2OneField("Country", models.ForeignKeyFieldParams{RelationModel: pool.Country(), Required: true})
	countryState.AddCharField("Code", models.StringFieldParams{String: "State Code", Size: 3,
		Help: "The state code in max. three chars.", Required: true})

	country := pool.Country().DeclareModel()
	country.AddCharField("Name", models.StringFieldParams{String: "Country Name", Help: "The full name of the country.",
		Translate: true, Required: true, Unique: true})
	country.AddCharField("Code", models.StringFieldParams{String: "Country Code", Size: 2, Unique: true,
		Help: "The ISO country code in two chars.\nYou can use this field for quick search."})
	country.AddTextField("AddressFormat", models.StringFieldParams{Default: func(env models.Environment, fMap models.FieldMap) interface{} {
		return "%(Street)s\n%(Street2)s\n%(City)s %(StateCode)s %(Zip)s\n%(CountryName)s"
	}, Help: "You can state here the usual format to use for the addresses belonging to this country."})
	country.AddMany2OneField("Currency", models.ForeignKeyFieldParams{RelationModel: pool.Currency()})
	country.AddBinaryField("Image", models.SimpleFieldParams{})
	country.AddIntegerField("PhoneCode", models.SimpleFieldParams{String: "Country Calling Code"})
	country.AddMany2ManyField("CountryGroups", models.Many2ManyFieldParams{RelationModel: pool.CountryGroup()})
	country.AddOne2ManyField("States", models.ReverseFieldParams{RelationModel: pool.CountryState(), ReverseFK: "Country"})
}
