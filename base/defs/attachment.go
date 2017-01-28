// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
)

func initAttachment() {
	models.NewModel("IrAttachment", new(struct {
		ID          int64
		Name        string `yep:"string(Attachment Name)"`
		DatasFname  string `yep:"string(File Name)"`
		Description string
		//ResName     string      `yep:"string(Resource Name);compute(NameGetResName);store(true)"`
		ResModel string             `yep:"string(Resource Model);help(The database object this attachment will be attached to)"`
		ResId    int64              `yep:"string(Resource ID);help(The record id this is attached to)"`
		Company  pool.ResCompanySet `yep:"type(many2one)"`
		Type     string             `yep:"help(Binary File or URL)"`
		Url      string
		//Datas       string      `yep:"compute(DataGet);string(File Content)"`
		StoreFname string `yep:"string(Stored Filename)"`
		DbDatas    string `yep:"string(Database Data)"`
		FileSize   int    `yep:"string(File Size)"`
	}))
}
