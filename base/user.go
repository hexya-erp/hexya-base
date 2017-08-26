// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

// BaseAuthBackend is the authentication backend of the Base module
// Users are authenticated against the User model in the database
type BaseAuthBackend struct{}

// Authenticate the user defined by login and secret.
func (bab *BaseAuthBackend) Authenticate(login, secret string, context *types.Context) (uid int64, err error) {
	models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		uid, err = pool.User().NewSet(env).WithNewContext(context).Authenticate(login, secret)
	})
	return
}

func init() {
	cpWizard := pool.UserChangePasswordWizard().DeclareTransientModel()
	cpWizard.AddOne2ManyField("Users", models.ReverseFieldParams{RelationModel: pool.UserChangePasswordWizardLine(),
		ReverseFK: "Wizard", Default: func(env models.Environment, fMap models.FieldMap) interface{} {
			activeIds := env.Context().GetIntegerSlice("active_ids")
			userLines := pool.UserChangePasswordWizardLine().NewSet(env)
			for _, user := range pool.User().Search(env, pool.User().ID().In(activeIds)).Records() {
				ul := pool.UserChangePasswordWizardLine().Create(env, pool.UserChangePasswordWizardLineData{
					User:        user,
					UserLogin:   user.Login(),
					NewPassword: user.Password(),
				})
				userLines = userLines.Union(ul)
			}
			return userLines
		}})

	cpWizard.Methods().ChangePasswordButton().DeclareMethod(
		`ChangePasswordButton is called when the user clicks on 'Apply' button in the popup.
		It updates the user's password.`,
		func(rs pool.UserChangePasswordWizardSet) {
			for _, userLine := range rs.Users().Records() {
				userLine.User().SetPassword(userLine.NewPassword())
			}
		})

	cpWizardLine := pool.UserChangePasswordWizardLine().DeclareTransientModel()
	cpWizardLine.AddMany2OneField("Wizard", models.ForeignKeyFieldParams{RelationModel: pool.UserChangePasswordWizard()})
	cpWizardLine.AddMany2OneField("User", models.ForeignKeyFieldParams{RelationModel: pool.User(), OnDelete: models.Cascade})
	cpWizardLine.AddCharField("UserLogin", models.StringFieldParams{})
	cpWizardLine.AddCharField("NewPassword", models.StringFieldParams{})

	userModel := pool.User().DeclareModel()
	userModel.AddDateTimeField("LoginDate", models.SimpleFieldParams{})
	userModel.AddMany2OneField("Partner", models.ForeignKeyFieldParams{RelationModel: pool.Partner(), Embed: true})
	userModel.AddCharField("Login", models.StringFieldParams{Required: true, Unique: true})
	userModel.AddCharField("Password", models.StringFieldParams{})
	userModel.AddCharField("NewPassword", models.StringFieldParams{})
	userModel.AddTextField("Signature", models.StringFieldParams{})
	userModel.AddBooleanField("Active", models.SimpleFieldParams{Default: models.DefaultValue(true)})
	userModel.AddCharField("ActionID", models.StringFieldParams{GoType: new(actions.ActionRef)})
	userModel.AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: pool.Company()})
	userModel.AddMany2ManyField("Companies", models.Many2ManyFieldParams{RelationModel: pool.Company(), JSON: "company_ids"})
	userModel.AddMany2ManyField("Groups", models.Many2ManyFieldParams{RelationModel: pool.Group(), JSON: "group_ids"})
	userModel.AddBooleanField("Share", models.SimpleFieldParams{Compute: pool.User().Methods().ComputeShare(),
		String: "Share User", Stored: true, Help: "External user with limited access, created only for the purpose of sharing data."})

	userModel.Methods().ComputeShare().DeclareMethod(
		`ComputeShare checks if this is a shared user`,
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			return &pool.UserData{
				Share: !rs.HasGroup(GroupUser.ID),
			}, []models.FieldNamer{pool.User().Share()}
		})

	userModel.Methods().Write().Extend("",
		func(rs pool.UserSet, data models.FieldMapper, fieldsToUnset ...models.FieldNamer) bool {
			res := rs.Super().Write(data, fieldsToUnset...)
			fMap := data.FieldMap(fieldsToUnset...)
			_, ok1 := fMap["Groups"]
			_, ok2 := fMap["group_ids"]
			if ok1 || ok2 {
				// We get groups before removing all memberships otherwise we might get stuck with permissions if we
				// are modifying our own user memberships.
				rs.SyncMemberships()
			}
			return res
		})

	userModel.Methods().AddMandatoryGroups().DeclareMethod(
		`AddMandatoryGroups adds the group Everyone to everybody and the admin group to the admin`,
		func(rs pool.UserSet) {
			for _, user := range rs.Records() {
				dbGroupEveryone := pool.Group().Search(rs.Env(), pool.Group().GroupID().Equals(security.GroupEveryoneID))
				dbGroupAdmin := pool.Group().Search(rs.Env(), pool.Group().GroupID().Equals(security.GroupAdminID))
				groups := user.Groups()
				// Add groupAdmin for admin
				if user.ID() == security.SuperUserID {
					groups = groups.Union(dbGroupAdmin)
				}
				// Add groupEveryone if not already the case
				groups = groups.Union(dbGroupEveryone)

				user.SetGroups(groups)
			}
		})

	userModel.Methods().SyncMemberships().DeclareMethod(
		`SyncMemberships synchronises the users memberships with the Hexya internal registry`,
		func(rs pool.UserSet) {
			for _, user := range rs.Records() {
				if user.CheckGroupsSync() {
					continue
				}
				log.Debug("Updating user groups", "user", rs.Name(), "uid", rs.ID(), "groups", rs.Groups())
				// Push memberships to registry
				security.Registry.RemoveAllMembershipsForUser(user.ID())
				for _, dbGroup := range user.Groups().Records() {
					security.Registry.AddMembership(user.ID(), security.Registry.GetGroup(dbGroup.GroupID()))
				}
			}
		})

	userModel.Methods().CheckGroupsSync().DeclareMethod(
		`CheckGroupSync returns true if the groups in the internal registry match exactly
		database groups of the given users. This method must be called on a singleton`,
		func(rs pool.UserSet) bool {
			rs.EnsureOne()
		dbLoop:
			for _, dbGroup := range rs.Groups().Records() {
				for grp := range security.Registry.UserGroups(rs.ID()) {
					if grp.ID == dbGroup.GroupID() {
						continue dbLoop
					}
				}
				return false
			}
		rLoop:
			for grp := range security.Registry.UserGroups(rs.ID()) {
				for _, dbGroup := range rs.Groups().Records() {
					if grp.ID == dbGroup.GroupID() {
						continue rLoop
					}
				}
				return false
			}
			return true
		})

	userModel.Methods().NameGet().Extend("",
		func(rs pool.UserSet) string {
			res := rs.Super().NameGet()
			return fmt.Sprintf("%s (%s)", res, rs.Login())
		})

	userModel.Methods().ContextGet().DeclareMethod(
		`UsersContextGet returns a context with the user's lang, tz and uid
		This method must be called on a singleton.`,
		func(rs pool.UserSet) *types.Context {
			rs.EnsureOne()
			res := types.NewContext()
			res = res.WithKey("lang", rs.Lang())
			res = res.WithKey("tz", rs.TZ())
			res = res.WithKey("uid", rs.ID())
			return res
		})

	userModel.Methods().HasGroup().DeclareMethod(
		`HasGroup returns true if this user belongs to the group with the given ID.
		If this method is called on an empty RecordSet, then it checks if the current
		user belongs to the given group.`,
		func(rs pool.UserSet, groupID string) bool {
			userID := rs.ID()
			if userID == 0 {
				userID = rs.Env().Uid()
			}
			group := security.Registry.GetGroup(groupID)
			return security.Registry.HasMembership(userID, group)
		})

	userModel.Methods().Authenticate().DeclareMethod(
		"Authenticate the user defined by login and secret",
		func(rs pool.UserSet, login, secret string) (uid int64, err error) {
			user := rs.Search(pool.User().Login().Equals(login))
			if user.Len() == 0 {
				err = security.UserNotFoundError(login)
				return
			}
			if user.Password() != secret {
				err = security.InvalidCredentialsError(login)
				return
			}
			uid = user.ID()
			return
		})

	userModel.Methods().GetCompany().DeclareMethod(
		`GetCompany returns the current user's company.`,
		func(rs pool.UserSet) pool.CompanySet {
			return pool.User().Browse(rs.Env(), []int64{rs.Env().Uid()}).Company()
		})

	security.AuthenticationRegistry.RegisterBackend(new(BaseAuthBackend))

}
