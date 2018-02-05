// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/controllers"
	"github.com/hexya-erp/hexya/hexya/menus"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/hexya/tools/logging"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
)

const (
	commonCSSRoute   = "/web/assets/common.css"
	backendCSSRoute  = "/web/assets/backend.css"
	frontendCSSRoute = "/web/assets/frontend.css"
)

var (
	// CommonLess is the list of Less assets to import by the web client
	// that are common to the frontend and the backend. All less assets are
	// cat'ed together in the given order before being compiled.
	CommonLess []string
	// CommonCSS is the list of CSS files to include without compilation both
	// for the frontend and the backend.
	CommonCSS []string
	// CommonJS is the list of JavaScript assets to import by the web client
	// that are common to the frontend and the backend
	CommonJS []string
	// BackendLess is the list of Less assets to import by the web client
	// that are specific to the backend. All less assets are
	// cat'ed together in the given order before being compiled.
	BackendLess []string
	// BackendCSS is the list of CSS files to include without compilation for
	// the backend.
	BackendCSS []string
	// BackendJS is the list of JavaScript assets to import by the web client
	// that are specific to the backend.
	BackendJS []string
	// FrontendLess is the list of Less assets to import by the web client
	// that are specific to the frontend. All less assets are
	// cat'ed together in the given order before being compiled.
	FrontendLess []string
	// FrontendCSS is the list of CSS files to include without compilation for
	// the frontend.
	FrontendCSS []string
	// FrontendJS is the list of JavaScript assets to import by the web client
	// that are specific to the frontend.
	FrontendJS []string
	// LessHelpers are less files that must be imported for compiling any assets
	LessHelpers []string
)

type templateData struct {
	Menu               []Menu
	CommonCSS          []string
	BackendCSS         []string
	CommonCompiledCSS  string
	BackendCompiledCSS string
	BackendJS          []string
	CommonJS           []string
	Modules            []string
	SessionInfo        gin.H
}

// A Menu is the representation of a single menu item
type Menu struct {
	ID          string
	Name        string
	Children    []Menu
	ActionID    string
	ActionModel string
	HasChildren bool
	HasAction   bool
}

// getMenuTree returns a slice of Menu objects with all their descendants
// from a given slice of menus.Menu objects.
func getMenuTree(menus []*menus.Menu, lang string) []Menu {
	res := make([]Menu, len(menus))
	for i, m := range menus {
		var children []Menu
		if m.HasChildren {
			children = getMenuTree(m.Children.Menus, lang)
		}
		var model string
		if m.HasAction {
			model = m.Action.Model
		}
		name := m.Name
		if lang != "" {
			name = m.TranslatedName(lang)
		}
		res[i] = Menu{
			ID:          m.ID,
			Name:        name,
			ActionID:    m.ActionID,
			ActionModel: model,
			Children:    children,
			HasAction:   m.HasAction,
			HasChildren: m.HasChildren,
		}
	}
	return res
}

// WebClient is the controller for the application main page
func WebClient(c *server.Context) {
	var lang string
	if c.Session().Get("uid") != nil {
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			user := h.User().Search(env, q.User().ID().Equals(c.Session().Get("uid").(int64)))
			lang = user.ContextGet().GetString("lang")
		})
	}
	data := templateData{
		Menu:               getMenuTree(menus.Registry.Menus, lang),
		Modules:            server.Modules.Names(),
		CommonCompiledCSS:  commonCSSRoute,
		BackendCompiledCSS: backendCSSRoute,
		CommonCSS:          CommonCSS,
		BackendCSS:         BackendCSS,
		CommonJS:           CommonJS,
		BackendJS:          BackendJS,
		SessionInfo:        SessionInfo(c.Session()),
	}
	c.HTML(http.StatusOK, "web.webclient_bootstrap", data)
}

func init() {
	log = logging.GetLogger("web/controllers")
	initStaticPaths()
	initRoutes()
	os.Remove(getAssetTempFile(commonCSSRoute))
	os.Remove(getAssetTempFile(backendCSSRoute))
	os.Remove(getAssetTempFile(frontendCSSRoute))
}

func initStaticPaths() {
	LessHelpers = []string{
		"/static/web/lib/bootstrap/less/variables.less",
		"/static/web/lib/bootstrap/less/mixins/vendor-prefixes.less",
		"/static/web/lib/bootstrap/less/mixins/buttons.less",
		"/static/web/src/less/variables.less",
		"/static/web/src/less/utils.less",
	}
	CommonCSS = []string{
		"/static/web/lib/jquery.ui/jquery-ui.css",
		"/static/web/lib/fontawesome/css/font-awesome.css",
		"/static/web/lib/bootstrap-datetimepicker/css/bootstrap-datetimepicker.css",
		"/static/web/lib/select2/select2.css",
		"/static/web/lib/select2-bootstrap-css/select2-bootstrap.css",
	}
	CommonLess = []string{
		"/static/web/src/less/fonts.less",
		"/static/web/src/less/navbar.less",
		"/static/web/src/less/mimetypes.less",
		"/static/web/src/less/animation.less",
	}
	CommonJS = []string{
		"/static/web/lib/es5-shim/es5-shim.min.js",
		"/static/web/lib/underscore/underscore.js",
		"/static/web/lib/underscore.string/lib/underscore.string.js",
		"/static/web/lib/moment/moment.js",
		"/static/web/lib/jquery/jquery.js",
		"/static/web/lib/jquery.ui/jquery-ui.js",
		"/static/web/lib/jquery/jquery.browser.js",
		"/static/web/lib/jquery.blockUI/jquery.blockUI.js",
		"/static/web/lib/jquery.hotkeys/jquery.hotkeys.js",
		"/static/web/lib/jquery.placeholder/jquery.placeholder.js",
		"/static/web/lib/jquery.form/jquery.form.js",
		"/static/web/lib/jquery.ba-bbq/jquery.ba-bbq.js",
		"/static/web/lib/jquery.mjs.nestedSortable/jquery.mjs.nestedSortable.js",
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
		"/static/web/lib/qweb/qweb2.js",
		"/static/web/src/js/boot.js",
		"/static/web/src/js/config.js",
		"/static/web/src/js/framework/class.js",
		"/static/web/src/js/framework/translation.js",
		"/static/web/src/js/framework/ajax.js",
		"/static/web/src/js/framework/time.js",
		"/static/web/src/js/framework/mixins.js",
		"/static/web/src/js/framework/widget.js",
		"/static/web/src/js/framework/registry.js",
		"/static/web/src/js/framework/session.js",
		"/static/web/src/js/framework/model.js",
		"/static/web/src/js/framework/dom_utils.js",
		"/static/web/src/js/framework/utils.js",
		"/static/web/src/js/framework/qweb.js",
		"/static/web/src/js/framework/bus.js",
		"/static/web/src/js/services/core.js",
		"/static/web/src/js/framework/dialog.js",
		"/static/web/src/js/framework/local_storage.js",
		"/static/web/lib/bootstrap-datetimepicker/src/js/bootstrap-datetimepicker.js",
		"/static/web/lib/select2/select2.js",
	}
	BackendCSS = []string{
		"/static/web/lib/nvd3/nv.d3.css",
	}
	BackendLess = []string{
		"/static/web/src/less/import_bootstrap.less",
		"/static/web/src/less/bootstrap_overridden.less",
		"/static/web/src/less/webclient_extra.less",
		"/static/web/src/less/webclient_layout.less",
		"/static/web/src/less/webclient.less",
		"/static/web/src/less/datepicker.less",
		"/static/web/src/less/progress_bar.less",
		"/static/web/src/less/dropdown.less",
		"/static/web/src/less/tooltip.less",
		"/static/web/src/less/debug_manager.less",
		"/static/web/src/less/control_panel.less",
		"/static/web/src/less/control_panel_layout.less",
		"/static/web/src/less/views.less",
		"/static/web/src/less/pivot_view.less",
		"/static/web/src/less/graph_view.less",
		"/static/web/src/less/tree_view.less",
		"/static/web/src/less/form_view_layout.less",
		"/static/web/src/less/form_view.less",
		"/static/web/src/less/list_view.less",
		"/static/web/src/less/search_view.less",
		"/static/web/src/less/modal.less",
		"/static/web/src/less/data_export.less",
		"/static/web/src/less/switch_company_menu.less",
		"/static/web/src/less/dropdown_extra.less",
		"/static/web/src/less/views_extra.less",
		"/static/web/src/less/form_view_extra.less",
		"/static/web/src/less/form_view_layout_extra.less",
		"/static/web/src/less/search_view_extra.less",
		"/static/web/src/less/bootswatch.less",
	}
	BackendJS = []string{
		"/static/web/lib/jquery.scrollTo/jquery.scrollTo.js",
		"/static/web/lib/nvd3/d3.v3.js",
		"/static/web/lib/nvd3/nv.d3.js",
		"/static/web/lib/backbone/backbone.js",
		"/static/web/lib/fuzzy-master/fuzzy.js",
		"/static/web/lib/py.js/lib/py.js",
		"/static/web/lib/jquery.ba-bbq/jquery.ba-bbq.js",
		"/static/web/src/js/framework/data_model.js",
		"/static/web/src/js/framework/formats.js",
		"/static/web/src/js/framework/view.js",
		"/static/web/src/js/framework/pyeval.js",
		"/static/web/src/js/action_manager.js",
		"/static/web/src/js/control_panel.js",
		"/static/web/src/js/view_manager.js",
		"/static/web/src/js/abstract_web_client.js",
		"/static/web/src/js/web_client.js",
		"/static/web/src/js/framework/data.js",
		"/static/web/src/js/compatibility.js",
		"/static/web/src/js/framework/misc.js",
		"/static/web/src/js/framework/crash_manager.js",
		"/static/web/src/js/framework/data_manager.js",
		"/static/web/src/js/services/crash_manager.js",
		"/static/web/src/js/services/data_manager.js",
		"/static/web/src/js/services/session.js",
		"/static/web/src/js/widgets/auto_complete.js",
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
		"/static/web/src/js/widgets/switch_company_menu.js",
		"/static/web/src/js/widgets/user_menu.js",
		"/static/web/src/js/menu.js",
		"/static/web/src/js/views/list_common.js",
		"/static/web/src/js/views/list_view.js",
		"/static/web/src/js/views/form_view.js",
		"/static/web/src/js/views/form_common.js",
		"/static/web/src/js/views/form_widgets.js",
		"/static/web/src/js/views/form_upgrade_widgets.js",
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
	FrontendLess = []string{
		"/static/web/src/less/import_bootstrap.less",
		"/static/web/src/less/bootstrap_overridden.less",
		"/static/web/src/less/bootswatch.less",
	}
	FrontendCSS = []string{}
	FrontendJS = []string{
		"/static/web/src/js/services/session.js",
	}
}

func initRoutes() {
	root := controllers.Registry
	root.AddController(http.MethodGet, "/", func(c *server.Context) {
		c.Redirect(http.StatusSeeOther, "/web")
	})
	root.AddController(http.MethodGet, "/web/login", LoginGet)
	root.AddController(http.MethodPost, "/web/login", LoginPost)
	root.AddController(http.MethodGet, "/web/binary/company_logo", CompanyLogo)
	assets := root.AddGroup("/web/assets")
	{
		assets.AddController(http.MethodGet, "/common.css", AssetsCommonCSS)
		assets.AddController(http.MethodGet, "/backend.css", AssetsBackendCSS)
		assets.AddController(http.MethodGet, "/frontend.css", AssetsFrontendCSS)
	}

	root.AddStatic("/static", filepath.Join(generate.HexyaDir, "hexya", "server", "static"))
	web := root.AddGroup("/web")
	{
		web.AddMiddleWare(LoginRequired)
		web.AddController(http.MethodGet, "/", WebClient)
		web.AddController(http.MethodGet, "/image", Image)

		sess := web.AddGroup("/session")
		{
			sess.AddController(http.MethodPost, "/get_session_info", GetSessionInfo)
			sess.AddController(http.MethodPost, "/modules", Modules)
			sess.AddController(http.MethodGet, "/logout", Logout)
			sess.AddController(http.MethodPost, "/change_password", ChangePassword)
		}

		proxy := web.AddGroup("/proxy")
		{
			proxy.AddController(http.MethodPost, "/load", Load)
		}

		webClient := web.AddGroup("/webclient")
		{
			webClient.AddController(http.MethodGet, "/qweb", QWeb)
			webClient.AddController(http.MethodGet, "/locale", LoadLocale)
			webClient.AddController(http.MethodGet, "/locale/:lang", LoadLocale)
			webClient.AddController(http.MethodPost, "/translations", BootstrapTranslations)
			webClient.AddController(http.MethodPost, "/bootstrap_translations", BootstrapTranslations)
			webClient.AddController(http.MethodPost, "/csslist", CSSList)
			webClient.AddController(http.MethodPost, "/jslist", JSList)
			webClient.AddController(http.MethodPost, "/version_info", VersionInfo)
		}
		dataset := web.AddGroup("/dataset")
		{
			dataset.AddController(http.MethodPost, "/call_kw/*path", CallKW)
			dataset.AddController(http.MethodPost, "/search_read", SearchRead)
			dataset.AddController(http.MethodPost, "/call_button", CallButton)
		}
		action := web.AddGroup("/action")
		{
			action.AddController(http.MethodPost, "/load", ActionLoad)
			action.AddController(http.MethodPost, "/run", ActionRun)
		}
		menu := web.AddGroup("/menu")
		{
			menu.AddController(http.MethodPost, "/load_needaction", MenuLoadNeedaction)
		}
	}
}
