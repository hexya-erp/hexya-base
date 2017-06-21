// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	company := pool.Company().DeclareModel()
	company.AddCharField("Name", models.StringFieldParams{String: "Company Name", Size: 128, Required: true, Related: "Partner.Name", Unique: true})
	company.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: pool.Company(), String: "Parent Company", Index: true})
	company.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: pool.Company(), ReverseFK: "Parent", String: "Child Companies"})
	company.AddMany2OneField("Partner", models.ForeignKeyFieldParams{RelationModel: pool.Partner(), Required: true, Index: true})
	company.AddCharField("Tagline", models.StringFieldParams{})
	company.AddBinaryField("Logo", models.SimpleFieldParams{Related: "Partner.Image"})
	company.AddBinaryField("LogoWeb", models.SimpleFieldParams{Compute: "ComputeLogoWeb", Stored: true, Depends: []string{"Partner", "Partner.Image"}})
	company.AddMany2OneField("Currency", models.ForeignKeyFieldParams{RelationModel: pool.Currency(), Required: true})
	company.AddMany2ManyField("Users", models.Many2ManyFieldParams{RelationModel: pool.User(), String: "Accepted Users"})
	company.AddCharField("Street", models.StringFieldParams{Related: "Partner.Street"})
	company.AddCharField("Street2", models.StringFieldParams{Related: "Partner.Street2"})
	company.AddCharField("Zip", models.StringFieldParams{Related: "Partner.Zip"})
	company.AddCharField("City", models.StringFieldParams{Related: "Partner.City"})
	company.AddMany2OneField("State", models.ForeignKeyFieldParams{RelationModel: pool.CountryState(), Related: "Partner.State"})
	company.AddMany2OneField("Country", models.ForeignKeyFieldParams{RelationModel: pool.Country(), Related: "Partner.Country"})
	company.AddCharField("Email", models.StringFieldParams{Related: "Partner.Email"})
	company.AddCharField("Phone", models.StringFieldParams{Related: "Partner.Phone"})
	company.AddCharField("Fax", models.StringFieldParams{Related: "Partner.Fax"})
	company.AddCharField("Website", models.StringFieldParams{Related: "Partner.Website"})
	company.AddCharField("VAT", models.StringFieldParams{Related: "Partner.VAT"})
	company.AddCharField("CompanyRegistry", models.StringFieldParams{Size: 64})

	company.Methods().ComputeLogoWeb().DeclareMethod(
		`ComputeLogoWeb returns a resized version of the company logo`,
		func(rs pool.CompanySet) (*pool.CompanyData, []models.FieldNamer) {
			res := pool.CompanyData{
				LogoWeb: rs.Logo(),
			}
			return &res, []models.FieldNamer{rs.Model().LogoWeb()}
		})
}
