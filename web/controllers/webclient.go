// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/npiganeau/yep/yep/server"
	"github.com/npiganeau/yep/yep/tools"
	"github.com/npiganeau/yep/yep/tools/xmlutils"
)

func QWeb(c *gin.Context) {
	mods := strings.Split(c.Query("mods"), ",")
	fileNames := tools.ListStaticFiles("src/xml", mods, true)
	res, _ := xmlutils.ConcatXML(fileNames)
	c.String(http.StatusOK, string(res))
}

func BootstrapTranslations(c *gin.Context) {
	res := gin.H{
		"lang_parameters": tools.LangParameters{
			DateFormat:   "%m/%d/%Y",
			Direction:    tools.LangDirectionLTR,
			ThousandsSep: ",",
			TimeFormat:   "%H:%M:%S",
			DecimalPoint: ".",
			ID:           1,
			Grouping:     "[]",
		},
		"modules": gin.H{},
	}
	server.RPC(c, http.StatusOK, res)
}

func CSSList(c *gin.Context) {
	Params := struct {
		Mods string `json:"mods"`
	}{}
	server.BindRPCParams(c, &Params)
	mods := strings.Split(Params.Mods, ",")
	fileNames := tools.ListStaticFiles("src/css", mods, false)
	server.RPC(c, http.StatusOK, fileNames)
}

func JSList(c *gin.Context) {
	Params := struct {
		Mods string `json:"mods"`
	}{}
	server.BindRPCParams(c, &Params)
	mods := strings.Split(Params.Mods, ",")
	fileNames := tools.ListStaticFiles("src/js", mods, false)
	server.RPC(c, http.StatusOK, fileNames)
}

func VersionInfo(c *gin.Context) {
	data := gin.H{
		"server_serie":        "9.0",
		"server_version_info": []int8{9, 0, 0, 0, 0},
		"server_version":      "9.0c",
		"protocol":            1,
	}
	server.RPC(c, http.StatusOK, data)
}

func LoadLocale(c *gin.Context) {
	// TODO Implement Loadlocale
	//langFull := strings.ToLower(strings.Replace(lang, "_", "-", -1))
	//langShort := strings.Split(lang, "_")[0]
	c.String(http.StatusOK, "OK")
}
