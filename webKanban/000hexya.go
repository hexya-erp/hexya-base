// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package webKanban

import (
	_ "github.com/hexya-erp/hexya-base/web"
	"github.com/hexya-erp/hexya-base/web/controllers"
	"github.com/hexya-erp/hexya/hexya/server"
)

const MODULE_NAME = "webKanban"

func init() {
	server.RegisterModule(&server.Module{
		Name:     MODULE_NAME,
		PostInit: func() {},
	})

	controllers.BackendLess = append(controllers.BackendLess,
		"/static/webKanban/src/less/kanban_dashboard.less",
		"/static/webKanban/src/less/kanban_view.less",
	)

	controllers.BackendJS = append(controllers.BackendJS,
		"/static/webKanban/src/js/kanban_record.js",
		"/static/webKanban/src/js/kanban_relational.js",
		"/static/webKanban/src/js/kanban_quick_create.js",
		"/static/webKanban/src/js/kanban_column.js",
		"/static/webKanban/src/js/kanban_view.js",
		"/static/webKanban/src/js/kanban_widgets.js",
		"/static/webKanban/src/js/compatibility.js",
	)
}
