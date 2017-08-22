// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	attachment := pool.Attachment().DeclareModel()
	attachment.AddCharField("Name", models.StringFieldParams{String: "Attachment Name", Required: true})
	attachment.AddCharField("DatasFname", models.StringFieldParams{String: "File Name"})
	attachment.AddTextField("Description", models.StringFieldParams{})
	attachment.AddCharField("ResName", models.StringFieldParams{String: "Resource Name",
		Compute: pool.Attachment().Methods().ComputeResName(), Stored: true, Depends: []string{"ResModel", "ResID"}})
	attachment.AddCharField("ResModel", models.StringFieldParams{String: "Resource Model", Help: "The database object this attachment will be attached to"})
	attachment.AddIntegerField("ResField", models.SimpleFieldParams{String: "Resource Field"})
	attachment.AddIntegerField("ResID", models.SimpleFieldParams{String: "Resource ID", Help: "The record id this is attached to"})
	attachment.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: pool.Company(), Default: func(env models.Environment, fMap models.FieldMap) interface{} {
		currentUser := pool.User().Search(env, pool.User().ID().Equals(env.Uid()))
		return currentUser.Company()
	}})
	attachment.AddSelectionField("Type", models.SelectionFieldParams{Selection: types.Selection{"binary": "Binary", "url": "URL"}})
	attachment.AddCharField("URL", models.StringFieldParams{})
	attachment.AddBinaryField("Datas", models.SimpleFieldParams{String: "File Content"}) //, Compute: "DataGet", Inverse: "DataSet"})
	attachment.AddCharField("StoreFname", models.StringFieldParams{String: "Stored Filename"})
	attachment.AddCharField("DBDatas", models.StringFieldParams{String: "Database Data"})
	attachment.AddIntegerField("FileSize", models.SimpleFieldParams{})
	attachment.AddCharField("MimeType", models.StringFieldParams{})
	attachment.AddBooleanField("Public", models.SimpleFieldParams{String: "Is a public document"})

	attachment.Methods().ComputeResName().DeclareMethod(
		`ComputeResName computes the display name of the ressource this document is attached to.`,
		func(rs pool.AttachmentSet) (*pool.AttachmentData, []models.FieldNamer) {
			var res pool.AttachmentData
			if rs.ResModel() != "" && rs.ResID() != 0 {
				record := rs.Env().Pool(rs.ResModel()).Search(models.Registry.MustGet(rs.ResModel()).Field("ID").Equals(rs.ResID()))
				res.ResName = record.Get("DisplayName").(string)
			}
			return &res, []models.FieldNamer{pool.Attachment().ResName()}
		})

}
