// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"time"

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
)

type ResPartner struct {
	ID   int64
	Name string
	Date time.Time `yep:"type(date)"`
	//Title            *PartnerTitle
	Parent   pool.ResPartnerSet `yep:"type(many2one)"`
	Children pool.ResPartnerSet `yep:"type(one2many);fk(Parent)"`
	Ref      string
	Lang     string
	TZ       string
	TzOffset string
	User     pool.ResUsersSet `yep:"type(many2one)"`
	VAT      string
	//Banks            []*PartnerBank
	Website string
	Comment string
	//Categories       []*PartnerCategory
	CreditLimit float64
	EAN13       string
	Active      bool
	Customer    bool
	Supplier    bool
	Employee    bool
	Function    string
	Type        string
	Street      string
	Street2     string
	Zip         string
	City        string
	//State            *CountryState
	//Country          *Country
	Email            string
	Phone            string
	Fax              string
	Mobile           string
	Birthdate        models.Date
	IsCompany        bool
	UseParentAddress bool
	//Image            image.Image
	//Company          *Company
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

func initPartner() {
	models.CreateModel("ResPartner")
	models.ExtendModel("ResPartner", new(ResPartner))
}
