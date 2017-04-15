// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package methods

import (
	"encoding/json"
	"strings"

	"github.com/npiganeau/yep-base/web/webdata"
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/types"
	"github.com/npiganeau/yep/yep/tools/etree"
	"github.com/npiganeau/yep/yep/tools/logging"
	"github.com/npiganeau/yep/yep/views"
)

func createMixinMethods() {
	commonMixin := pool.CommonMixin()

	commonMixin.ExtendMethod("Create", "",
		func(rc models.RecordCollection, data interface{}) models.RecordCollection {
			fMap := rc.Call("ProcessDataValues", data).(models.FieldMap)
			res := rc.Super().Call("Create", fMap).(models.RecordCollection)
			return res
		})

	commonMixin.ExtendMethod("Write", "",
		func(rs pool.CommonMixinSet, data interface{}, fieldsToUnset ...models.FieldNamer) bool {
			fMap := rs.ProcessDataValues(data)
			res := rs.Super().Write(fMap, fieldsToUnset...)
			return res
		})

	commonMixin.AddMethod("ProcessDataValues",
		`ProcessDataValues updates the given data values for Write and Create methods to be
		compatible with the ORM`,
		func(rs pool.CommonMixinSet, data interface{}) models.FieldMap {
			fMap := models.ConvertInterfaceToFieldMap(data)
			fInfos := rs.FieldsGet(models.FieldsGetArgs{})
			for f, v := range fMap {
				fJSON := rs.Model().JSONizeFieldName(f)
				if _, exists := fInfos[fJSON]; !exists {
					logging.LogAndPanic(log, "Unable to find field", "model", rs.ModelName(), "field", f)
				}
				switch fInfos[fJSON].Type {
				case types.Many2Many:
					fMap[f] = rs.NormalizeM2MData(f, fInfos[fJSON], v)
				}
			}
			return fMap
		})

	commonMixin.AddMethod("NormalizeM2MData",
		`NormalizeM2MData converts the list of triplets received from the client into the final list of ids
		to keep in the Many2Many relationship of this model through the given field.`,
		func(rs pool.BaseMixinSet, fieldName string, info *models.FieldInfo, value interface{}) interface{} {
			switch v := value.(type) {
			case []interface{}:
				resSet := rs.Env().Pool(info.Relation)
				if len(v) == 0 {
					return resSet
				}
				// We assume we have a list of triplets from client
				for _, triplet := range v {
					// TODO manage effectively multi-tuple input
					action := int(triplet.([]interface{})[0].(float64))
					switch action {
					case 0:
					case 1:
					case 2:
					case 3:
					case 4:
					case 5:
					case 6:
						idList := triplet.([]interface{})[2].([]interface{})
						ids := make([]int64, len(idList))
						for i, id := range idList {
							ids[i] = int64(id.(float64))
						}
						return resSet.Search(resSet.Model().Field("ID").In(ids))
					}
				}
			}
			return value
		})

	commonMixin.AddMethod("GetFormviewId",
		`GetFormviewId returns an view id to open the document with.
		This method is meant to be overridden in addons that want
 		to give specific view ids for example.`,
		func(rc models.RecordCollection) string {
			return ""
		})

	commonMixin.AddMethod("GetFormviewAction",
		`GetFormviewAction returns an action to open the document.
		This method is meant to be overridden in addons that want
		to give specific view ids for example.`,
		func(rc models.RecordCollection) *actions.BaseAction {
			viewID := rc.Call("GetFormviewId").(string)
			return &actions.BaseAction{
				Type:        actions.ActionActWindow,
				Model:       rc.ModelName(),
				ActViewType: actions.ActionViewTypeForm,
				ViewMode:    "form",
				Views:       []views.ViewTuple{{ID: viewID, Type: views.VIEW_TYPE_FORM}},
				Target:      "current",
				ResID:       rc.Get("id").(int64),
				Context:     rc.Env().Context(),
			}
		})

	commonMixin.AddMethod("FieldsViewGet",
		`FieldsViewGet is the base implementation of the 'FieldsViewGet' method which
		gets the detailed composition of the requested view like fields, mixin,
		view architecture.`,
		func(rc models.RecordCollection, args webdata.FieldsViewGetParams) *webdata.FieldsViewData {
			view := views.Registry.GetByID(args.ViewID)
			if view == nil {
				view = views.Registry.GetFirstViewForModel(rc.ModelName(), views.ViewType(args.ViewType))
			}
			cols := make([]models.FieldName, len(view.Fields))
			for i, f := range view.Fields {
				cols[i] = models.FieldName(rc.Model().JSONizeFieldName(string(f)))
			}
			fInfos := rc.Call("FieldsGet", models.FieldsGetArgs{Fields: cols}).(map[string]*models.FieldInfo)
			arch := rc.Call("ProcessView", view.Arch, fInfos).(string)
			toolbar := rc.Call("GetToolbar").(webdata.Toolbar)
			res := webdata.FieldsViewData{
				Name:    view.Name,
				Arch:    arch,
				ViewID:  args.ViewID,
				Model:   view.Model,
				Type:    view.Type,
				Toolbar: toolbar,
				Fields:  fInfos,
			}
			return &res
		})

	commonMixin.AddMethod("GetToolbar",
		`GetToolbar returns a toolbar populated with the actions linked to this model`,
		func(rs pool.BaseMixinSet) webdata.Toolbar {
			var res webdata.Toolbar
			for _, a := range actions.Registry.GetActionLinksForModel(rs.ModelName()) {
				switch a.Type {
				case actions.ActionActWindow, actions.ActionServer:
					res.Action = append(res.Action, a)
				}
			}
			return res
		})

	commonMixin.AddMethod("ProcessView",
		`ProcessView makes all the necessary modifications to the view
		arch and returns the new xml string.`,
		func(rc models.RecordCollection, arch string, fieldInfos map[string]*models.FieldInfo) string {
			// Load arch as etree
			doc := etree.NewDocument()
			if err := doc.ReadFromString(arch); err != nil {
				logging.LogAndPanic(log, "Unable to parse view arch", "arch", arch, "error", err)
			}
			// Apply changes
			rc.Call("UpdateFieldNames", doc, &fieldInfos)
			rc.Call("AddModifiers", doc, fieldInfos)
			// Dump xml to string and return
			res, err := doc.WriteToString()
			if err != nil {
				logging.LogAndPanic(log, "Unable to render XML", "error", err)
			}
			return res
		})

	commonMixin.AddMethod("AddModifiers",
		`AddModifiers adds the modifiers attribute nodes to given xml doc.`,
		func(rc models.RecordCollection, doc *etree.Document, fieldInfos map[string]*models.FieldInfo) {
			allModifiers := make(map[*etree.Element]map[string]interface{})
			// Process attrs on all nodes
			for _, attrsTag := range doc.FindElements("[@attrs]") {
				allModifiers[attrsTag] = rc.Call("ProcessElementAttrs", attrsTag).(map[string]interface{})
			}
			// Process field nodes
			for _, fieldTag := range doc.FindElements("//field") {
				mods, exists := allModifiers[fieldTag]
				if !exists {
					mods = map[string]interface{}{"readonly": false, "required": false, "invisible": false}

				}
				allModifiers[fieldTag] = rc.Call("ProcessFieldElementModifiers", fieldTag, fieldInfos, mods).(map[string]interface{})
			}
			// Set modifier attributes on elements
			for element, modifiers := range allModifiers {
				// Remove false keys
				for mod, val := range modifiers {
					v, ok := val.(bool)
					if ok && !v {
						delete(modifiers, mod)
					}
				}
				modJSON, _ := json.Marshal(modifiers)
				element.CreateAttr("modifiers", string(modJSON))
			}
		})

	commonMixin.AddMethod("ProcessFieldElementModifiers",
		`ProcessFieldElementModifiers modifies the given modifiers map by taking into account:
		- 'invisible', 'readonly' and 'required' attributes in field tags
		- 'ReadOnly' and 'Required' parameters of the model's field'
		It returns the modified map.`,
		func(rc models.RecordCollection, element *etree.Element, fieldInfos map[string]*models.FieldInfo, modifiers map[string]interface{}) map[string]interface{} {
			fieldName := element.SelectAttr("name").Value
			// Check if we have the modifier as attribute in the field node
			for modifier := range modifiers {
				modTag := element.SelectAttrValue(modifier, "")
				if modTag != "" && modTag != "0" && modTag != "false" {
					modifiers[modifier] = true
				}
			}
			// Force modifiers if defined in the model
			if fieldInfos[fieldName].ReadOnly {
				modifiers["readonly"] = true
			}
			if fieldInfos[fieldName].Required {
				modifiers["required"] = true
			}
			return modifiers
		})

	commonMixin.AddMethod("ProcessElementAttrs",
		`ProcessElementAttrs returns a modifiers map according to the domain
		in attrs of the given element`,
		func(rc models.RecordCollection, element *etree.Element) map[string]interface{} {
			modifiers := map[string]interface{}{"readonly": false, "required": false, "invisible": false}
			attrStr := element.SelectAttrValue("attrs", "")
			if attrStr == "" {
				return modifiers
			}
			var attrs map[string]models.Domain
			err := json.Unmarshal([]byte(attrStr), &attrs)
			if err != nil {
				logging.LogAndPanic(log, "Invalid attrs definition", "model", rc.ModelName(), "attrs", attrStr)
			}
			for modifier := range modifiers {
				cond := models.ParseDomain(attrs[modifier])
				if cond == nil {
					continue
				}
				modifiers[modifier] = attrs[modifier]
			}
			return modifiers
		})

	commonMixin.AddMethod("UpdateFieldNames",
		`UpdateFieldNames changes the field names in the view to the column names.
		If a field name is already column names then it does nothing.
		This method also modifies the fields in the given fieldInfo to match the new name.`,
		func(rc models.RecordCollection, doc *etree.Document, fieldInfos *map[string]*models.FieldInfo) {
			for _, fieldTag := range doc.FindElements("//field") {
				fieldName := fieldTag.SelectAttr("name").Value
				fieldJSON := rc.Model().JSONizeFieldName(fieldName)
				fieldTag.RemoveAttr("name")
				fieldTag.CreateAttr("name", fieldJSON)
			}
			for _, labelTag := range doc.FindElements("//label") {
				fieldName := labelTag.SelectAttr("for").Value
				fieldJSON := rc.Model().JSONizeFieldName(fieldName)
				labelTag.RemoveAttr("for")
				labelTag.CreateAttr("for", fieldJSON)
			}
		})

	commonMixin.AddMethod("SearchRead",
		`SearchRead retrieves database records according to the filters defined in params.`,
		func(rc models.RecordCollection, params webdata.SearchParams) []models.FieldMap {
			if searchCond := models.ParseDomain(params.Domain); searchCond != nil {
				rc = rc.Search(searchCond)
			}
			// Limit
			rc = rc.Limit(models.ConvertLimitToInt(params.Limit))

			// Offset
			if params.Offset != 0 {
				rc = rc.Offset(params.Offset)
			}

			// Order
			if params.Order != "" {
				rc = rc.OrderBy(strings.Split(params.Order, ",")...)
			}

			rSet := rc.Fetch()
			return rSet.Call("Read", params.Fields).([]models.FieldMap)
		})

	baseMixin := pool.BaseMixin()

	baseMixin.AddMethod("ToggleActive",
		`ToggleActive toggles the Active field of this object`,
		func(rs pool.BaseMixinSet) {
			if rs.Active() {
				rs.SetActive(false)
			} else {
				rs.SetActive(true)
			}
		})
}
