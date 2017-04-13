// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import "github.com/npiganeau/yep/yep/models"

func initPartner() {
	resPartner := models.NewModel("ResPartner")
	resPartner.AddCharField("Name", models.StringFieldParams{})
	resPartner.AddDateField("Date", models.SimpleFieldParams{})
	//Title            *PartnerTitle
	resPartner.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: "ResPartner"})
	resPartner.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: "ResPartner", ReverseFK: "Parent"})
	resPartner.AddCharField("Ref", models.StringFieldParams{})
	resPartner.AddCharField("Lang", models.StringFieldParams{})
	resPartner.AddCharField("TZ", models.StringFieldParams{})
	resPartner.AddCharField("TZOffset", models.StringFieldParams{})
	resPartner.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: "ResUsers"})
	resPartner.AddCharField("VAT", models.StringFieldParams{})
	//Banks            []*PartnerBank
	resPartner.AddCharField("Website", models.StringFieldParams{})
	resPartner.AddCharField("Comment", models.StringFieldParams{})
	//Categories       []*PartnerCategory
	resPartner.AddFloatField("CreditLimit", models.FloatFieldParams{})
	resPartner.AddCharField("EAN13", models.StringFieldParams{})
	resPartner.AddBooleanField("Active", models.SimpleFieldParams{})
	resPartner.AddBooleanField("Customer", models.SimpleFieldParams{})
	resPartner.AddBooleanField("Supplier", models.SimpleFieldParams{})
	resPartner.AddBooleanField("Employee", models.SimpleFieldParams{})
	resPartner.AddCharField("Function", models.StringFieldParams{})
	resPartner.AddCharField("Type", models.StringFieldParams{})
	resPartner.AddCharField("Street", models.StringFieldParams{})
	resPartner.AddCharField("Street2", models.StringFieldParams{})
	resPartner.AddCharField("ZIP", models.StringFieldParams{})
	resPartner.AddCharField("City", models.StringFieldParams{})
	//State            *CountryState
	//Country          *Country
	resPartner.AddCharField("Email", models.StringFieldParams{})
	resPartner.AddCharField("Phone", models.StringFieldParams{})
	resPartner.AddCharField("Fax", models.StringFieldParams{})
	resPartner.AddCharField("Mobile", models.StringFieldParams{})
	resPartner.AddDateField("Birthdate", models.SimpleFieldParams{})
	resPartner.AddBooleanField("IsCompany", models.SimpleFieldParams{})
	resPartner.AddBooleanField("UseParentAddress", models.SimpleFieldParams{})
	resPartner.AddBinaryField("Image", models.SimpleFieldParams{})
	resPartner.AddBinaryField("ImageMedium", models.SimpleFieldParams{})
	resPartner.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: "ResCompany"})
	//Color            color.Color
	//Users []*ResUsers `orm:"reverse(many)"`

	//'has_image': fields.function(_has_image, type="boolean"),
	//'company_id': fields.many2one('res.company', 'Company', select=1),
	//'color': fields.integer('Color Index'),
	//'user_ids': fields.one2many('res.users', 'partner_id', 'Users'),
	//'contact_address': fields.function(_address_display,  type='char', string='Complete Address'),
	//
	//# technical field used for managing commercial fields
	//'commercial_partner_id': fields.function(_commercial_partner_id, type='many2one', relation='res.partner', string='Commercial Entity', store=_commercial_partner_store_triggers)

}
