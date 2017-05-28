// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/types"
)

func initAttachment() {
	models.NewModel("Attachment")
	attachment := pool.Attachment()
	attachment.AddCharField("Name", models.StringFieldParams{String: "Attachment Name"})
	attachment.AddCharField("DatasFname", models.StringFieldParams{String: "File Name"})
	attachment.AddTextField("Description", models.StringFieldParams{})
	attachment.AddCharField("ResName", models.StringFieldParams{String: "Resource Name"}) //, Compute: "NameGetResName", Stored: true})
	attachment.AddCharField("ResModel", models.StringFieldParams{String: "Resource Model", Help: "The database object this attachment will be attached to"})
	attachment.AddIntegerField("ResID", models.SimpleFieldParams{String: "Resource ID", Help: "The record id this is attached to"})
	attachment.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: "Company"})
	attachment.AddSelectionField("Type", models.SelectionFieldParams{Selection: types.Selection{"binary": "Binary", "url": "URL"}})
	attachment.AddCharField("URL", models.StringFieldParams{})
	attachment.AddBinaryField("Datas", models.SimpleFieldParams{String: "File Content"}) //, Compute: "DataGet"})
	attachment.AddCharField("StoreFname", models.StringFieldParams{String: "Stored Filename"})
	attachment.AddCharField("DBDatas", models.StringFieldParams{String: "Database Data"})
	attachment.AddIntegerField("FileSize", models.SimpleFieldParams{})
}
