// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/operator"
	"github.com/hexya-erp/hexya/pool"
)

// CompanyGetUserCurrency returns the currency of the current user's company if it exists
// or the default currency otherwise
func CompanyGetUserCurrency(env models.Environment, fMap models.FieldMap) interface{} {
	currency := pool.User().NewSet(env).GetCompany().Currency()
	if currency.IsEmpty() {
		return pool.Company().NewSet(env).GetEuro()
	}
	return currency
}

func init() {
	companyModel := pool.Company().DeclareModel()
	companyModel.AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{String: "Company Name", Size: 128, Required: true,
			Related: "Partner.Name", Unique: true},
		"Parent": models.Many2OneField{RelationModel: pool.Company(),
			String: "Parent Company", Index: true, Constraint: pool.Company().Methods().CheckParent()},
		"Children": models.One2ManyField{RelationModel: pool.Company(),
			ReverseFK: "Parent", String: "Child Companies"},
		"Partner": models.Many2OneField{RelationModel: pool.Partner(),
			Required: true, Index: true},
		"Tagline": models.CharField{},
		"Logo":    models.BinaryField{Related: "Partner.Image"},
		"LogoWeb": models.BinaryField{Compute: pool.Company().Methods().ComputeLogoWeb(),
			Stored: true, Depends: []string{"Partner", "Partner.Image"}},
		"Currency": models.Many2OneField{RelationModel: pool.Currency(),
			Required: true, Default: CompanyGetUserCurrency},
		"Users":   models.Many2ManyField{RelationModel: pool.User(), String: "Accepted Users"},
		"Street":  models.CharField{Related: "Partner.Street"},
		"Street2": models.CharField{Related: "Partner.Street2"},
		"Zip":     models.CharField{Related: "Partner.Zip"},
		"City":    models.CharField{Related: "Partner.City"},
		"State": models.Many2OneField{RelationModel: pool.CountryState(),
			Related: "Partner.State", OnChange: pool.Company().Methods().OnChangeState()},
		"Country": models.Many2OneField{RelationModel: pool.Country(),
			Related: "Partner.Country", OnChange: pool.Company().Methods().OnChangeCountry()},
		"Email":           models.CharField{Related: "Partner.Email"},
		"Phone":           models.CharField{Related: "Partner.Phone"},
		"Fax":             models.CharField{Related: "Partner.Fax"},
		"Website":         models.CharField{Related: "Partner.Website"},
		"VAT":             models.CharField{Related: "Partner.VAT"},
		"CompanyRegistry": models.CharField{Size: 64},
	})

	companyModel.Methods().Copy().Extend("",
		func(rs pool.CompanySet, overrides *pool.CompanyData, fieldsToReset ...models.FieldNamer) pool.CompanySet {
			rs.EnsureOne()
			_, eName := overrides.Get(pool.Company().Name(), fieldsToReset...)
			_, ePartner := overrides.Get(pool.Company().Partner(), fieldsToReset...)
			if !eName && !ePartner {
				copyPartner := rs.Partner().Copy(new(pool.PartnerData))
				overrides.Partner = copyPartner
				overrides.Name = copyPartner.Name()
				fieldsToReset = append(fieldsToReset, pool.Company().Partner(), pool.Company().Name())
			}
			return rs.Super().Copy(overrides, fieldsToReset...)
		})

	companyModel.Methods().ComputeLogoWeb().DeclareMethod(
		`ComputeLogoWeb returns a resized version of the company logo`,
		func(rs pool.CompanySet) (*pool.CompanyData, []models.FieldNamer) {
			res := pool.CompanyData{
				LogoWeb: rs.Logo(),
			}
			return &res, []models.FieldNamer{rs.Model().LogoWeb()}
		})

	companyModel.Methods().OnChangeState().DeclareMethod(
		`OnchangeState sets the country to the country of the state when you select one.`,
		func(rs pool.CompanySet) (*pool.CompanyData, []models.FieldNamer) {
			return &pool.CompanyData{
				Country: rs.State().Country(),
			}, []models.FieldNamer{pool.Company().Country()}
		})

	companyModel.Methods().GetEuro().DeclareMethod(
		`GetEuro returns the currency with rate 1 (euro by default, unless changed by the user)`,
		func(rs pool.CompanySet) pool.CurrencySet {
			return pool.CurrencyRate().Search(rs.Env(), pool.CurrencyRate().Rate().Equals(1)).Limit(1).Currency()
		})

	companyModel.Methods().OnChangeCountry().DeclareMethod(
		`OnChangeCountry updates the currency of this company on a country change`,
		func(rs pool.CompanySet) (*pool.CompanyData, []models.FieldNamer) {
			if rs.Country().IsEmpty() {
				userCurrency := CompanyGetUserCurrency(rs.Env(), rs.First().FieldMap()).(pool.CurrencySet)
				return &pool.CompanyData{
					Currency: userCurrency,
				}, []models.FieldNamer{pool.Company().Currency()}
			}
			return &pool.CompanyData{
				Currency: rs.Country().Currency(),
			}, []models.FieldNamer{pool.Company().Currency()}
		})

	companyModel.Methods().CompanyDefaultGet().DeclareMethod(
		`CompanyDefaultGet returns the default company (usually the user's company).`,
		func(rs pool.CompanySet) pool.CompanySet {
			return pool.User().NewSet(rs.Env()).GetCompany()
		})

	companyModel.Methods().Create().Extend("",
		func(rs pool.CompanySet, data *pool.CompanyData) pool.CompanySet {
			if !data.Partner.IsEmpty() {
				return rs.Super().Create(data)
			}
			partner := pool.Partner().Create(rs.Env(), &pool.PartnerData{
				Name:        data.Name,
				CompanyType: "company",
				Image:       data.Logo,
				Customer:    false,
				Email:       data.Email,
				Phone:       data.Phone,
				Website:     data.Website,
				VAT:         data.VAT,
			})
			data.Partner = partner
			company := rs.Super().Create(data)
			partner.SetCompany(company)
			return company
		})

	companyModel.Methods().CheckParent().DeclareMethod(
		`CheckParent checks that there is no recursion in the company tree`,
		func(rs pool.CompanySet) {
			rs.CheckRecursion()
		})

	companyModel.Methods().SearchByName().Extend("",
		func(rs pool.CompanySet, name string, op operator.Operator, additionalCond pool.CompanyCondition, limit int) pool.CompanySet {
			// We browse as superuser. Otherwise, the user would be able to
			// select only the currently visible companies (according to rules,
			// which are probably to allow to see the child companies) even if
			// she belongs to some other companies.
			rSet := rs
			companies := pool.Company().NewSet(rs.Env())
			if rs.Env().Context().HasKey("user_preference") {
				currentUser := pool.User().NewSet(rs.Env()).CurrentUser().Sudo()
				companies = currentUser.Companies().Union(currentUser.Company())
				rSet = rSet.Sudo()
			}
			return rSet.Super().SearchByName(name, op, additionalCond, limit).Union(companies)
		})
}
