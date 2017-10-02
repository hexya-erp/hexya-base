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
	attachment.AddFields(map[string]models.FieldDefinition{
		"Name":        models.CharField{String: "Attachment Name", Required: true},
		"DatasFname":  models.CharField{String: "File Name"},
		"Description": models.TextField{},
		"ResName": models.CharField{String: "Resource Name",
			Compute: pool.Attachment().Methods().ComputeResName(), Stored: true, Depends: []string{"ResModel", "ResID"}},
		"ResModel": models.CharField{String: "Resource Model", Help: "The database object this attachment will be attached to"},
		"ResField": models.CharField{String: "Resource Field"},
		"ResID":    models.IntegerField{String: "Resource ID", Help: "The record id this is attached to"},
		"Company": models.Many2OneField{RelationModel: pool.Company(), Default: func(env models.Environment, fMap models.FieldMap) interface{} {
			currentUser := pool.User().Search(env, pool.User().ID().Equals(env.Uid()))
			return currentUser.Company()
		}},
		"Type":       models.SelectionField{Selection: types.Selection{"binary": "Binary", "url": "URL"}},
		"URL":        models.CharField{},
		"Datas":      models.BinaryField{String: "File Content"}, //, Compute: "DataGet", Inverse: "DataSet"},
		"StoreFname": models.CharField{String: "Stored Filename"},
		"DBDatas":    models.CharField{String: "Database Data"},
		"FileSize":   models.IntegerField{},
		"MimeType":   models.CharField{},
		"Public":     models.BooleanField{String: "Is a public document"},
	})

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
