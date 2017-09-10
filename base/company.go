// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
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
	companyModel.AddCharField("Name", models.StringFieldParams{String: "Company Name", Size: 128, Required: true,
		Related: "Partner.Name", Unique: true})
	companyModel.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: pool.Company(),
		String: "Parent Company", Index: true, Constraint: pool.Company().Methods().CheckParent()})
	companyModel.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: pool.Company(),
		ReverseFK: "Parent", String: "Child Companies"})
	companyModel.AddMany2OneField("Partner", models.ForeignKeyFieldParams{RelationModel: pool.Partner(),
		Required: true, Index: true})
	companyModel.AddCharField("Tagline", models.StringFieldParams{})
	companyModel.AddBinaryField("Logo", models.SimpleFieldParams{Related: "Partner.Image"})
	companyModel.AddBinaryField("LogoWeb", models.SimpleFieldParams{Compute: pool.Company().Methods().ComputeLogoWeb(),
		Stored: true, Depends: []string{"Partner", "Partner.Image"}})
	companyModel.AddMany2OneField("Currency", models.ForeignKeyFieldParams{RelationModel: pool.Currency(),
		Required: true, Default: CompanyGetUserCurrency})
	companyModel.AddMany2ManyField("Users", models.Many2ManyFieldParams{RelationModel: pool.User(), String: "Accepted Users"})
	companyModel.AddCharField("Street", models.StringFieldParams{Related: "Partner.Street"})
	companyModel.AddCharField("Street2", models.StringFieldParams{Related: "Partner.Street2"})
	companyModel.AddCharField("Zip", models.StringFieldParams{Related: "Partner.Zip"})
	companyModel.AddCharField("City", models.StringFieldParams{Related: "Partner.City"})
	companyModel.AddMany2OneField("State", models.ForeignKeyFieldParams{RelationModel: pool.CountryState(),
		Related: "Partner.State", OnChange: pool.Company().Methods().OnChangeState()})
	companyModel.AddMany2OneField("Country", models.ForeignKeyFieldParams{RelationModel: pool.Country(),
		Related: "Partner.Country", OnChange: pool.Company().Methods().OnChangeCountry()})
	companyModel.AddCharField("Email", models.StringFieldParams{Related: "Partner.Email"})
	companyModel.AddCharField("Phone", models.StringFieldParams{Related: "Partner.Phone"})
	companyModel.AddCharField("Fax", models.StringFieldParams{Related: "Partner.Fax"})
	companyModel.AddCharField("Website", models.StringFieldParams{Related: "Partner.Website"})
	companyModel.AddCharField("VAT", models.StringFieldParams{Related: "Partner.VAT"})
	companyModel.AddCharField("CompanyRegistry", models.StringFieldParams{Size: 64})

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
		func(rs pool.CompanySet, data models.FieldMapper) pool.CompanySet {
			vals, _ := rs.DataStruct(data.FieldMap())
			if !vals.Partner.IsEmpty() {
				return rs.Super().Create(data)
			}
			partner := pool.Partner().Create(rs.Env(), &pool.PartnerData{
				Name:        vals.Name,
				CompanyType: "company",
				Image:       vals.Logo,
				Customer:    false,
				Email:       vals.Email,
				Phone:       vals.Phone,
				Website:     vals.Website,
				VAT:         vals.VAT,
			})
			vals.Partner = partner
			company := rs.Super().Create(vals)
			partner.SetCompany(company)
			return company
		})

	companyModel.Methods().CheckParent().DeclareMethod(
		`CheckParent checks that there is no recursion in the company tree`,
		func(rs pool.CompanySet) {
			rs.CheckRecursion()
		})
}
