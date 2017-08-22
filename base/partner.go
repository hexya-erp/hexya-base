// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/mail"
	"path/filepath"
	"text/template"

	"github.com/hexya-erp/hexya-base/base/basetypes"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/b64image"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
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
	partnerCategory.AddCharField("CompleteName", models.StringFieldParams{String: "Full Name",
		Compute: pool.PartnerCategory().Methods().ComputeCompleteName()})
	partnerCategory.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: pool.PartnerCategory(),
		ReverseFK: "Parent", String: "Children Tags"})
	partnerCategory.AddMany2ManyField("Partners", models.Many2ManyFieldParams{RelationModel: pool.Partner()})

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
	partnerModel.AddCharField("Name", models.StringFieldParams{Required: true, Index: true, NoCopy: true})
	//partnerModel.Fields().DisplayName().SetDepends([]string{"IsCompany", "Name", "Parent.Name", "Type", "CompanyName"})
	partnerModel.AddDateField("Date", models.SimpleFieldParams{Index: true})
	partnerModel.AddMany2OneField("Title", models.ForeignKeyFieldParams{RelationModel: pool.PartnerTitle()})
	partnerModel.AddMany2OneField("Parent", models.ForeignKeyFieldParams{RelationModel: pool.Partner(), Index: true,
		Constraint: pool.Partner().Methods().CheckParent()})
	partnerModel.AddCharField("ParentName", models.StringFieldParams{Related: "Parent.Name"}).
		RevokeAccess(security.GroupEveryone, security.Write)
	partnerModel.AddOne2ManyField("Children", models.ReverseFieldParams{RelationModel: pool.Partner(),
		ReverseFK: "Parent", Filter: pool.Partner().Active().Equals(true)})
	partnerModel.AddCharField("Ref", models.StringFieldParams{String: "Internal Reference", Index: true})
	partnerModel.AddCharField("Lang", models.StringFieldParams{String: "Language",
		Default: func(env models.Environment, maps models.FieldMap) interface{} {
			return env.Context().GetString("lang")
		}, Help: `If the selected language is loaded in the system, all documents related to
this contact will be printed in this language. If not, it will be English.`})
	partnerModel.AddCharField("TZ", models.StringFieldParams{String: "Timezone",
		Default: func(env models.Environment, maps models.FieldMap) interface{} {
			return env.Context().GetString("tz")
		}, Help: `"The partner's timezone, used to output proper date and time values
inside printed reports. It is important to set a value for this field.
You should use the same timezone that is otherwise used to pick and
render date and time values: your computer's timezone.`})
	partnerModel.AddCharField("TZOffset", models.StringFieldParams{Compute: pool.Partner().Methods().ComputeTZOffset(),
		String: "Timezone Offset", Depends: []string{"TZ"}})
	partnerModel.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: pool.User(),
		String: "Salesperson", Help: "The internal user that is in charge of communicating with this contact if any."})
	partnerModel.AddCharField("VAT", models.StringFieldParams{String: "TIN", Help: `Tax Identification Number.
Fill it if the company is subjected to taxes.
Used by the some of the legal statements.`})
	partnerModel.AddOne2ManyField("Banks", models.ReverseFieldParams{RelationModel: pool.Bank(),
		ReverseFK: "Partner"})
	partnerModel.AddCharField("Website", models.StringFieldParams{Help: "Website of Partner or Company"})
	partnerModel.AddCharField("Comment", models.StringFieldParams{String: "Notes"})
	partnerModel.AddMany2ManyField("Categories", models.Many2ManyFieldParams{RelationModel: pool.PartnerCategory(),
		String: "Tags", Default: func(env models.Environment, maps models.FieldMap) interface{} {
			return pool.PartnerCategory().Browse(env, []int64{env.Context().GetInteger("category_id")})
		}})
	partnerModel.AddFloatField("CreditLimit", models.FloatFieldParams{})
	partnerModel.AddCharField("Barcode", models.StringFieldParams{})
	partnerModel.AddBooleanField("Active", models.SimpleFieldParams{Default: models.DefaultValue(true)})
	partnerModel.AddBooleanField("Customer", models.SimpleFieldParams{String: "Is a Customer",
		Default: models.DefaultValue(true), Help: "Check this box if this contact is a customer."})
	partnerModel.AddBooleanField("Supplier", models.SimpleFieldParams{String: "Is a Vendor",
		Help: `Check this box if this contact is a vendor.
If it's not checked, purchase people will not see it when encoding a purchase order.`})
	partnerModel.AddBooleanField("Employee", models.SimpleFieldParams{
		Help: "Check this box if this contact is an Employee."})
	partnerModel.AddCharField("Function", models.StringFieldParams{String: "Job Position"})
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
		Filter: pool.CountryState().Country().EqualsEval("country_id"), OnDelete: models.Restrict})
	partnerModel.AddMany2OneField("Country", models.ForeignKeyFieldParams{RelationModel: pool.Country(),
		OnDelete: models.Restrict})
	partnerModel.AddCharField("Email", models.StringFieldParams{})
	partnerModel.AddCharField("EmailFormatted", models.StringFieldParams{Compute: pool.Partner().Methods().ComputeEmailFormatted(),
		Help: "Formatted email address 'Name <email@domain>'", Depends: []string{"Name", "Email"}})
	partnerModel.AddCharField("Phone", models.StringFieldParams{})
	partnerModel.AddCharField("Fax", models.StringFieldParams{})
	partnerModel.AddCharField("Mobile", models.StringFieldParams{})
	partnerModel.AddBooleanField("IsCompany", models.SimpleFieldParams{Compute: pool.Partner().Methods().ComputeIsCompany(),
		Stored: true, Depends: []string{"CompanyType"}})
	partnerModel.AddSelectionField("CompanyType", models.SelectionFieldParams{
		Selection: types.Selection{"person": "Individual", "company": "Company"},
		OnChange:  pool.Partner().Methods().ComputeIsCompany(), Default: models.DefaultValue("person")})
	partnerModel.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: pool.Company()})
	partnerModel.AddIntegerField("Color", models.SimpleFieldParams{})
	partnerModel.AddOne2ManyField("Users", models.ReverseFieldParams{RelationModel: pool.User(), ReverseFK: "Partner"})
	partnerModel.AddBooleanField("PartnerShare", models.SimpleFieldParams{String: "Share Partner",
		Compute: pool.Partner().Methods().ComputePartnerShare(), Stored: true, Depends: []string{"Users.Share"},
		Help: `Either customer (no user), either shared user. Indicated the current partner is a customer without
access or with a limited access created for sharing data.`})
	partnerModel.AddCharField("ContactAddress", models.StringFieldParams{Compute: pool.Partner().Methods().ComputeContactAddress(),
		String: "Complete Address", Depends: []string{"Street", "Street2", "Zip", "City", "State", "Country",
			"Country.AddressFormat", "Country.Code", "Country.Name", "CompanyName", "State.Code", "State.Name"}})

	partnerModel.AddMany2OneField("CommercialPartner", models.ForeignKeyFieldParams{RelationModel: pool.Partner(),
		Compute: pool.Partner().Methods().ComputeCommercialPartner(), String: "Commercial Entity", Stored: true, Index: true})
	partnerModel.AddCharField("CommercialCompanyName", models.StringFieldParams{
		Compute: pool.Partner().Methods().ComputeCommercialCompanyName(), Stored: true})
	partnerModel.AddCharField("CompanyName", models.StringFieldParams{})

	partnerModel.AddBinaryField("Image", models.SimpleFieldParams{
		Help: "This field holds the image used as avatar for this contact, limited to 1024x1024px"})
	partnerModel.AddBinaryField("ImageMedium", models.SimpleFieldParams{
		Help: `Medium-sized image of this contact. It is automatically
resized as a 128x128px image, with aspect ratio preserved.
Use this field in form views or some kanban views.`})
	partnerModel.AddBinaryField("ImageSmall", models.SimpleFieldParams{
		Help: `Small-sized image of this contact. It is automatically
resized as a 64x64px image, with aspect ratio preserved.
Use this field anywhere a small image is required.`})

	partnerModel.AddSQLConstraint("check_name",
		"CHECK( (type='contact' AND name IS NOT NULL) or (type != 'contact') )",
		"Contacts require a name.")

	partnerModel.Methods().ComputeIsCompany().DeclareMethod(
		`ComputeIsCompany computes the IsCompany field from the selected CompanyType`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			var res pool.PartnerData
			res.IsCompany = rs.CompanyType() == "company"
			return &res, []models.FieldNamer{pool.Partner().IsCompany()}
		})

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

	partnerModel.Methods().ComputeEmailFormatted().DeclareMethod(
		`ComputeEmailFormatted returns a 'Name <email@domain>' formatted string`,
		func(rs pool.PartnerSet) (*pool.PartnerData, []models.FieldNamer) {
			addr := mail.Address{Name: rs.Name(), Address: rs.Email()}
			return &pool.PartnerData{
				EmailFormatted: addr.String(),
			}, []models.FieldNamer{pool.Partner().EmailFormatted()}
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
				path := filepath.Join(generate.HexyaDir, "hexya", "server", "static", "base", "src", "img", imgFileName)
				content, err := ioutil.ReadFile(path)
				if err != nil {
					log.Warn("Missing ressource", "image", imgFileName)
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
		func(rs pool.PartnerSet, defaults models.FieldMapper, fieldsToUnset ...models.FieldNamer) pool.PartnerSet {
			rs.EnsureOne()
			vals := rs.DataStruct(defaults.FieldMap(fieldsToUnset...))
			vals.Name = rs.T("%s (copy)", rs.Name())
			fieldsToUnset = append(fieldsToUnset, pool.Partner().Name())
			return rs.Super().Copy(vals, fieldsToUnset...)
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
