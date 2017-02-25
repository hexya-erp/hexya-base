// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"path"

	"github.com/npiganeau/yep/yep/controllers"
	"github.com/npiganeau/yep/yep/menus"
	"github.com/npiganeau/yep/yep/server"
	"github.com/npiganeau/yep/yep/tools/generate"
)

var (
	// CommonCSS is the list of CSS assets to import by the web client
	// that are common to the frontend and the backend
	CommonCSS []string
	// CommonJS is the list of JavaScript assets to import by the web client
	// that are common to the frontend and the backend
	CommonJS []string
	// CommonCSS is the list of CSS assets to import by the web client
	// that are specific to the backend
	BackendCSS []string
	// CommonJS is the list of JavaScript assets to import by the web client
	// that are specific to the backend
	BackendJS []string
)

type templateData struct {
	Menu      *menus.Collection
	CSS       []string
	BackendJS []string
	CommonJS  []string
	Modules   []string
}

// WebClient is the controller for the application main page
func WebClient(c *server.Context) {
	sess := c.Session()
	sess.Set("uid", int64(1))
	sess.Set("ID", 123)
	sess.Set("login", "admin")
	sess.Save()
	data := templateData{
		Menu:      menus.Registry,
		Modules:   server.Modules.Names(),
		CSS:       append(CommonCSS, BackendCSS...),
		CommonJS:  CommonJS,
		BackendJS: BackendJS,
	}
	c.HTML(http.StatusOK, "web.webclient_bootstrap", data)
}

func init() {
	initStaticPaths()
	initRoutes()
}

func initStaticPaths() {
	CommonCSS = []string{
		"/static/web/lib/jquery.ui/jquery-ui.css",
		"/static/web/lib/fontawesome/css/font-awesome.css",
		"/static/web/src/fonts/lato/stylesheet.css",
		"/static/web/src/less/mimetypes.css",
		"/static/web/src/less/animation.css",
		"/static/web/lib/bootstrap-datetimepicker/css/bootstrap-datetimepicker.css",
		"/static/web/lib/select2/select2.css",
		"/static/web/lib/select2-bootstrap-css/select2-bootstrap.css",
	}
	CommonJS = []string{
		"/static/web/lib/es5-shim/es5-shim.min.js",
		"/static/web/lib/underscore/underscore.js",
		"/static/web/lib/underscore.string/lib/underscore.string.js",
		"/static/web/lib/spinjs/spin.js",
		"/static/web/lib/moment/moment.js",
		"/static/web/lib/autosize/autosize.js",
		"/static/web/lib/jquery/jquery.js",
		"/static/web/lib/jquery.ui/jquery-ui.js",
		"/static/web/lib/jquery/jquery.browser.js",
		"/static/web/lib/jquery.blockUI/jquery.blockUI.js",
		"/static/web/lib/jquery.hotkeys/jquery.hotkeys.js",
		"/static/web/lib/jquery.placeholder/jquery.placeholder.js",
		"/static/web/lib/jquery.timeago/jquery.timeago.js",
		"/static/web/lib/jquery.form/jquery.form.js",
		"/static/web/lib/jquery.ba-bbq/jquery.ba-bbq.js",
		"/static/web/lib/qweb/qweb2.js",
		"/static/web/src/js/boot.js",
		"/static/web/src/js/framework/class.js",
		"/static/web/src/js/framework/translation.js",
		"/static/web/src/js/framework/ajax.js",
		"/static/web/src/js/framework/time.js",
		"/static/web/src/js/framework/mixins.js",
		"/static/web/src/js/framework/widget.js",
		"/static/web/src/js/framework/registry.js",
		"/static/web/src/js/framework/session.js",
		"/static/web/src/js/framework/model.js",
		"/static/web/src/js/framework/utils.js",
		"/static/web/src/js/framework/core.js",
		"/static/web/src/js/framework/dialog.js",
		"/static/web/src/js/tour.js",
		"/static/web/test/menu.js",
		"/static/web/test/x2many.js",
		"/static/web/lib/bootstrap-datetimepicker/src/js/bootstrap-datetimepicker.js",
		"/static/web/lib/select2/select2.js",
	}
	BackendCSS = []string{
		"/static/web/lib/jquery.textext/jquery.textext.css",
		"/static/web/lib/jquery.ui.notify/css/ui.notify.css",
		"/static/web/lib/nvd3/nv.d3.css",
		"/static/web/src/css/base.css",
		"/static/web/src/css/data_export.css",
		"/static/base/src/css/modules.css",
		"/static/web/src/less/import_bootstrap.css",
		"/static/web/src/less/variables.css",
		"/static/web/src/less/enterprise_compatibility.css",
		"/static/web/src/less/utils.css",
		"/static/web/src/less/modal.css",
		"/static/web/src/less/notification.css",
	}
	BackendJS = []string{
		"/static/web/lib/jquery.validate/jquery.validate.js",
		"/static/web/lib/jquery.scrollTo/jquery.scrollTo.js",
		"/static/web/lib/jquery.textext/jquery.textext.js",
		"/static/web/lib/bootstrap/js/affix.js",
		"/static/web/lib/bootstrap/js/alert.js",
		"/static/web/lib/bootstrap/js/button.js",
		"/static/web/lib/bootstrap/js/carousel.js",
		"/static/web/lib/bootstrap/js/collapse.js",
		"/static/web/lib/bootstrap/js/dropdown.js",
		"/static/web/lib/bootstrap/js/modal.js",
		"/static/web/lib/bootstrap/js/tooltip.js",
		"/static/web/lib/bootstrap/js/popover.js",
		"/static/web/lib/bootstrap/js/scrollspy.js",
		"/static/web/lib/bootstrap/js/tab.js",
		"/static/web/lib/bootstrap/js/transition.js",
		"/static/web/lib/jquery.ui.notify/js/jquery.notify.js",
		"/static/web/lib/nvd3/d3.v3.js",
		"/static/web/lib/nvd3/nv.d3.js",
		"/static/web/lib/backbone/backbone.js",
		"/static/web/lib/py.js/lib/py.js",
		"/static/web/lib/jquery.ba-bbq/jquery.ba-bbq.js",
		"/static/web/src/js/framework/data_model.js",
		"/static/web/src/js/framework/formats.js",
		"/static/web/src/js/framework/view.js",
		"/static/web/src/js/framework/pyeval.js",
		"/static/web/src/js/action_manager.js",
		"/static/web/src/js/control_panel.js",
		"/static/web/src/js/view_manager.js",
		"/static/web/src/js/web_client.js",
		"/static/web/src/js/framework/data.js",
		"/static/web/src/js/compatibility.js",
		"/static/web/src/js/framework/misc.js",
		"/static/web/src/js/framework/session_instance.js",
		"/static/web/src/js/framework/crash_manager.js",
		"/static/web/src/js/widgets/change_password.js",
		"/static/web/src/js/widgets/debug_manager.js",
		"/static/web/src/js/widgets/data_export.js",
		"/static/web/src/js/widgets/date_picker.js",
		"/static/web/src/js/widgets/loading.js",
		"/static/web/src/js/widgets/notification.js",
		"/static/web/src/js/widgets/sidebar.js",
		"/static/web/src/js/widgets/priority.js",
		"/static/web/src/js/widgets/progress_bar.js",
		"/static/web/src/js/widgets/pager.js",
		"/static/web/src/js/widgets/systray_menu.js",
		"/static/web/src/js/widgets/user_menu.js",
		"/static/web/src/js/menu.js",
		"/static/web/src/js/views/list_common.js",
		"/static/web/src/js/views/list_view.js",
		"/static/web/src/js/views/form_view.js",
		"/static/web/src/js/views/form_common.js",
		"/static/web/src/js/views/form_widgets.js",
		"/static/web/src/js/views/form_relational_widgets.js",
		"/static/web/src/js/views/list_view_editable.js",
		"/static/web/src/js/views/pivot_view.js",
		"/static/web/src/js/views/graph_view.js",
		"/static/web/src/js/views/graph_widget.js",
		"/static/web/src/js/views/search_view.js",
		"/static/web/src/js/views/search_filters.js",
		"/static/web/src/js/views/search_inputs.js",
		"/static/web/src/js/views/search_menus.js",
		"/static/web/src/js/views/tree_view.js",
		"/static/web/src/js/apps.js",
	}
}

func initRoutes() {
	root := controllers.Registry
	root.AddController(http.MethodGet, "/", func(c *server.Context) {
		c.Redirect(http.StatusSeeOther, "/web")
	})

	root.AddStatic("/static", path.Join(generate.YEPDir, "yep", "server", "static"))
	web := root.AddGroup("/web")
	{
		web.AddController(http.MethodGet, "/", WebClient)
		web.AddController(http.MethodGet, "/image", Image)
		binary := web.AddGroup("/binary")
		{
			binary.AddController(http.MethodGet, "/company_logo", CompanyLogo)
		}

		sess := web.AddGroup("/session")
		{
			sess.AddController(http.MethodPost, "/get_session_info", GetSessionInfo)
			sess.AddController(http.MethodPost, "/modules", Modules)
		}

		proxy := web.AddGroup("/proxy")
		{
			proxy.AddController(http.MethodPost, "/load", Load)
		}

		webClient := web.AddGroup("/webclient")
		{
			webClient.AddController(http.MethodGet, "/qweb", QWeb)
			webClient.AddController(http.MethodGet, "/locale/:lang", LoadLocale)
			webClient.AddController(http.MethodPost, "/translations", BootstrapTranslations)
			webClient.AddController(http.MethodPost, "/csslist", CSSList)
			webClient.AddController(http.MethodPost, "/jslist", JSList)
			webClient.AddController(http.MethodPost, "/version_info", VersionInfo)
		}
		dataset := web.AddGroup("/dataset")
		{
			dataset.AddController(http.MethodPost, "/call_kw/*path", CallKW)
			dataset.AddController(http.MethodPost, "/search_read", SearchRead)
		}
		action := web.AddGroup("/action")
		{
			action.AddController(http.MethodPost, "/load", ActionLoad)
		}
		menu := web.AddGroup("/menu")
		{
			menu.AddController(http.MethodPost, "/load_needaction", MenuLoadNeedaction)
		}
	}
}
