// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"strings"

	"fmt"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/i18n"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools"
	"github.com/hexya-erp/hexya/hexya/tools/xmlutils"
)

// QWeb returns a concatenation of all client qweb templates
func QWeb(c *server.Context) {
	mods := strings.Split(c.Query("mods"), ",")
	fileNames := tools.ListStaticFiles("src/xml", mods, true)
	res, _ := xmlutils.ConcatXML(fileNames)
	c.String(http.StatusOK, string(res))
}

// BootstrapTranslations returns data about the current language
func BootstrapTranslations(c *server.Context) {
	params := struct {
		Lang    string   `json:"lang"`
		Modules []string `json:"mods"`
	}{}
	c.BindRPCParams(&params)
	res := gin.H{
		"lang_parameters": i18n.GetLangParameters(params.Lang),
		"modules":         i18n.ListModuleTranslations(params.Lang),
		"multi_lang":      true,
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
	lang := c.Param("lang")
	var outstr string
	langFull := strings.ToLower(strings.Replace(lang, "_", "-", -1))
	path := fmt.Sprintf("%s/src/github.com/hexya-erp/hexya/hexya/server/static/web/lib/moment/locale/%s.js", os.Getenv("GOPATH"), langFull)
	content, err := ioutil.ReadFile(path)
	var err2 error
	if err != nil {
		langShort := strings.Split(lang, "_")[0]
		path := fmt.Sprintf("%s/src/github.com/hexya-erp/hexya/hexya/server/static/web/lib/moment/locale/%s.js", os.Getenv("GOPATH"), langShort)
		content, err2 = ioutil.ReadFile(path)
	}
	if len(content) > 2 {
		outstr = string(content)
	} else {
		outstr = fmt.Sprintf("LOCALE NOT FOUND FOR '%s'\n%s\n%s", lang, err, err2)
	}
	c.Header("Content-Type", "application/javascript")
	c.String(http.StatusOK, outstr)

}
