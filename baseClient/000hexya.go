// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

// Package baseClient provides features on base objects that
// require the web module to be installed.
package baseClient

import (
	_ "github.com/hexya-erp/hexya-base/web"
	"github.com/hexya-erp/hexya/hexya/server"
)

// MODULE_NAME is the name of this module
const MODULE_NAME = "baseClient"

func init() {
	server.RegisterModule(&server.Module{
		Name:     MODULE_NAME,
		PostInit: func() {},
	})
}
