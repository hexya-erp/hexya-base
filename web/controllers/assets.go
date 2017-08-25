// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package controllers

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hexya-erp/hexya/hexya/server"
	"github.com/hexya-erp/hexya/hexya/tools/assets"
	"github.com/hexya-erp/hexya/hexya/tools/generate"
)

func getAssetTempFile(asset string) string {
	return filepath.Join(os.TempDir(), strings.Replace(asset, "/", "_", -1))
}

func createCSSAssets(files []string, asset string, includePaths ...string) {
	var readers []io.Reader
	for _, fileName := range files {
		filePath := filepath.Join(generate.HexyaDir, "hexya", "server", fileName)
		f, err := os.Open(filePath)
		if err != nil {
			log.Panic("Error while reading less file", "filename", fileName, "error", err)
		}
		readers = append(readers, f, strings.NewReader("\n"))
	}
	tmpFile, err := os.Create(getAssetTempFile(asset))
	defer tmpFile.Close()
	if err != nil {
		log.Panic("Error while opening asset file", "error", err)
	}
	err = assets.CompileLess(io.MultiReader(readers...), tmpFile, includePaths...)
	if err != nil {
		tmpFile.Close()
		os.Remove(getAssetTempFile(asset))
		log.Panic("Error while generating asset file", "error", err)
	}
}

// AssetsCommonCSS returns the compiled CSS for the common assets
func AssetsCommonCSS(c *server.Context) {
	fName := getAssetTempFile(commonCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		createCSSAssets(append(LessHelpers, CommonLess...), commonCSSRoute)
	}
	c.File(fName)
}

// AssetsBackendCSS returns the compiled CSS for the backend assets
func AssetsBackendCSS(c *server.Context) {
	fName := getAssetTempFile(backendCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		bootstrapDir := filepath.Join(generate.HexyaDir, "hexya", "server", "static", "web", "lib", "bootstrap", "less")
		createCSSAssets(append(LessHelpers, BackendLess...), backendCSSRoute, bootstrapDir)
	}
	c.File(fName)
}

// AssetsFrontendCSS returns the compiled CSS for the frontend assets
func AssetsFrontendCSS(c *server.Context) {
	fName := getAssetTempFile(frontendCSSRoute)
	if _, err := os.Stat(fName); err != nil {
		bootstrapDir := filepath.Join(generate.HexyaDir, "hexya", "server", "static", "web", "lib", "bootstrap", "less")
		createCSSAssets(append(LessHelpers, FrontendLess...), frontendCSSRoute, bootstrapDir)
	}
	c.File(fName)
}
