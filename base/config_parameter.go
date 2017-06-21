// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	pool.ConfigParameter().DeclareModel()
	pool.ConfigParameter().AddCharField("Key", models.StringFieldParams{Index: true, Required: true, Unique: true})
	pool.ConfigParameter().AddTextField("Value", models.StringFieldParams{Required: true})
	pool.ConfigParameter().AddMany2ManyField("Groups", models.Many2ManyFieldParams{RelationModel: pool.Group()})
}
