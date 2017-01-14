// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/npiganeau/yep/yep/server"
	"github.com/npiganeau/yep/yep/tools"
)

func Load(c *gin.Context) {
	qwebParams := struct {
		Path string `json:"path"`
	}{}
	server.BindRPCParams(c, &qwebParams)
	path, _ := url.ParseRequestURI(qwebParams.Path)
	targetURL := tools.AbsolutizeURL(c.Request, path.RequestURI())
	resp, err := http.Get(targetURL)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	server.RPC(c, http.StatusOK, string(body))
}
