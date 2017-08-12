// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	partnerTitle := pool.PartnerTitle().DeclareModel()
	partnerTitle.AddCharField("Name", models.StringFieldParams{String: "Title", Required: true, Translate: true, Unique: true})
	partnerTitle.AddCharField("Shortcut", models.StringFieldParams{String: "Abbreviation", Translate: true})

	partnerCategory := pool.PartnerCategory().DeclareModel()
	partnerCategory.AddCharField("Name", models.StringFieldParams{String: "Category Name", Required: true, Translate: true})
	partnerCategory.AddIntegerField("Color", models.SimpleFieldParams{String: "Color Index"})
	partnerCategory.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: pool.PartnerCategory(),
		String: "Parent Tag", Index: true, OnDelete: models.Cascade})
	partnerCategory.AddCharField("CompleteName", models.StringFieldParams{String: "Full Name", Compute: "ComputeCompleteName"})
	partnerCategory.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: pool.PartnerCategory(),
		ReverseFK: "Parent", String: "Children Tags"})
	partnerCategory.AddMany2ManyField("Partners", models.Many2ManyFieldParams{RelationModel: pool.Partner()})

	partnerCategory.AddMethod("ComputeCompleteName",
		`ComputeCompleteName returns the complete name of the tag with all the parents`,
		func(s pool.PartnerCategorySet) (*pool.PartnerCategoryData, []models.FieldNamer) {
			completeName := s.Name()
			for rs := s; !rs.Parent().IsEmpty(); rs = rs.Parent() {
				completeName = fmt.Sprintf("%s/%s", rs.Parent().Name(), completeName)
			}
			res := &pool.PartnerCategoryData{
				CompleteName: completeName,
			}
			return res, []models.FieldNamer{pool.PartnerCategory().CompleteName()}
		})

	partnerModel := pool.Partner().DeclareModel()
	partnerModel.AddCharField("Name", models.StringFieldParams{Required: true, Index: true, NoCopy: true})
	partnerModel.AddDateField("Date", models.SimpleFieldParams{})
	partnerModel.AddMany2OneField("Title", models.ForeignKeyFieldParams{RelationModel: pool.PartnerTitle()})
	partnerModel.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: pool.Partner()})
	partnerModel.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: pool.Partner(), ReverseFK: "Parent"})
	partnerModel.AddCharField("Ref", models.StringFieldParams{})
	partnerModel.AddCharField("Lang", models.StringFieldParams{})
	partnerModel.AddCharField("TZ", models.StringFieldParams{})
	partnerModel.AddCharField("TZOffset", models.StringFieldParams{})
	partnerModel.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: pool.User()})
	partnerModel.AddCharField("VAT", models.StringFieldParams{})
	//Banks            []*PartnerBank
	partnerModel.AddCharField("Website", models.StringFieldParams{})
	partnerModel.AddCharField("Comment", models.StringFieldParams{})
	partnerModel.AddMany2ManyField("Categories", models.Many2ManyFieldParams{RelationModel: pool.PartnerCategory()})
	partnerModel.AddFloatField("CreditLimit", models.FloatFieldParams{})
	partnerModel.AddCharField("EAN13", models.StringFieldParams{})
	partnerModel.AddBooleanField("Active", models.SimpleFieldParams{Default: models.DefaultValue(true)})
	partnerModel.AddBooleanField("Customer", models.SimpleFieldParams{})
	partnerModel.AddBooleanField("Supplier", models.SimpleFieldParams{})
	partnerModel.AddBooleanField("Employee", models.SimpleFieldParams{})
	partnerModel.AddCharField("Function", models.StringFieldParams{})
	partnerModel.AddSelectionField("Type", models.SelectionFieldParams{Selection: types.Selection{
		"contact": "Contact", "invoice": "Invoice Address", "delivery": "Shipping Address", "other": "Other Address"},
		Help:    "Used to select automatically the right address according to the context in sales and purchases documents.",
		Default: models.DefaultValue("contact"),
	})
	partnerModel.AddCharField("Street", models.StringFieldParams{})
	partnerModel.AddCharField("Street2", models.StringFieldParams{})
	partnerModel.AddCharField("Zip", models.StringFieldParams{})
	partnerModel.AddCharField("City", models.StringFieldParams{})
	partnerModel.AddMany2OneField("State", models.ForeignKeyFieldParams{RelationModel: pool.CountryState(),
		Filter: pool.CountryState().Country().EqualsEval("country_id")})
	partnerModel.AddMany2OneField("Country", models.ForeignKeyFieldParams{RelationModel: pool.Country()})
	partnerModel.AddCharField("Email", models.StringFieldParams{})
	partnerModel.AddCharField("Phone", models.StringFieldParams{})
	partnerModel.AddCharField("Fax", models.StringFieldParams{})
	partnerModel.AddCharField("Mobile", models.StringFieldParams{})
	partnerModel.AddDateField("Birthdate", models.SimpleFieldParams{})
	partnerModel.AddBooleanField("IsCompany", models.SimpleFieldParams{Compute: "ComputeIsCompany", Stored: true, Depends: []string{"CompanyType"}})
	partnerModel.AddBooleanField("UseParentAddress", models.SimpleFieldParams{})
	partnerModel.AddBinaryField("Image", models.SimpleFieldParams{})
	partnerModel.AddBinaryField("ImageMedium", models.SimpleFieldParams{})
	partnerModel.AddBinaryField("ImageSmall", models.SimpleFieldParams{})
	partnerModel.AddSelectionField("CompanyType", models.SelectionFieldParams{Selection: types.Selection{"person": "Individual", "company": "Company"},
		OnChange: "ComputeIsCompany", Default: models.DefaultValue("person")})
	partnerModel.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: pool.Company()})
	partnerModel.AddIntegerField("Color", models.SimpleFieldParams{})
	partnerModel.AddOne2ManyField("Users", models.ReverseFieldParams{RelationModel: pool.User(), ReverseFK: "Partner"})

	partnerModel.Methods().ComputeIsCompany().DeclareMethod(
		`ComputeIsCompany computes the IsCompany field from the selected CompanyType`,
		func(s pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			var res pool.PartnerData
			res.IsCompany = s.CompanyType() == "company"
			return &res, []models.FieldNamer{pool.Partner().IsCompany()}
		})

	partnerModel.AddSQLConstraint("check_name", "CHECK( (type='contact' AND name IS NOT NULL) or (type != 'contact') )", "Contacts require a name.")
	//'has_image': fields.function(_has_image, type="boolean"),
	//'contact_address': fields.function(_address_display,  type='char', string='Complete Address'),
	//'commercial_partner_id': fields.function(_commercial_partner_id, type='many2one', relation='res.partner', string='Commercial Entity', store=_commercial_partner_store_triggers)

}
