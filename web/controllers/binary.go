// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/hexya-erp/hexya/hexya/menus"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
	"github.com/hexya-erp/hexya/pool/h"
)

// CompanyLogo serves the logo of the company
func CompanyLogo(c *server.Context) {
	info := GetSessionInfo(c.Session())
	var img string
	switch {
	case info == nil:
		// Not connected. Get image of administrator company
		models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			img = h.User().NewSet(env).Browse([]int64{security.SuperUserID}).Company().LogoWeb()
		})
	default:
		// Connected. Get image of session's company
		models.ExecuteInNewEnvironment(info.UID, func(env models.Environment) {
			img = h.Company().NewSet(env).Browse([]int64{info.CompanyID}).LogoWeb()
		})
	}
	res, err := base64.StdEncoding.DecodeString(img)
	if err != nil || img == "" {
		c.File(filepath.Join(generate.HexyaDir, "hexya", "server", "static", "web", "src", "img", "nologo.png"))
		return
	}
	c.Data(http.StatusOK, "image/png", res)
}

// Image serves the image stored in the database (base64 encoded)
// in the given model and given field
func Image(c *server.Context) {
	model := c.Query("model")
	field := c.Query("field")
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	uid := c.Session().Get("uid").(int64)
	img, gErr := getFieldValue(uid, id, model, field)
	if gErr != nil {
		c.Error(fmt.Errorf("unable to fetch image: %s", gErr))
		return
	}
	if img.(string) == "" {
		c.File(filepath.Join(generate.HexyaDir, "hexya", "server", "static", "web", "src", "img", "placeholder.png"))
		return
	}
	res, err := base64.StdEncoding.DecodeString(img.(string))
	if err != nil {
		c.Error(fmt.Errorf("unable to convert image: %s", err))
		return
	}
	c.Data(http.StatusOK, "image/png", res)
}

// MenuImage serves the image for the given menu
func MenuImage(c *server.Context) {
	menuID := c.Param("menu_id")
	menu := menus.Registry.GetByID(menuID)
	if menu != nil && menu.WebIcon != "" {
		fp := filepath.Join(generate.HexyaDir, "hexya", "server", menu.WebIcon)
		c.File(fp)
		return
	}
	c.File(filepath.Join(generate.HexyaDir, "hexya", "server", "static", "web", "src", "img", "placeholder.png"))
}
