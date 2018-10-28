// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/json"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/menus"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools"
	"github.com/hexya-erp/hexya/hexya/tools/hweb"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
)

// QWeb returns a concatenation of all client qweb templates
func QWeb(c *server.Context) {
	mods := strings.Split(c.Query("mods"), ",")
	fileNames := tools.ListStaticFiles("src/xml", mods, true)
	res, _, err := xmlutils.ConcatXML(fileNames)
	if err != nil {
		c.Error(fmt.Errorf("error while generating client side QWeb: %s", err.Error()))
	}
	c.String(http.StatusOK, string(res))
}

// BootstrapTranslations returns data about the current language
func BootstrapTranslations(c *server.Context) {
	res := gin.H{
		"lang_parameters": i18n.LangParameters{
			DateFormat:   "%m/%d/%Y",
			Direction:    i18n.LangDirectionLTR,
			ThousandsSep: ",",
			TimeFormat:   "%H:%M:%S",
			DecimalPoint: ".",
			ID:           1,
			Grouping:     "[]",
		},
		"modules":    gin.H{},
		"multi_lang": true,
	}
	c.RPC(http.StatusOK, res)
}

// CSSList returns the list of CSS files
func CSSList(c *server.Context) {
	Params := struct {
		Mods string `json:"mods"`
	}{}
	c.BindRPCParams(&Params)
	mods := strings.Split(Params.Mods, ",")
	fileNames := tools.ListStaticFiles("src/css", mods, false)
	c.RPC(http.StatusOK, fileNames)
}

// JSList returns the list of JS files
func JSList(c *server.Context) {
	Params := struct {
		Mods string `json:"mods"`
	}{}
	c.BindRPCParams(&Params)
	mods := strings.Split(Params.Mods, ",")
	fileNames := tools.ListStaticFiles("src/js", mods, false)
	c.RPC(http.StatusOK, fileNames)
}

// VersionInfo returns server version information to the client
func VersionInfo(c *server.Context) {
	data := gin.H{
		"server_serie":        "0.9beta",
		"server_version_info": []int8{0, 9, 0, 0, 0},
		"server_version":      "0.9beta",
		"protocol":            1,
	}
	c.RPC(http.StatusOK, data)
}

// LoadLocale returns the locale's JS file
func LoadLocale(c *server.Context) {
	// TODO Implement Loadlocale
	//langFull := strings.ToLower(strings.Replace(lang, "_", "-", -1))
	//langShort := strings.Split(lang, "_")[0]
	c.String(http.StatusOK, "OK")
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

// getMenuTree returns a tree of menus with all their descendants
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
	rootMenu := Menu{
		Name:        "root",
		Children:    getMenuTree(menus.Registry.Menus, lang),
		HasAction:   false,
		HasChildren: true,
	}

	siBytes, err := json.Marshal(SessionInfo(c.Session()))
	if err != nil {
		c.Error(err)
		return
	}
	modBytes, err := json.Marshal(server.Modules.Names())
	if err != nil {
		c.Error(err)
		return
	}

	data := hweb.Context{
		"menu_data":          rootMenu,
		"modules":            string(modBytes),
		"session_info":       string(siBytes),
		"debug":              false,
		"commonCSS":          CommonCSS,
		"commonCompiledCSS":  commonCSSRoute,
		"commonJS":           CommonJS,
		"backendCSS":         BackendCSS,
		"backendCompiledCSS": backendCSSRoute,
		"backendJS":          BackendJS,
	}
	templateName := strings.TrimPrefix(path.Join(lang, "web.webclient_bootstrap"), "/")
	c.HTML(http.StatusOK, templateName, data)
}
