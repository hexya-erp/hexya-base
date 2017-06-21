// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"fmt"

	"regexp"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	pool.Bank().DeclareModel()
	pool.Bank().AddCharField("Name", models.StringFieldParams{Required: true})
	pool.Bank().AddCharField("Street", models.StringFieldParams{})
	pool.Bank().AddCharField("Street2", models.StringFieldParams{})
	pool.Bank().AddCharField("Zip", models.StringFieldParams{})
	pool.Bank().AddCharField("City", models.StringFieldParams{})
	pool.Bank().AddMany2OneField("State", models.ForeignKeyFieldParams{RelationModel: pool.CountryState(), String: "Fed. State"}) // domain="[('country_id', '=', country)]"
	pool.Bank().AddMany2OneField("Country", models.ForeignKeyFieldParams{RelationModel: pool.Country()})
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

	pool.BankAccount().DeclareModel()
	pool.BankAccount().AddCharField("AccountType", models.StringFieldParams{Compute: "ComputeAccountType"})
	pool.BankAccount().AddCharField("Name", models.StringFieldParams{String: "Account Number", Required: true})
	pool.BankAccount().AddCharField("SanitizedAccountNumber", models.StringFieldParams{Compute: "ComputeSanitizedAccountNumber",
		Stored: true})
	pool.BankAccount().AddMany2OneField("Partner", models.ForeignKeyFieldParams{RelationModel: pool.Partner(),
		String: "Account Holder", OnDelete: models.Cascade, Index: true}) //domain=['|', ('is_company', '=', True), ('parent_id', '=', False)])
	pool.BankAccount().AddMany2OneField("Bank", models.ForeignKeyFieldParams{RelationModel: pool.Bank()})
	pool.BankAccount().AddCharField("BankName", models.StringFieldParams{Related: "Bank.Name"})
	pool.BankAccount().AddCharField("BankBIC", models.StringFieldParams{Related: "Bank.BIC"})
	pool.BankAccount().AddIntegerField("Sequence", models.SimpleFieldParams{})
	pool.BankAccount().AddMany2OneField("Currency", models.ForeignKeyFieldParams{RelationModel: pool.Currency()})
	pool.BankAccount().AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: pool.Company()})

	pool.BankAccount().Methods().ComputeAccountType().DeclareMethod(
		`ComputeAccountType computes the type of account from the account number`,
		func(rs pool.BankAccountSet) (*pool.BankAccountData, []models.FieldNamer) {
			return &pool.BankAccountData{
				AccountType: "bank",
			}, []models.FieldNamer{pool.BankAccount().AccountType()}
		})

	pool.BankAccount().Methods().ComputeSanitizedAccountNumber().DeclareMethod(
		`ComputeSanitizedAccountNumber removes all spaces and invalid characters from account number`,
		func(rs pool.BankAccountSet) (*pool.BankAccountData, []models.FieldNamer) {
			rg, _ := regexp.Compile("\\W+")
			san := rg.ReplaceAllString(rs.Name(), "")
			return &pool.BankAccountData{
				SanitizedAccountNumber: san,
			}, []models.FieldNamer{pool.BankAccount().SanitizedAccountNumber()}
		})
}
