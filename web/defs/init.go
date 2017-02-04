// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/tools/logging"
)

var log log15.Logger

func init() {
	log = logging.GetLogger("web")
	initBaseMixin()
}
