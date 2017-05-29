// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func initPartner() {
	models.NewModel("Partner")

	partner := pool.Partner()
	partner.AddCharField("Name", models.StringFieldParams{})
	partner.AddDateField("Date", models.SimpleFieldParams{})
	//Title            *PartnerTitle
	partner.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: "Partner"})
	partner.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: "Partner", ReverseFK: "Parent"})
	partner.AddCharField("Ref", models.StringFieldParams{})
	partner.AddCharField("Lang", models.StringFieldParams{})
	partner.AddCharField("TZ", models.StringFieldParams{})
	partner.AddCharField("TZOffset", models.StringFieldParams{})
	partner.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: "User"})
	partner.AddCharField("VAT", models.StringFieldParams{})
	//Banks            []*PartnerBank
	partner.AddCharField("Website", models.StringFieldParams{})
	partner.AddCharField("Comment", models.StringFieldParams{})
	//Categories       []*PartnerCategory
	partner.AddFloatField("CreditLimit", models.FloatFieldParams{})
	partner.AddCharField("EAN13", models.StringFieldParams{})
	partner.AddBooleanField("Active", models.SimpleFieldParams{})
	partner.AddBooleanField("Customer", models.SimpleFieldParams{})
	partner.AddBooleanField("Supplier", models.SimpleFieldParams{})
	partner.AddBooleanField("Employee", models.SimpleFieldParams{})
	partner.AddCharField("Function", models.StringFieldParams{})
	partner.AddCharField("Type", models.StringFieldParams{})
	partner.AddCharField("Street", models.StringFieldParams{})
	partner.AddCharField("Street2", models.StringFieldParams{})
	partner.AddCharField("ZIP", models.StringFieldParams{})
	partner.AddCharField("City", models.StringFieldParams{})
	//State            *CountryState
	//Country          *Country
	partner.AddCharField("Email", models.StringFieldParams{})
	partner.AddCharField("Phone", models.StringFieldParams{})
	partner.AddCharField("Fax", models.StringFieldParams{})
	partner.AddCharField("Mobile", models.StringFieldParams{})
	partner.AddDateField("Birthdate", models.SimpleFieldParams{})
	partner.AddBooleanField("IsCompany", models.SimpleFieldParams{})
	partner.AddBooleanField("UseParentAddress", models.SimpleFieldParams{})
	partner.AddBinaryField("Image", models.SimpleFieldParams{})
	partner.AddBinaryField("ImageMedium", models.SimpleFieldParams{})
	partner.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: "Company"})
	//Color            color.Color
	//Users []*User `orm:"reverse(many)"`

	//'has_image': fields.function(_has_image, type="boolean"),
	//'company_id': fields.many2one('res.company', 'Company', select=1),
	//'color': fields.integer('Color Index'),
	//'user_ids': fields.one2many('res.users', 'partner_id', 'Users'),
	//'contact_address': fields.function(_address_display,  type='char', string='Complete Address'),
	//
	//# technical field used for managing commercial fields
	//'commercial_partner_id': fields.function(_commercial_partner_id, type='many2one', relation='res.partner', string='Commercial Entity', store=_commercial_partner_store_triggers)

	partner.Methods().NameGet().Extend("",
		func(rs pool.PartnerSet) string {
			res := rs.Super().NameGet()
			return fmt.Sprintf("%s (%d)", res, rs.ID())
		})

}
