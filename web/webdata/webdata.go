// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package webdata

import (
	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/views"
)

// FieldsViewGetParams is the args struct for the FieldsViewGet function
type FieldsViewGetParams struct {
	ViewID   string `json:"view_id"`
	ViewType string `json:"view_type"`
	Toolbar  bool   `json:"toolbar"`
}

// FieldsViewData is the return type string for the FieldsViewGet function
type FieldsViewData struct {
	Name        string                       `json:"name"`
	Arch        string                       `json:"arch"`
	ViewID      string                       `json:"view_id"`
	Model       string                       `json:"model"`
	Type        views.ViewType               `json:"type"`
	Fields      map[string]*models.FieldInfo `json:"fields"`
	Toolbar     Toolbar                      `json:"toolbar"`
	FieldParent string                       `json:"field_parent"`
}

// SearchParams is the args struct for the SearchRead method
type SearchParams struct {
	Domain models.Domain `json:"domain"`
	Fields []string      `json:"fields"`
	Offset int           `json:"offset"`
	Limit  interface{}   `json:"limit"`
	Order  string        `json:"order"`
}

// A Toolbar holds the actions in the toolbar of the action manager
type Toolbar struct {
	Print  []*actions.BaseAction `json:"print"`
	Action []*actions.BaseAction `json:"action"`
	Relate []*actions.BaseAction `json:"relate"`
}
