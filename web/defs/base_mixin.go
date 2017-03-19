// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/yep/models"
)

func initBaseMixin() {
	webMixIn := models.NewMixinModel("WebMixin")
	models.MixInAllModels(webMixIn)
}
