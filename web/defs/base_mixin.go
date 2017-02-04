// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/npiganeau/yep/yep/models"
)

func initBaseMixin() {
	models.NewMixinModel("WebMixin", new(struct{}))
	models.MixInAllModels("WebMixin")
}
