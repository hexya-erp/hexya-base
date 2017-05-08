// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package methods

import "github.com/npiganeau/yep/yep/tools/logging"

var log *logging.Logger

func init() {
	log = logging.GetLogger("base")
	initGroups()
	initFilters()
	initUsers()
	initPartner()
}
