// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"fmt"
	"regexp"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/operator"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
)

func init() {
	h.Bank().DeclareModel()
	h.Bank().AddFields(map[string]models.FieldDefinition{
		"Name":    models.CharField{Required: true},
		"Street":  models.CharField{},
		"Street2": models.CharField{},
		"Zip":     models.CharField{},
		"City":    models.CharField{},
		"State": models.Many2OneField{RelationModel: h.CountryState(), String: "Fed. State",
			Filter: q.CountryState().Country().EqualsFunc(func(rs models.RecordSet) models.RecordSet {
				bank := rs.(h.BankSet)
				return bank.Country()
			})},
		"Country": models.Many2OneField{RelationModel: h.Country()},
		"Email":   models.CharField{},
		"Phone":   models.CharField{},
		"Fax":     models.CharField{},
		"Active":  models.BooleanField{Default: models.DefaultValue(true)},
		"BIC":     models.CharField{String: "Bank Identifier Cord", Index: true, Help: "Sometimes called BIC or Swift."},
	})
	h.Bank().Methods().NameGet().Extend("",
		func(rs h.BankSet) string {
			res := rs.Name()
			if rs.BIC() != "" {
				res = fmt.Sprintf("%s - %s", res, rs.BIC())
			}
			return res
		})

	h.Bank().Methods().SearchByName().Extend("",
		func(rs h.BankSet, name string, op operator.Operator, additionalCond q.BankCondition, limit int) h.BankSet {
			if name == "" {
				return rs.Super().SearchByName(name, op, additionalCond, limit)
			}
			cond := q.Bank().BIC().ILike(name+"%").Or().Name().AddOperator(op, name)
			if !additionalCond.Underlying().IsEmpty() {
				cond = cond.AndCond(additionalCond)
			}
			return h.Bank().Search(rs.Env(), cond).Limit(limit)
		})

	h.BankAccount().DeclareModel()
	h.BankAccount().AddFields(map[string]models.FieldDefinition{
		"AccountType": models.CharField{Compute: h.BankAccount().Methods().ComputeAccountType(), Depends: []string{""}},
		"Name":        models.CharField{String: "Account Number", Required: true},
		"SanitizedAccountNumber": models.CharField{Compute: h.BankAccount().Methods().ComputeSanitizedAccountNumber(),
			Stored: true, Depends: []string{"Name"}},
		"Partner": models.Many2OneField{RelationModel: h.Partner(),
			String: "Account Holder", OnDelete: models.Cascade, Index: true,
			Filter: q.Partner().IsCompany().Equals(true).Or().Parent().IsNull()},
		"Bank":     models.Many2OneField{RelationModel: h.Bank()},
		"BankName": models.CharField{Related: "Bank.Name"},
		"BankBIC":  models.CharField{Related: "Bank.BIC"},
		"Sequence": models.IntegerField{},
		"Currency": models.Many2OneField{RelationModel: h.Currency()},
		"Company":  models.Many2OneField{RelationModel: h.Company()},
	})
	h.BankAccount().AddSQLConstraint("unique_number", "unique(sanitized_account_number, company_id)", "Account Number must be unique")

	h.BankAccount().Methods().ComputeAccountType().DeclareMethod(
		`ComputeAccountType computes the type of account from the account number`,
		func(rs h.BankAccountSet) *h.BankAccountData {
			return &h.BankAccountData{
				AccountType: "bank",
			}
		})

	h.BankAccount().Methods().ComputeSanitizedAccountNumber().DeclareMethod(
		`ComputeSanitizedAccountNumber removes all spaces and invalid characters from account number`,
		func(rs h.BankAccountSet) *h.BankAccountData {
			rg, _ := regexp.Compile("\\W+")
			san := rg.ReplaceAllString(rs.Name(), "")
			return &h.BankAccountData{
				SanitizedAccountNumber: san,
			}
		})

}
