// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/types"
)

func initCurrency() {
	resCurrencyRate := models.NewModel("ResCurrencyRate")
	resCurrencyRate.AddDateTimeField("Name", models.SimpleFieldParams{String: "Date", Required: true, Index: true})
	resCurrencyRate.AddFloatField("Rate", models.FloatFieldParams{Digits: types.Digits{Precision: 12, Scale: 6}, Help: "The rate of the currency to the currency of rate 1"})
	resCurrencyRate.AddMany2OneField("Currency", models.ForeignKeyFieldParams{RelationModel: "ResCurrency"})
	resCurrencyRate.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: "ResCompany"})

	resCurrency := models.NewModel("ResCurrency")
	resCurrency.AddCharField("Name", models.StringFieldParams{String: "Currency", Help: "Currency Code [ISO 4217]", Size: 3, Unique: true})
	resCurrency.AddCharField("Symbol", models.StringFieldParams{Help: "urrency sign, to be used when printing amounts", Size: 4})
	resCurrency.AddFloatField("Rate", models.FloatFieldParams{String: "Current Rate", Help: "The rate of the currency to the currency of rate 1", Digits: types.Digits{Precision: 12, Scale: 6}}) //, Compute: "ComputeCurrentRate"})
	resCurrency.AddOne2ManyField("Rates", models.ReverseFieldParams{RelationModel: "ResCurrencyRate", ReverseFK: "Currency"})
	resCurrency.AddFloatField("Rounding", models.FloatFieldParams{String: "Rounding Factor", Digits: types.Digits{Precision: 12, Scale: 6}})
	resCurrency.AddIntegerField("DecimalPlaces", models.SimpleFieldParams{}) //Compute: "ComputeDecimalPlaces"})
	resCurrency.AddBooleanField("Active", models.SimpleFieldParams{})
	resCurrency.AddSelectionField("Position", models.SelectionFieldParams{Selection: models.Selection{"after": "After Amount", "before": "Before Amount"}, String: "Symbol Position", Help: "Determines where the currency symbol should be placed after or before the amount."})
}
