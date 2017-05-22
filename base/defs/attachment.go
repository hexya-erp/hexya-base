// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/types"
)

func initAttachment() {
	irAttachment := models.NewModel("IrAttachment")
	irAttachment.AddCharField("Name", models.StringFieldParams{String: "Attachment Name"})
	irAttachment.AddCharField("DatasFname", models.StringFieldParams{String: "File Name"})
	irAttachment.AddTextField("Description", models.StringFieldParams{})
	irAttachment.AddCharField("ResName", models.StringFieldParams{String: "Resource Name"}) //, Compute: "NameGetResName", Stored: true})
	irAttachment.AddCharField("ResModel", models.StringFieldParams{String: "Resource Model", Help: "The database object this attachment will be attached to"})
	irAttachment.AddIntegerField("ResID", models.SimpleFieldParams{String: "Resource ID", Help: "The record id this is attached to"})
	irAttachment.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: "ResCompany"})
	irAttachment.AddSelectionField("Type", models.SelectionFieldParams{Selection: types.Selection{"binary": "Binary", "url": "URL"}})
	irAttachment.AddCharField("URL", models.StringFieldParams{})
	irAttachment.AddBinaryField("Datas", models.SimpleFieldParams{String: "File Content"}) //, Compute: "DataGet"})
	irAttachment.AddCharField("StoreFname", models.StringFieldParams{String: "Stored Filename"})
	irAttachment.AddCharField("DBDatas", models.StringFieldParams{String: "Database Data"})
	irAttachment.AddIntegerField("FileSize", models.SimpleFieldParams{})
}
