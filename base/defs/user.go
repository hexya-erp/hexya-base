// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/models"
)

func initUsers() {
	resUsers := models.NewModel("ResUsers")
	resUsers.AddDateTimeField("LoginDate", models.SimpleFieldParams{})
	resUsers.AddMany2OneField("Partner", models.ForeignKeyFieldParams{RelationModel: "ResPartner", Embed: true})
	resUsers.AddCharField("Name", models.StringFieldParams{})
	resUsers.AddCharField("Login", models.StringFieldParams{})
	resUsers.AddCharField("Password", models.StringFieldParams{})
	resUsers.AddCharField("NewPassword", models.StringFieldParams{})
	resUsers.AddTextField("Signature", models.StringFieldParams{})
	resUsers.AddBooleanField("Active", models.SimpleFieldParams{})
	resUsers.AddCharField("ActionID", models.StringFieldParams{GoType: new(actions.ActionRef)})
	resUsers.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: "ResCompany"})
	resUsers.AddMany2ManyField("Companies", models.Many2ManyFieldParams{RelationModel: "ResCompany", JSON: "company_ids"})
	resUsers.AddBinaryField("ImageSmall", models.SimpleFieldParams{})
	//GroupIds []*ir.Group `yep:"json(groups_id)"`
}
