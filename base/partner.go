// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/http"
	"net/mail"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hexya-erp/hexya-base/base/basetypes"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/b64image"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/typesutils"
	"github.com/hexya-erp/hexya/hexya/views"
	"github.com/hexya-erp/hexya/pool"
)

const gravatarBaseURL string = "https://www.gravatar.com/avatar"

func init() {
	partnerTitle := pool.PartnerTitle().DeclareModel()
	partnerTitle.AddFields(map[string]models.FieldDefinition{
		"Name":     models.CharField{String: "Title", Required: true, Translate: true, Unique: true},
		"Shortcut": models.CharField{String: "Abbreviation", Translate: true},
	})

	partnerCategory := pool.PartnerCategory().DeclareModel()
	partnerCategory.AddFields(map[string]models.FieldDefinition{
		"Name":  models.CharField{String: "Category Name", Required: true, Translate: true},
		"Color": models.IntegerField{String: "Color Index"},
		"Parent": models.Many2OneField{RelationModel: pool.PartnerCategory(),
			String: "Parent Tag", Index: true, OnDelete: models.Cascade},
		"CompleteName": models.CharField{String: "Full Name",
			Compute: pool.PartnerCategory().Methods().ComputeCompleteName(), Depends: []string{"Parent", "Name"}},
		"Children": models.One2ManyField{RelationModel: pool.PartnerCategory(),
			ReverseFK: "Parent", String: "Children Tags"},
		"Partners": models.Many2ManyField{RelationModel: pool.Partner()},
	})

	partnerCategory.Methods().ComputeCompleteName().DeclareMethod(
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
	partnerModel.AddFields(map[string]models.FieldDefinition{
		"Name":  models.CharField{Required: true, Index: true, NoCopy: true},
		"Date":  models.DateField{Index: true},
		"Title": models.Many2OneField{RelationModel: pool.PartnerTitle()},
		"Parent": models.Many2OneField{RelationModel: pool.Partner(), Index: true,
			Constraint: pool.Partner().Methods().CheckParent(), OnChange: pool.Partner().Methods().OnchangeParent()},
		"ParentName": models.CharField{Related: "Parent.Name"},

		"Children": models.One2ManyField{RelationModel: pool.Partner(),
			ReverseFK: "Parent", Filter: pool.Partner().Active().Equals(true)},
		"Ref": models.CharField{String: "Internal Reference", Index: true},
		"Lang": models.CharField{String: "Language",
			Default: func(env models.Environment, maps models.FieldMap) interface{} {
				return env.Context().GetString("lang")
			}, Help: `If the selected language is loaded in the system, all documents related to
this contact will be printed in this language. If not, it will be English.`},
		"TZ": models.CharField{String: "Timezone",
			Default: func(env models.Environment, maps models.FieldMap) interface{} {
				return env.Context().GetString("tz")
			}, Help: `"The partner's timezone, used to output proper date and time values
inside printed reports. It is important to set a value for this field.
You should use the same timezone that is otherwise used to pick and
render date and time values: your computer's timezone.`},
		"TZOffset": models.CharField{Compute: pool.Partner().Methods().ComputeTZOffset(),
			String: "Timezone Offset", Depends: []string{"TZ"}},
		"User": models.Many2OneField{RelationModel: pool.User(),
			String: "Salesperson", Help: "The internal user that is in charge of communicating with this contact if any."},
		"VAT": models.CharField{String: "TIN", Help: `Tax Identification Number.
Fill it if the company is subjected to taxes.
Used by the some of the legal statements.`},
		"Banks":   models.One2ManyField{String: "Bank Accounts", RelationModel: pool.BankAccount(), ReverseFK: "Partner"},
		"Website": models.CharField{Help: "Website of Partner or Company"},
		"Comment": models.CharField{String: "Notes"},
		"Categories": models.Many2ManyField{RelationModel: pool.PartnerCategory(),
			String: "Tags", Default: func(env models.Environment, maps models.FieldMap) interface{} {
				return pool.PartnerCategory().Browse(env, []int64{env.Context().GetInteger("category_id")})
			}},
		"CreditLimit": models.FloatField{},
		"Barcode":     models.CharField{},
		"Active":      models.BooleanField{Default: models.DefaultValue(true)},
		"Customer": models.BooleanField{String: "Is a Customer", Default: models.DefaultValue(true),
			Help: "Check this box if this contact is a customer."},
		"Supplier": models.BooleanField{String: "Is a Vendor",
			Help: `Check this box if this contact is a vendor.
If it's not checked, purchase people will not see it when encoding a purchase order.`},
		"Employee": models.BooleanField{Help: "Check this box if this contact is an Employee."},
		"Function": models.CharField{String: "Job Position"},
		"Type": models.SelectionField{Selection: types.Selection{
			"contact": "Contact", "invoice": "Invoice Address", "delivery": "Shipping Address", "other": "Other Address"},
			Help:    "Used to select automatically the right address according to the context in sales and purchases documents.",
			Default: models.DefaultValue("contact"),
		},
		"Street":  models.CharField{},
		"Street2": models.CharField{},
		"Zip":     models.CharField{},
		"City":    models.CharField{},
		"State": models.Many2OneField{RelationModel: pool.CountryState(),
			Filter: pool.CountryState().Country().EqualsEval("country_id"), OnDelete: models.Restrict},
		"Country": models.Many2OneField{RelationModel: pool.Country(),
			OnDelete: models.Restrict},
		"Email":          models.CharField{OnChange: pool.Partner().Methods().OnchangeEmail()},
		"EmailFormatted": models.CharField{Compute: pool.Partner().Methods().ComputeEmailFormatted(), Help: "Formatted email address 'Name <email@domain>'", Depends: []string{"Name", "Email"}},
		"Phone":          models.CharField{},
		"Fax":            models.CharField{},
		"Mobile":         models.CharField{},
		"IsCompany": models.BooleanField{Default: models.DefaultValue(false),
			Help: "Check if the contact is a company, otherwise it is a person"},
		// CompanyType is only an interface field, do not use it in business logic
		"CompanyType": models.SelectionField{
			Selection: types.Selection{"person": "Individual", "company": "Company"},
			Compute:   pool.Partner().Methods().ComputeCompanyType(),
			Depends:   []string{"IsCompany"}, Inverse: pool.Partner().Methods().InverseCompanyType(),
			OnChange: pool.Partner().Methods().OnchangeCompanyType(),
			Default:  models.DefaultValue("person")},
		"Company": models.Many2OneField{RelationModel: pool.Company()},
		"Color":   models.IntegerField{},
		"Users":   models.One2ManyField{RelationModel: pool.User(), ReverseFK: "Partner"},
		"PartnerShare": models.BooleanField{String: "Share Partner",
			Compute: pool.Partner().Methods().ComputePartnerShare(), Stored: true, Depends: []string{"Users", "Users.Share"},
			Help: `Either customer (no user), either shared user. Indicated the current partner is a customer without
access or with a limited access created for sharing data.`},
		"ContactAddress": models.CharField{Compute: pool.Partner().Methods().ComputeContactAddress(),
			String: "Complete Address", Depends: []string{"Street", "Street2", "Zip", "City", "State", "Country",
				"Country.AddressFormat", "Country.Code", "Country.Name", "CompanyName", "State.Code", "State.Name"}},

		"CommercialPartner": models.Many2OneField{RelationModel: pool.Partner(),
			Compute: pool.Partner().Methods().ComputeCommercialPartner(), String: "Commercial Entity", Stored: true,
			Index: true, Depends: []string{"IsCompany", "Parent", "Parent.CommercialPartner"}},
		"CommercialCompanyName": models.CharField{
			Compute: pool.Partner().Methods().ComputeCommercialCompanyName(), Stored: true,
			Depends: []string{"CompanyName", "Parent", "Parent.IsCompany", "CommercialPartner", "CommercialPartner.Name"}},
		"CompanyName": models.CharField{},

		"Image": models.BinaryField{
			Help: "This field holds the image used as avatar for this contact, limited to 1024x1024px"},
		"ImageMedium": models.BinaryField{
			Help: `Medium-sized image of this contact. It is automatically
resized as a 128x128px image, with aspect ratio preserved.
Use this field in form views or some kanban views.`},
		"ImageSmall": models.BinaryField{
			Help: `Small-sized image of this contact. It is automatically
resized as a 64x64px image, with aspect ratio preserved.
Use this field anywhere a small image is required.`},
	})

	partnerModel.Fields().ParentName().RevokeAccess(security.GroupEveryone, security.Write)
	//partnerModel.Fields().DisplayName().SetDepends([]string{"IsCompany", "Name", "Parent.Name", "Type", "CompanyName"})

	partnerModel.AddSQLConstraint("check_name",
		"CHECK( (type='contact' AND name IS NOT NULL) or (type != 'contact') )",
		"Contacts require a name.")

	partnerModel.Methods().ComputeDisplayName().Extend("",
		func(rs pool.PartnerSet) (models.FieldMap, []models.FieldNamer) {
			rSet := rs.
				WithContext("show_address", false).
				WithContext("show_address_only", false).
				WithContext("show_email", false)
			return rSet.Super().ComputeDisplayName()
		})

	partnerModel.Methods().ComputeTZOffset().DeclareMethod(
		`ComputeTZOffset computes the timezone offset`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			// TODO Implement TZOffset
			return &pool.PartnerData{
				TZOffset: "",
			}, []models.FieldNamer{pool.Partner().TZOffset()}
		})

	partnerModel.Methods().ComputePartnerShare().DeclareMethod(
		`ComputePartnerShare computes the PartnerShare field`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			var partnerShare bool
			if rs.Users().IsEmpty() {
				partnerShare = true
			}
			for _, user := range rs.Users().Records() {
				if user.Share() {
					partnerShare = true
					break
				}
			}
			return &pool.PartnerData{
				PartnerShare: partnerShare,
			}, []models.FieldNamer{pool.Partner().PartnerShare()}
		})

	partnerModel.Methods().ComputeContactAddress().DeclareMethod(
		`ComputeContactAddress computes the contact's address according to the contact's country standards`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			return &pool.PartnerData{
				ContactAddress: rs.DisplayAddress(false),
			}, []models.FieldNamer{pool.Partner().ContactAddress()}
		})

	partnerModel.Methods().ComputeCommercialPartner().DeclareMethod(
		`ComputeCommercialPartner computes the commercial partner, which is the first company ancestor or the top
		ancestor if none are companies`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			commercialPartner := rs
			if !rs.IsCompany() && !rs.Parent().IsEmpty() {
				commercialPartner = rs.Parent().CommercialPartner()
			}
			return &pool.PartnerData{
				CommercialPartner: commercialPartner,
			}, []models.FieldNamer{pool.Partner().CommercialPartner()}
		})

	partnerModel.Methods().ComputeCommercialCompanyName().DeclareMethod(
		`ComputeCommercialCompanyName returns the name of the commercial partner company`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			commPartnerName := rs.CommercialPartner().Name()
			if !rs.CommercialPartner().IsCompany() {
				commPartnerName = rs.CompanyName()
			}
			return &pool.PartnerData{
				CommercialCompanyName: commPartnerName,
			}, []models.FieldNamer{pool.Partner().CommercialCompanyName()}
		})

	partnerModel.Methods().GetDefaultImage().DeclareMethod(
		`GetDefaultImage returns a default image for the partner (base64 encoded)`,
		func(rs pool.PartnerSet, partnerType string, isCompany bool, Parent pool.PartnerSet) string {
			if rs.Env().Context().HasKey("install_mode") {
				return ""
			}
			var img string
			if partnerType == "other" && !Parent.IsEmpty() {
				parentImage := Parent.Image()
				if parentImage != "" {
					img = parentImage
				}
			}
			if img == "" {
				var (
					colorize    bool
					imgFileName string
				)
				switch {
				case partnerType == "invoice":
					imgFileName = "money.png"
				case partnerType == "delivery":
					imgFileName = "truck.png"
				case isCompany:
					imgFileName = "company_logo.png"
				default:
					imgFileName = "avatar.png"
					colorize = true
				}
				path := filepath.Join(generate.HexyaDir, "hexya", "server", "static", "base", "img", imgFileName)
				content, err := ioutil.ReadFile(path)
				if err != nil {
					log.Warn("Missing ressource", "image", path)
				}
				img = base64.StdEncoding.EncodeToString(content)
				if colorize {
					img = b64image.Colorize(img, color.RGBA{})
				}
			}
			return img
		})

	partnerModel.Methods().CheckParent().DeclareMethod(
		`CheckParent checks for recursion in the partners parenthood`,
		func(rs pool.PartnerSet) {
			if !rs.CheckRecursion() {
				log.Panic(rs.T("You cannot create recursive Partner hierarchies."))
			}
		})

	partnerModel.Methods().Copy().Extend("",
		func(rs pool.PartnerSet, overrides *pool.PartnerData, fieldsToUnset ...models.FieldNamer) pool.PartnerSet {
			rs.EnsureOne()
			vals, fieldsToUnset := rs.DataStruct(overrides.FieldMap(fieldsToUnset...))
			vals.Name = rs.T("%s (copy)", rs.Name())
			fieldsToUnset = append(fieldsToUnset, pool.Partner().Name())
			return rs.Super().Copy(vals, fieldsToUnset...)
		})

	partnerModel.Methods().OnchangeParent().DeclareMethod(
		`OnchangeParent updates the current partner data when its parent field
		is modified`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			if rs.Parent().IsEmpty() || rs.Type() != "contact" {
				return &pool.PartnerData{}, []models.FieldNamer{}
			}

			var parentHasAddress bool
			for _, addrField := range rs.AddressFields() {
				if !typesutils.IsZero(rs.Parent().Get(addrField.String())) {
					parentHasAddress = true
					break
				}
			}
			if !parentHasAddress {
				return &pool.PartnerData{}, []models.FieldNamer{}
			}
			resMap := make(models.FieldMap)
			for _, addrField := range rs.AddressFields() {
				resMap.Set(addrField.String(), rs.Parent().Get(addrField.String()), pool.Partner().Underlying())
			}

			return rs.DataStruct(resMap)
		})

	partnerModel.Methods().OnchangeEmail().DeclareMethod(
		`OnchangeEmail updates the user Gravatar image`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			if rs.Image() != "" || rs.Email() == "" || rs.Env().Context().HasKey("no_gravatar") {
				return &pool.PartnerData{}, []models.FieldNamer{}
			}
			return &pool.PartnerData{
				Image: rs.GetGravatarImage(rs.Email()),
			}, []models.FieldNamer{pool.Partner().Image()}
		})

	partnerModel.Methods().ComputeEmailFormatted().DeclareMethod(
		`ComputeEmailFormatted returns a 'Name <email@domain>' formatted string`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			addr := mail.Address{Name: rs.Name(), Address: rs.Email()}
			return &pool.PartnerData{
				EmailFormatted: addr.String(),
			}, []models.FieldNamer{pool.Partner().EmailFormatted()}
		})

	partnerModel.Methods().ComputeCompanyType().DeclareMethod(
		`ComputeIsCompany computes the IsCompany field from the selected CompanyType`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			companyType := "person"
			if rs.IsCompany() {
				companyType = "company"
			}
			return &pool.PartnerData{
				CompanyType: companyType,
			}, []models.FieldNamer{pool.Partner().CompanyType()}
		})

	partnerModel.Methods().InverseCompanyType().DeclareMethod(
		`InverseCompanyType sets the IsCompany field according to the given CompanyType`,
		func(rs pool.PartnerSet, vals models.FieldMapper) {
			values, _ := rs.DataStruct(vals.FieldMap())
			rs.SetIsCompany(values.CompanyType == "company")
		})

	partnerModel.Methods().OnchangeCompanyType().DeclareMethod(
		`OnchangeCompanyType updates the IsCompany field according to the selected type`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			res := &pool.PartnerData{
				IsCompany: rs.CompanyType() == "company",
			}
			return res, []models.FieldNamer{pool.Partner().IsCompany()}

		})

	partnerModel.Methods().UpdateFieldValues().DeclareMethod(
		`UpdateFieldValues returns a PartnerData struct with its values set to
		this partner's values on the given fields. The other fields are left to their
		Go default value. This method is used to update fields from a partner to its
		relatives.`,
		func(rs pool.PartnerSet, fields ...models.FieldNamer) (*pool.PartnerData, []models.FieldNamer) {
			res := make(models.FieldMap)
			fInfos := rs.FieldsGet(models.FieldsGetArgs{})
			for _, f := range fields {
				fJSON := pool.Partner().JSONizeFieldName(f.String())
				if fInfos[fJSON].Type == fieldtype.One2Many {
					log.Panic(rs.T("One2Many fields cannot be synchronized as part of 'commercial_fields' or 'address fields'"))
				}
				res[fJSON] = rs.Get(fJSON)
			}
			return rs.DataStruct(res)
		})

	partnerModel.Methods().AddressFields().DeclareMethod(
		`AddressFields returns the list of fields which are part of the address.
		These are used to automate behaviours on contact addresses.`,
		func(rs pool.PartnerSet) []models.FieldNamer {
			return []models.FieldNamer{
				pool.Partner().Street(), pool.Partner().Street2(), pool.Partner().Zip(),
				pool.Partner().City(), pool.Partner().State(), pool.Partner().Country(),
			}
		})

	partnerModel.Methods().UpdateAddress().DeclareMethod(
		`UpdateAddress updates this PartnerSet only with the address fields of
		the given vals. Other values passed are discarded.`,
		func(rs pool.PartnerSet, vals *pool.PartnerData, fieldsToReset ...models.FieldNamer) bool {
			valsMap := vals.FieldMap(fieldsToReset...)
			res := make(models.FieldMap)
			for _, addrField := range rs.AddressFields() {
				fValue, _ := valsMap.Get(addrField.String(), pool.Partner().Underlying())
				if !typesutils.IsZero(fValue) {
					res[addrField.String()], _ = vals.FieldMap(fieldsToReset...).Get(addrField.String(), pool.Partner().Underlying())
				}
			}
			if len(res) == 0 {
				return false
			}
			wData, fields := rs.DataStruct(res)
			return rs.WithContext("goto_super", true).Write(wData, fields...)
		})

	partnerModel.Methods().CommercialFields().DeclareMethod(
		`CommercialFields returns the list of fields that are managed by the commercial entity
        to which a partner belongs. These fields are meant to be hidden on
        partners that aren't "commercial entities"" themselves, and will be
        delegated to the parent "commercial entity"". The list is meant to be
        extended by inheriting classes.`,
		func(rs pool.PartnerSet) []models.FieldNamer {
			return []models.FieldNamer{
				pool.Partner().VAT(),
				pool.Partner().CreditLimit(),
			}
		})

	partnerModel.Methods().CommercialSyncFromCompany().DeclareMethod(
		`CommercialSyncFromCompany handle sync of commercial fields when a new parent commercial entity is set,
        as if they were related fields.`,
		func(rs pool.PartnerSet) bool {
			if rs.Equals(rs.CommercialPartner()) {
				return false
			}
			values, fieldsToUnset := rs.CommercialPartner().UpdateFieldValues(rs.CommercialFields()...)
			return rs.Write(values, fieldsToUnset...)
		})

	partnerModel.Methods().CommercialSyncToChildren().DeclareMethod(
		`CommercialSyncToChildren handle sync of commercial fields to descendants`,
		func(rs pool.PartnerSet) bool {
			partnerData, fieldsToUnset := rs.CommercialPartner().UpdateFieldValues(rs.CommercialFields()...)
			syncChildren := rs.Children().Search(pool.Partner().IsCompany().NotEquals(true))
			for _, child := range syncChildren.Records() {
				child.CommercialSyncToChildren()
			}
			partnerData.CommercialPartner = rs.CommercialPartner()
			fieldsToUnset = append(fieldsToUnset, pool.Partner().CommercialPartner())
			return syncChildren.WithContext("hexya_force_compute_write", true).Write(partnerData, fieldsToUnset...)
		})

	partnerModel.Methods().FieldsSync().DeclareMethod(
		`FieldsSync syncs commercial fields and address fields from company and to children after create/update,
        just as if those were all modeled as fields.related to the parent`,
		func(rs pool.PartnerSet, vals *pool.PartnerData, fieldsToUnset ...models.FieldNamer) {
			values, fieldsToUnset := rs.DataStruct(vals.FieldMap(fieldsToUnset...))
			// 1. From UPSTREAM: sync from parent
			// 1a. Commercial fields: sync if parent changed
			if !values.Parent.IsEmpty() {
				rs.CommercialSyncFromCompany()
			}
			// 1b. Address fields: sync if parent or use_parent changed *and* both are now set
			if !rs.Parent().IsEmpty() && rs.Type() == "contact" {
				onchangePartnerData, fieldsToReset := rs.OnchangeParent()
				rs.UpdateAddress(onchangePartnerData, fieldsToReset...)
			}
			// 2. To DOWNSTREAM: sync children
			if rs.Children().IsEmpty() {
				return
			}
			// 2a. Commercial Fields: sync if commercial entity
			if rs.Equals(rs.CommercialPartner()) {
				for _, commField := range rs.CommercialFields() {
					if !typesutils.IsZero(rs.Parent().Get(commField.String())) {
						rs.CommercialSyncToChildren()
						break
					}
				}
			}
			for _, child := range rs.Children().Search(pool.Partner().IsCompany().NotEquals(true)).Records() {
				if !child.CommercialPartner().Equals(rs.CommercialPartner()) {
					rs.CommercialSyncToChildren()
					break
				}

			}
			// 2b. Address fields: sync if address changed
			valsMap := vals.FieldMap(fieldsToUnset...)
			for _, addrField := range rs.AddressFields() {
				fValue, _ := valsMap.Get(addrField.String(), pool.Partner().Underlying())
				if !typesutils.IsZero(fValue) {
					contacts := rs.Children().Search(pool.Partner().Type().Equals("contact"))
					contacts.UpdateAddress(vals, fieldsToUnset...)
					break
				}
			}
		})

	partnerModel.Methods().HandleFirsrtContactCreation().DeclareMethod(
		`HandleFirsrtContactCreation: on creation of first contact for a company (or root) that has no address,
		assume contact address was meant to be company address`,
		func(rs pool.PartnerSet) {
			if !rs.Parent().IsCompany() && !rs.Parent().Parent().IsEmpty() {
				// Our parent is not a company, nor a root contact
				return
			}
			if rs.Parent().Children().Len() > 1 {
				// Our parent already has other children
				return
			}
			var addressDefined, parentAddressDefined bool
			for _, addrField := range rs.AddressFields() {
				if !typesutils.IsZero(rs.Parent().Get(addrField.String())) {
					parentAddressDefined = true
				}
				if !typesutils.IsZero(rs.Get(addrField.String())) {
					addressDefined = true
				}
			}
			if addressDefined && !parentAddressDefined {
				partnerData, fieldsToUnset := rs.UpdateFieldValues(rs.AddressFields()...)
				rs.Parent().UpdateAddress(partnerData, fieldsToUnset...)
			}
		})

	partnerModel.Methods().CleanWebsite().DeclareMethod(
		`CleanWebsite returns a cleaned website url including scheme.`,
		func(rs pool.PartnerSet, website string) string {
			websiteURL, err := url.Parse(website)
			if err != nil {
				log.Panic("Invalid URL for website", "URL", website)
			}
			if websiteURL.Scheme == "" {
				websiteURL.Scheme = "http"
			}
			return websiteURL.String()
		})

	partnerModel.Methods().Write().Extend("",
		func(rs pool.PartnerSet, vals *pool.PartnerData, fieldsToUnset ...models.FieldNamer) bool {
			if rs.Env().Context().HasKey("goto_super") {
				return rs.Super().Write(vals, fieldsToUnset...)
			}
			values, fieldsToUnset := rs.DataStruct(vals.FieldMap(fieldsToUnset...))
			if values.Website != "" {
				values.Website = rs.CleanWebsite(values.Website)
			}
			if !values.Parent.IsEmpty() {
				values.CompanyName = ""
			}
			// Partner must only allow to set the Company of a partner if it
			// is the same as the Company of all users that inherit from this partner
			// (this is to allow the code from User to write to the Partner!) or
			// if setting the Company to nil (this is compatible with any user
			// company)
			if !values.Company.IsEmpty() {
				for _, partner := range rs.Records() {
					for _, user := range partner.Users().Records() {
						if !user.Company().Equals(values.Company) {
							log.Panic(rs.T("You can not change the company as the partner/user has multiple users linked with different companies.", "company", values.Company.Name()))
						}
					}
				}
			}
			// TODO Resize images
			// tools.image_resize_images(vals)
			res := rs.Super().Write(values, fieldsToUnset...)
			for _, partner := range rs.Records() {
				for _, user := range partner.Users().Records() {
					if user.HasGroup("base_group_user") {
						pool.User().NewSet(rs.Env()).CheckExecutionPermission(pool.CommonMixin().Methods().Write().Underlying())
						break
					}
				}
				partner.FieldsSync(values, fieldsToUnset...)
			}
			return res
		})

	partnerModel.Methods().Create().Extend("",
		func(rs pool.PartnerSet, vals *pool.PartnerData) pool.PartnerSet {
			if vals.Website != "" {
				vals.Website = rs.CleanWebsite(vals.Website)
			}
			if !vals.Parent.IsEmpty() {
				vals.CompanyName = ""
			}
			if vals.Image == "" {
				vals.Image = rs.GetDefaultImage(vals.Type, vals.IsCompany, vals.Parent)
			}
			// TODO Resize images
			// tools.image_resize_images(vals)
			partner := rs.Super().Create(vals)
			partner.FieldsSync(vals)
			partner.HandleFirsrtContactCreation()
			return partner
		})

	partnerModel.Methods().CreateCompany().DeclareMethod(
		`CreateCompany creates the parent company of this partner if it has been given a CompanyName.`,
		func(rs pool.PartnerSet) bool {
			rs.EnsureOne()
			if rs.CompanyName() != "" {
				// Create parent company
				values, _ := rs.UpdateFieldValues(rs.AddressFields()...)
				values.Name = rs.CompanyName()
				values.IsCompany = true
				newCompany := rs.Create(values)
				// Set newCompany as my parent
				rs.SetParent(newCompany)
				rs.Children().Write(&pool.PartnerData{Parent: newCompany}, pool.Partner().Parent())
			}
			return true
		})

	partnerModel.Methods().OpenCommercialEntity().DeclareMethod(
		`OpenCommercialEntity is a utility method used to add an "Open Company" button in partner views`,
		func(rs pool.PartnerSet) actions.BaseAction {
			rs.EnsureOne()
			return actions.BaseAction{
				Type:     actions.ActionActWindow,
				Model:    "Partner",
				ViewMode: "form",
				ResID:    rs.CommercialPartner().ID(),
				Target:   "current",
				Flags:    map[string]interface{}{"form": map[string]interface{}{"action_buttons": true}},
			}
		})

	partnerModel.Methods().OpenParent().DeclareMethod(
		`OpenParent is a utility method used to add an "Open Parent" button in partner views`,
		func(rs pool.PartnerSet) actions.BaseAction {
			rs.EnsureOne()
			addressFormID := "base_view_partner_address_form"
			return actions.BaseAction{
				Type:     actions.ActionActWindow,
				Model:    "Partner",
				ViewMode: "form",
				Views:    []views.ViewTuple{{ID: addressFormID, Type: views.VIEW_TYPE_FORM}},
				ResID:    rs.Parent().ID(),
				Target:   "new",
				Flags:    map[string]interface{}{"form": map[string]interface{}{"action_buttons": true}},
			}
		})

	partnerModel.Methods().NameGet().Extend("",
		func(rs pool.PartnerSet) string {
			name := rs.Name()
			if rs.CompanyName() != "" || !rs.Parent().IsEmpty() {
				if name == "" {
					switch rs.Type() {
					case "invoice", "delivery", "other":
						fInfo := rs.FieldGet(pool.Partner().Type())
						name = fInfo.Selection[rs.Type()]
					}
				}
				if !rs.IsCompany() {
					name = fmt.Sprintf("%s, %s", rs.CommercialCompanyName(), name)
				}
			}
			if rs.Env().Context().GetBool("show_address_only") {
				name = rs.DisplayAddress(true)
			}
			if rs.Env().Context().GetBool("show_address") {
				name = name + "\n" + rs.DisplayAddress(true)
			}
			name = strings.Replace(name, "\n\n", "\n", -1)
			name = strings.Replace(name, "\n\n", "\n", -1)
			if rs.Env().Context().GetBool("show_email") && rs.Email() != "" {
				name = rs.EmailFormatted()
			}
			if rs.Env().Context().GetBool("html_format") {
				name = strings.Replace(name, "\n", "<br/>", -1)
			}
			return name
		})

	partnerModel.Methods().ParsePartnerName().DeclareMethod(
		`ParsePartnerName parses an email address to get the partner's name.
		It returns the name as first argument and the email as the second.

		Supported syntax:
            - 'Raoul <raoul@grosbedon.fr>': will find name and email address
            - otherwise: default, everything is set as the name (email is returned empty)`,
		func(rs pool.PartnerSet, email string) (string, string) {
			addr, err := mail.ParseAddress(email)
			if err != nil || addr.Name == "" {
				return email, ""
			}
			return addr.Name, addr.Address
		})

	partnerModel.Methods().FindOrCreate().DeclareMethod(
		`FindOrCreate finds a partner with the given 'email' or creates one.
		The given string should contain at least one email,
                e.g. "Raoul Grosbedon <r.g@grosbedon.fr>"`,
		func(rs pool.PartnerSet, email string) pool.PartnerSet {
			name, emailParsed := rs.ParsePartnerName(email)
			partners := pool.Partner().Search(rs.Env(), pool.Partner().Email().ILike(emailParsed)).Limit(1)
			if partners.IsEmpty() {
				rs.Create(&pool.PartnerData{
					Name:  name,
					Email: emailParsed,
				})
			}
			return partners
		})

	partnerModel.Methods().GetGravatarImage().DeclareMethod(
		`GetGravatarImage returns the image from Gravatar associated with the given email.
		Image is returned as a base64 encoded string.`,
		func(rs pool.PartnerSet, email string) string {
			emailHash := md5.Sum([]byte(strings.ToLower(email)))
			gravatarURL := fmt.Sprintf("%s/%x?%s", gravatarBaseURL, emailHash, "d=404&s=128")
			client := &http.Client{
				Timeout: 1 * time.Second,
			}
			resp, err := client.Get(gravatarURL)
			if resp.StatusCode == http.StatusNotFound || err != nil {
				return ""
			}
			img, err := ioutil.ReadAll(resp.Body)
			if len(img) == 0 || err != nil {
				return ""
			}
			return base64.StdEncoding.EncodeToString(img)
		})

	partnerModel.Methods().AddressGet().DeclareMethod(
		`AddressGet finds contacts/addresses of the right type(s) by doing a depth-first-search
        through descendants within company boundaries (stop at entities flagged 'IsCompany')
        then continuing the search at the ancestors that are within the same company boundaries.
        Defaults to partners of type 'default' when the exact type is not found, or to the
        provided partner itself if no type 'default' is found either.

		Result map keys are the contact types, such as 'contact', 'delivery', etc.`,
		func(rs pool.PartnerSet, addrTypes []string) map[string]pool.PartnerSet {
			atMap := make(map[string]bool)
			for _, at := range addrTypes {
				atMap[at] = true
			}
			if _, exists := atMap["contact"]; !exists {
				atMap["contact"] = true
			}
			result := make(map[string]pool.PartnerSet)
			visited := make(map[int64]bool)
			for _, partner := range rs.Records() {
				currentPartner := partner
				for !currentPartner.IsEmpty() {
					toScan := []pool.PartnerSet{currentPartner}
					for len(toScan) > 0 {
						record := toScan[0]
						toScan = toScan[1:]
						visited[record.ID()] = true
						if _, exists := result[record.Type()]; atMap[record.Type()] && !exists {
							result[record.Type()] = record
						}
						if len(result) == len(atMap) {
							return result
						}
						for _, child := range record.Children().Records() {
							if !visited[child.ID()] && !child.IsCompany() {
								toScan = append(toScan, child)
							}
						}
					}
					// Continue scanning at ancestor if current_partner is not a commercial entity
					if currentPartner.IsCompany() || currentPartner.Parent().IsEmpty() {
						break

					}
					currentPartner = currentPartner.Parent()
				}
			}
			return nil
		})

	partnerModel.Methods().DisplayAddress().DeclareMethod(
		`DisplayAddress builds and returns an address formatted accordingly to the
        standards of the country where it belongs.`,
		func(rs pool.PartnerSet, withoutCompany bool) string {
			addressFormat := rs.Country().AddressFormat()
			if addressFormat == "" {
				addressFormat = "{{ .Street }}\n{{ .Street2 }}\n{{ .City }} {{ .StateCode }} {{ .Zip }}\n{{ .CountryName}}"
			}
			data := basetypes.AddressData{
				Street:      rs.Street(),
				Street2:     rs.Street2(),
				City:        rs.City(),
				Zip:         rs.Zip(),
				StateCode:   rs.State().Code(),
				StateName:   rs.State().Name(),
				CountryCode: rs.Country().Code(),
				CountryName: rs.Country().Name(),
				CompanyName: rs.CommercialCompanyName(),
			}
			if data.CompanyName != "" {
				addressFormat = "{{ .CompanyName }}\n" + addressFormat
			}
			addressTemplate := template.Must(template.New("").Parse(addressFormat))
			var buf bytes.Buffer
			err := addressTemplate.Execute(&buf, data)
			if err != nil {
				log.Panic("Error while parsing address", "format", addressFormat, "data", data)
			}
			return buf.String()
		})

}
