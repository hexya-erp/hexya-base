// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import (
	"encoding/json"
	"strings"

	"github.com/hexya-erp/hexya-base/web/domains"
	"github.com/hexya-erp/hexya-base/web/webdata"
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/operator"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/tools/etree"
	"github.com/hexya-erp/hexya/hexya/views"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	commonMixin := pool.CommonMixin()

	commonMixin.Methods().Create().Extend("",
		func(rs pool.CommonMixinSet, data models.FieldMapper) pool.CommonMixinSet {
			fMap := rs.ProcessDataValues(data)
			res := rs.Super().Create(fMap)
			return res
		})

	commonMixin.Methods().Write().Extend("",
		func(rs pool.CommonMixinSet, data models.FieldMapper, fieldsToUnset ...models.FieldNamer) bool {
			fMap := rs.ProcessDataValues(data)
			res := rs.Super().Write(fMap, fieldsToUnset...)
			return res
		})

	commonMixin.Methods().Read().Extend("",
		func(rs pool.CommonMixinSet, fields []string) []models.FieldMap {
			res := rs.Super().Read(fields)
			for i, fMap := range res {
				rec := pool.CommonMixin().NewSet(rs.Env(), rs.ModelName()).Search(pool.CommonMixin().ID().Equals(fMap["id"].(int64)))
				fInfos := rec.FieldsGet(models.FieldsGetArgs{})
				res[i] = rec.AddNamesToRelations(fMap, fInfos)
			}
			return res
		})

	commonMixin.Methods().AddNamesToRelations().DeclareMethod(
		`AddNameToRelations returns the given FieldMap after getting the name of all 2one relation ids`,
		func(rs pool.CommonMixinSet, fMap models.FieldMap, fInfos map[string]*models.FieldInfo) models.FieldMap {
			for fName, value := range fMap {
				fi := fInfos[fName]
				switch v := value.(type) {
				case models.RecordCollection:
					switch {
					case fi.Type.Is2OneRelationType():
						if rcId := v.Get("id"); rcId != int64(0) {
							value = [2]interface{}{rcId, v.Call("NameGet").(string)}
						} else {
							value = false
						}
					case fi.Type.Is2ManyRelationType():
						value = v.Ids()
					}
				case int64:
					if fi.Type.Is2OneRelationType() {
						rSet := rs.Env().Pool(fi.Relation).Search(rs.Model().Field("id").Equals(v))
						value = [2]interface{}{v, rSet.Call("NameGet").(string)}
					}
				}
				fMap[fName] = value
			}
			return fMap
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().NameSearch().DeclareMethod(
		`NameSearch searches for records that have a display name matching the given
		"name" pattern when compared with the given "operator", while also
		matching the optional search domain ("args").

		This is used for example to provide suggestions based on a partial
		value for a relational field. Sometimes be seen as the inverse
		function of NameGet but it is not guaranteed to be.`,
		func(rc models.RecordCollection, params webdata.NameSearchParams) []webdata.RecordIDWithName {
			if params.Operator == "" {
				params.Operator = operator.IContains
			}
			searchRs := rc.Model().Search(rc.Env(), rc.Model().Field("Name").AddOperator(params.Operator, params.Name)).Limit(models.ConvertLimitToInt(params.Limit))
			if extraCondition := domains.ParseDomain(params.Args); extraCondition != nil {
				searchRs = searchRs.Search(extraCondition)
			}

			searchRs.Load("ID", "DisplayName")

			res := make([]webdata.RecordIDWithName, searchRs.Len())
			for i, rec := range searchRs.Records() {
				res[i].ID = rec.Get("id").(int64)
				res[i].Name = rec.Get("display_name").(string)
			}
			return res
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().ProcessDataValues().DeclareMethod(
		`ProcessDataValues updates the given data values for Write and Create methods to be
		compatible with the ORM, in particular for relation fields`,
		func(rs pool.CommonMixinSet, data models.FieldMapper) models.FieldMap {
			fMap := data.FieldMap()
			fInfos := rs.FieldsGet(models.FieldsGetArgs{})
			for f, v := range fMap {
				fJSON := rs.Model().JSONizeFieldName(f)
				if _, exists := fInfos[fJSON]; !exists {
					log.Panic("Unable to find field", "model", rs.ModelName(), "field", f)
				}
				switch fInfos[fJSON].Type {
				case fieldtype.Many2Many:
					fMap[f] = rs.NormalizeM2MData(f, fInfos[fJSON], v)
				case fieldtype.One2Many:
					fMap[f] = rs.ExecuteO2MActions(f, fInfos[fJSON], v)
				}
			}
			return fMap
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().ExecuteO2MActions().DeclareMethod(
		`ExecuteO2MActions executes the actions on one2many fields given by
		the list of triplets received from the client`,
		func(rs pool.CommonMixinSet, fieldName string, info *models.FieldInfo, value interface{}) interface{} {
			switch v := value.(type) {
			case []interface{}:
				relSet := rs.Env().Pool(info.Relation)
				recs := rs.Get(fieldName).(models.RecordCollection)
				if len(v) == 0 {
					return nil
				}
				// We assume we have a list of triplets from client
				for _, triplet := range v {
					action := int(triplet.([]interface{})[0].(float64))
					var values models.FieldMap
					switch val := triplet.([]interface{})[2].(type) {
					case bool:
					case map[string]interface{}:
						values = models.FieldMap(val)
					case models.FieldMap:
						values = val
					}
					switch action {
					case 0:
						// Create a new record with values
						newRec := relSet.Call("Create", models.FieldMap(values)).(models.RecordSet).Collection()
						recs = recs.Union(newRec)
					case 1:
						// Update the id record with the given values
						id := int(triplet.([]interface{})[1].(float64))
						rec := relSet.Search(relSet.Model().Field("ID").Equals(id))
						rec.Call("Write", values)
						// add rec to recs in case we are in create
						recs = recs.Union(rec)
					case 2:
						// Remove and delete the id record
						id := int(triplet.([]interface{})[1].(float64))
						rec := relSet.Search(relSet.Model().Field("ID").Equals(id))
						recs = recs.Subtract(rec)
						rec.Call("Unlink")
					case 3:
						// Detach the id record
						id := int(triplet.([]interface{})[1].(float64))
						rec := relSet.Search(relSet.Model().Field("ID").Equals(id))
						recs = recs.Subtract(rec)
					}
				}
				return recs.Ids()
			}
			return value
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().NormalizeM2MData().DeclareMethod(
		`NormalizeM2MData converts the list of triplets received from the client into the final list of ids
		to keep in the Many2Many relationship of this model through the given field.`,
		func(rs pool.CommonMixinSet, fieldName string, info *models.FieldInfo, value interface{}) interface{} {
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
						return ids
					}
				}
			}
			return value
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().GetFormviewId().DeclareMethod(
		`GetFormviewID returns an view id to open the document with.
		This method is meant to be overridden in addons that want
 		to give specific view ids for example.`,
		func(rs pool.CommonMixinSet) string {
			return ""
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().GetFormviewAction().DeclareMethod(
		`GetFormviewAction returns an action to open the document.
		This method is meant to be overridden in addons that want
		to give specific view ids for example.`,
		func(rs pool.CommonMixinSet) *actions.BaseAction {
			viewID := rs.GetFormviewId()
			return &actions.BaseAction{
				Type:        actions.ActionActWindow,
				Model:       rs.ModelName(),
				ActViewType: actions.ActionViewTypeForm,
				ViewMode:    "form",
				Views:       []views.ViewTuple{{ID: viewID, Type: views.VIEW_TYPE_FORM}},
				Target:      "current",
				ResID:       rs.ID(),
				Context:     rs.Env().Context(),
			}
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().FieldsViewGet().DeclareMethod(
		`FieldsViewGet is the base implementation of the 'FieldsViewGet' method which
		gets the detailed composition of the requested view like fields, mixin,
		view architecture.`,
		func(rs pool.CommonMixinSet, args webdata.FieldsViewGetParams) *webdata.FieldsViewData {
			view := views.Registry.GetByID(args.ViewID)
			if view == nil {
				view = views.Registry.GetFirstViewForModel(rs.ModelName(), views.ViewType(args.ViewType))
			}
			cols := make([]models.FieldName, len(view.Fields))
			for i, f := range view.Fields {
				cols[i] = models.FieldName(rs.Model().JSONizeFieldName(string(f)))
			}
			fInfos := rs.FieldsGet(models.FieldsGetArgs{Fields: cols})
			arch := rs.ProcessView(view.Arch, fInfos)
			toolbar := rs.GetToolbar()
			res := webdata.FieldsViewData{
				Name:    view.Name,
				Arch:    arch,
				ViewID:  args.ViewID,
				Model:   view.Model,
				Type:    view.Type,
				Toolbar: toolbar,
				Fields:  fInfos,
			}
			for field, sViews := range view.SubViews {
				fJSON := rs.Model().JSONizeFieldName(field)
				relRS := rs.Env().Pool(fInfos[fJSON].Relation)
				for svType, sv := range sViews {
					svFields := relRS.Call("FieldsGet", models.FieldsGetArgs{Fields: sv.Fields}).(map[string]*models.FieldInfo)
					if res.Fields[fJSON].Views == nil {
						res.Fields[fJSON].Views = make(map[string]interface{})
					}
					res.Fields[fJSON].Views[string(svType)] = &webdata.SubViewData{
						Fields: svFields,
						Arch:   relRS.Call("ProcessView", sv.Arch, svFields).(string),
					}
				}
			}
			return &res
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().GetToolbar().DeclareMethod(
		`GetToolbar returns a toolbar populated with the actions linked to this model`,
		func(rs pool.CommonMixinSet) webdata.Toolbar {
			var res webdata.Toolbar
			for _, a := range actions.Registry.GetActionLinksForModel(rs.ModelName()) {
				switch a.Type {
				case actions.ActionActWindow, actions.ActionServer:
					res.Action = append(res.Action, a)
				}
			}
			return res
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().ProcessView().DeclareMethod(
		`ProcessView makes all the necessary modifications to the view
		arch and returns the new xml string.`,
		func(rs pool.CommonMixinSet, arch string, fieldInfos map[string]*models.FieldInfo) string {
			// Load arch as etree
			doc := etree.NewDocument()
			if err := doc.ReadFromString(arch); err != nil {
				log.Panic("Unable to parse view arch", "arch", arch, "error", err)
			}
			// Apply changes
			rs.UpdateFieldNames(doc, &fieldInfos)
			rs.SanitizeSearchView(doc)
			rs.AddModifiers(doc, fieldInfos)
			// Dump xml to string and return
			res, err := doc.WriteToString()
			if err != nil {
				log.Panic("Unable to render XML", "error", err)
			}
			return res
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().SanitizeSearchView().DeclareMethod(
		`SanitizeSearchView adds the missing attribute if it does not exist`,
		func(rs pool.CommonMixinSet, doc *etree.Document) {
			if doc.Root().Tag != "search" {
				return
			}
			for _, fieldTag := range doc.FindElements("//field") {
				if fieldTag.SelectAttrValue("domain", "") == "" {
					fieldTag.CreateAttr("domain", "[]")
				}
			}
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().AddModifiers().DeclareMethod(
		`AddModifiers adds the modifiers attribute nodes to given xml doc.`,
		func(rs pool.CommonMixinSet, doc *etree.Document, fieldInfos map[string]*models.FieldInfo) {
			allModifiers := make(map[*etree.Element]map[string]interface{})
			// Process attrs on all nodes
			for _, attrsTag := range doc.FindElements("[@attrs]") {
				allModifiers[attrsTag] = rs.ProcessElementAttrs(attrsTag)
			}
			// Process field nodes
			for _, fieldTag := range doc.FindElements("//field") {
				mods, exists := allModifiers[fieldTag]
				if !exists {
					mods = map[string]interface{}{"readonly": false, "required": false, "invisible": false}
				}
				allModifiers[fieldTag] = rs.ProcessFieldElementModifiers(fieldTag, fieldInfos, mods)
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
				// Remove required if field is invisible or readonly
				if req, ok := modifiers["required"].(bool); ok && req {
					inv, ok2 := modifiers["invisible"].(bool)
					ro, ok3 := modifiers["readonly"].(bool)
					if ok2 && inv || ok3 && ro {
						delete(modifiers, "required")
					}
				}

				modJSON, _ := json.Marshal(modifiers)
				element.CreateAttr("modifiers", string(modJSON))
			}
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().ProcessFieldElementModifiers().DeclareMethod(
		`ProcessFieldElementModifiers modifies the given modifiers map by taking into account:
		- 'invisible', 'readonly' and 'required' attributes in field tags
		- 'ReadOnly' and 'Required' parameters of the model's field'
		It returns the modified map.`,
		func(rs pool.CommonMixinSet, element *etree.Element, fieldInfos map[string]*models.FieldInfo, modifiers map[string]interface{}) map[string]interface{} {
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
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().ProcessElementAttrs().DeclareMethod(
		`ProcessElementAttrs returns a modifiers map according to the domain
		in attrs of the given element`,
		func(rc models.RecordCollection, element *etree.Element) map[string]interface{} {
			modifiers := map[string]interface{}{"readonly": false, "required": false, "invisible": false}
			attrStr := element.SelectAttrValue("attrs", "")
			if attrStr == "" {
				return modifiers
			}
			var attrs map[string]domains.Domain
			attrStr = strings.Replace(attrStr, "(", "[", -1)
			attrStr = strings.Replace(attrStr, ")", "]", -1)
			attrStr = strings.Replace(attrStr, "'", "\"", -1)
			attrStr = strings.Replace(attrStr, "True", "true", -1)
			attrStr = strings.Replace(attrStr, "False", "false", -1)
			err := json.Unmarshal([]byte(attrStr), &attrs)
			if err != nil {
				log.Panic("Invalid attrs definition", "model", rc.ModelName(), "attrs", attrStr)
			}
			for modifier := range modifiers {
				cond := domains.ParseDomain(attrs[modifier])
				if cond == nil {
					continue
				}
				modifiers[modifier] = attrs[modifier]
			}
			return modifiers
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().UpdateFieldNames().DeclareMethod(
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
				if labelTag.SelectAttr("for") == nil {
					continue
				}
				fieldName := labelTag.SelectAttr("for").Value
				fieldJSON := rc.Model().JSONizeFieldName(fieldName)
				labelTag.RemoveAttr("for")
				labelTag.CreateAttr("for", fieldJSON)
			}
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().SearchRead().DeclareMethod(
		`SearchRead retrieves database records according to the filters defined in params.`,
		func(rs pool.CommonMixinSet, params webdata.SearchParams) []models.FieldMap {
			rSet := rs.AddDomainLimitOffset(params.Domain, models.ConvertLimitToInt(params.Limit), params.Offset, params.Order).Fetch()

			records := rSet.Read(params.Fields)
			return records
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().AddDomainLimitOffset().DeclareMethod(
		`AddDomainLimitOffsetOrder adds the given domain, limit, offset
		and order to the current RecordSet query.`,
		func(rc models.RecordCollection, domain domains.Domain, limit int, offset int, order string) models.RecordCollection {
			if searchCond := domains.ParseDomain(domain); searchCond != nil {
				rc = rc.Search(searchCond)
			}
			// Limit
			rc = rc.Limit(limit)

			// Offset
			if offset != 0 {
				rc = rc.Offset(offset)
			}

			// Order
			if order != "" {
				rc = rc.OrderBy(strings.Split(order, ",")...)
			}
			return rc
		}).AllowGroup(security.GroupEveryone)

	commonMixin.Methods().ReadGroup().DeclareMethod(
		`Get a list of record aggregates according to the given parameters.`,
		func(rs pool.CommonMixinSet, params webdata.ReadGroupParams) []models.FieldMap {
			rSet := rs.AddDomainLimitOffset(params.Domain, models.ConvertLimitToInt(params.Limit), params.Offset, params.Order)
			rSet = rSet.GroupBy(models.ConvertToFieldNameSlice(params.GroupBy)...)
			aggregates := rSet.Aggregates(models.ConvertToFieldNameSlice(params.Fields)...)
			res := make([]models.FieldMap, len(aggregates))
			fInfos := rSet.FieldsGet(models.FieldsGetArgs{})
			for i, ag := range aggregates {
				line := rs.AddNamesToRelations(ag.Values, fInfos)
				line["__count"] = ag.Count
				line["__domain"] = ag.Condition.Serialize()
				res[i] = line
			}
			return res
		}).AllowGroup(security.GroupEveryone)

}
