// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import "github.com/npiganeau/yep/yep/models"

func initGroups() {
	resGroups := models.NewModel("ResGroups")
	resGroups.AddCharField("GroupID", models.StringFieldParams{Required: true})
	resGroups.AddCharField("Name", models.StringFieldParams{Required: true, Translate: true})
}
