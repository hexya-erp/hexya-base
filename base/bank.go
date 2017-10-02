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
	pool.Bank().AddFields(map[string]models.FieldDefinition{
		"Name":    models.CharField{Required: true},
		"Street":  models.CharField{},
		"Street2": models.CharField{},
		"Zip":     models.CharField{},
		"City":    models.CharField{},
		"State": models.Many2OneField{RelationModel: pool.CountryState(), String: "Fed. State",
			Filter: pool.CountryState().Country().EqualsFunc(func(rs models.RecordSet) pool.CountrySet {
				bank := rs.(pool.BankSet)
				return bank.Country()
			})},
		"Country": models.Many2OneField{RelationModel: pool.Country()},
		"Email":   models.CharField{},
		"Phone":   models.CharField{},
		"Fax":     models.CharField{},
		"Active":  models.BooleanField{Default: models.DefaultValue(true)},
		"BIC":     models.CharField{String: "Bank Identifier Cord", Index: true, Help: "Sometimes called BIC or Swift."},
	})
	pool.Bank().Methods().NameGet().Extend("",
		func(rs pool.BankSet) string {
			res := rs.Name()
			if rs.BIC() != "" {
				res = fmt.Sprintf("%s - %s", res, rs.BIC())
			}
			return res
		})

	pool.BankAccount().DeclareModel()
	pool.BankAccount().AddFields(map[string]models.FieldDefinition{
		"AccountType": models.CharField{Compute: pool.BankAccount().Methods().ComputeAccountType(), Depends: []string{""}},
		"Name":        models.CharField{String: "Account Number", Required: true},
		"SanitizedAccountNumber": models.CharField{Compute: pool.BankAccount().Methods().ComputeSanitizedAccountNumber(),
			Stored: true, Depends: []string{"Name"}},
		"Partner": models.Many2OneField{RelationModel: pool.Partner(),
			String: "Account Holder", OnDelete: models.Cascade, Index: true,
			Filter: pool.Partner().IsCompany().Equals(true).Or().Parent().IsNull()},
		"Bank":     models.Many2OneField{RelationModel: pool.Bank()},
		"BankName": models.CharField{Related: "Bank.Name"},
		"BankBIC":  models.CharField{Related: "Bank.BIC"},
		"Sequence": models.IntegerField{},
		"Currency": models.Many2OneField{RelationModel: pool.Currency()},
		"Company":  models.Many2OneField{RelationModel: pool.Company()},
	})
	pool.BankAccount().AddSQLConstraint("unique_number", "unique(sanitized_account_number, company_id)", "Account Number must be unique")

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
