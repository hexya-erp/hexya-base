// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/npiganeau/yep/yep/server"
	"github.com/npiganeau/yep/yep/tools"
)

func Load(c *server.Context) {
	qwebParams := struct {
		Path string `json:"path"`
	}{}
	c.BindRPCParams(&qwebParams)
	path, _ := url.ParseRequestURI(qwebParams.Path)
	targetURL := tools.AbsolutizeURL(c.Request, path.RequestURI())
	resp, err := http.Get(targetURL)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	c.RPC(http.StatusOK, string(body))
}
