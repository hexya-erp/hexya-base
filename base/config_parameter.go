// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/pool"
	"github.com/spf13/viper"
)

var defaultParameters map[string](func(env models.Environment) (string, pool.GroupSet)) = map[string](func(env models.Environment) (string, pool.GroupSet)){
	"web.base.url": func(env models.Environment) (string, pool.GroupSet) {
		return fmt.Sprintf("http://localhost:%s", viper.GetString("Server.Port")), pool.Group().NewSet(env)
	},
}

func init() {
	pool.ConfigParameter().DeclareModel()
	pool.ConfigParameter().AddFields(map[string]models.FieldDefinition{
		"Key":    models.CharField{Index: true, Required: true, Unique: true},
		"Value":  models.TextField{Required: true},
		"Groups": models.Many2ManyField{RelationModel: pool.Group()},
	})

	pool.ConfigParameter().Methods().Init().DeclareMethod(
		`Init Initializes the parameters listed in defaultParameters.
        It overrides existing parameters if force is 'true'.`,
		func(rs pool.ConfigParameterSet, force ...bool) {
			var forceInit bool
			if len(force) > 0 && force[0] {
				forceInit = true
			}
			for key, fnct := range defaultParameters {
				params := rs.Model().NewSet(rs.Env()).Sudo().Search(pool.ConfigParameter().Key().Equals(key))
				if forceInit || params.IsEmpty() {
					value, groups := fnct(rs.Env())
					pool.ConfigParameter().NewSet(rs.Env()).SetParam(key, value).LimitToGroups(groups)
				}
			}
		})

	pool.ConfigParameter().Methods().GetParam().DeclareMethod(
		`GetParam retrieves the value for a given key. It returns defaultValue if the parameter is missing.`,
		func(rs pool.ConfigParameterSet, key string, defaultValue string) string {
			param := rs.Model().Search(rs.Env(), pool.ConfigParameter().Key().Equals(key)).Limit(1).Load("Value")
			if param.Value() == "" {
				return defaultValue
			}
			return param.Value()
		})

	pool.ConfigParameter().Methods().SetParam().DeclareMethod(
		`SetParam sets the value of a parameter. It returns the parameter`,
		func(rs pool.ConfigParameterSet, key, value string) pool.ConfigParameterSet {
			var res pool.ConfigParameterSet
			param := rs.Model().Search(rs.Env(), pool.ConfigParameter().Key().Equals(key))
			if param.IsEmpty() {
				if value != "" {
					res = rs.Create(pool.ConfigParameterData{
						Key:   key,
						Value: value,
					})
				}
				return res
			}
			if value == "" {
				param.Unlink()
				return rs.Model().NewSet(rs.Env())
			}
			param.SetValue(value)
			return param
		})

	pool.ConfigParameter().Methods().LimitToGroups().DeclareMethod(
		`LimitToGroups limits the access to this key to the given list of groups`,
		func(rs pool.ConfigParameterSet, groups pool.GroupSet) {
			if rs.IsEmpty() {
				return
			}
			rs.SetGroups(groups)
		})

}
