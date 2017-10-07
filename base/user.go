// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/tools/emailutils"
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
	cpWizard.AddFields(map[string]models.FieldDefinition{
		"Users": models.One2ManyField{RelationModel: pool.UserChangePasswordWizardLine(),
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
			}},
	})

	cpWizard.Methods().ChangePasswordButton().DeclareMethod(
		`ChangePasswordButton is called when the user clicks on 'Apply' button in the popup.
		It updates the user's password.`,
		func(rs pool.UserChangePasswordWizardSet) {
			for _, userLine := range rs.Users().Records() {
				userLine.User().SetPassword(userLine.NewPassword())
			}
		})

	cpWizardLine := pool.UserChangePasswordWizardLine().DeclareTransientModel()
	cpWizardLine.AddFields(map[string]models.FieldDefinition{
		"Wizard":      models.Many2OneField{RelationModel: pool.UserChangePasswordWizard()},
		"User":        models.Many2OneField{RelationModel: pool.User(), OnDelete: models.Cascade},
		"UserLogin":   models.CharField{},
		"NewPassword": models.CharField{},
	})

	userLogModel := pool.UserLog().DeclareModel()
	userLogModel.SetDefaultOrder("id desc")

	userModel := pool.User().DeclareModel()
	userModel.SetDefaultOrder("Login")
	userModel.AddFields(map[string]models.FieldDefinition{
		"Partner": models.Many2OneField{RelationModel: pool.Partner(), Required: true, Embed: true,
			OnDelete: models.Restrict, String: "Related Partner", Help: "Partner-related data of the user"},
		"Login": models.CharField{Required: true, Unique: true, Help: "Used to log into the system",
			OnChange: pool.User().Methods().OnchangeLogin()},
		"Password": models.CharField{Default: models.DefaultValue(""), NoCopy: true,
			Help: "Keep empty if you don't want the user to be able to connect on the system."},
		"NewPassword": models.CharField{String: "Set Password", Compute: pool.User().Methods().ComputePassword(),
			Inverse: pool.User().Methods().InversePassword(), Depends: []string{""},
			Help: `Specify a value only when creating a user or if you're
changing the user's password, otherwise leave empty. After
a change of password, the user has to login again.`},
		"Signature": models.TextField{}, // TODO Switch to HTML field when implemented in client
		"Active":    models.BooleanField{Default: models.DefaultValue(true)},
		"ActionID": models.CharField{GoType: new(actions.ActionRef), String: "Home Action",
			Help: "If specified, this action will be opened at log on for this user, in addition to the standard menu."},
		"Groups": models.Many2ManyField{RelationModel: pool.Group(), JSON: "group_ids"},
		"Logs": models.One2ManyField{RelationModel: pool.UserLog(), ReverseFK: "CreateUID", String: "User log entries",
			JSON: "log_ids"},
		"LoginDate": models.DateTimeField{Related: "Logs.CreateDate", String: "Latest Connection"},
		"Share": models.BooleanField{Compute: pool.User().Methods().ComputeShare(), Depends: []string{"Groups"},
			String: "Share User", Stored: true, Help: "External user with limited access, created only for the purpose of sharing data."},
		"CompaniesCount": models.IntegerField{String: "Number of Companies",
			Compute: pool.User().Methods().ComputeCompaniesCount(), GoType: new(int)},
		"Company": models.Many2OneField{RelationModel: pool.Company(), Required: true, Default: func(env models.Environment, vals models.FieldMap) interface{} {
			return pool.Company().NewSet(env).CompanyDefaultGet()
		}, Help: "The company this user is currently working for.", Constraint: pool.User().Methods().CheckCompany()},
		"Companies": models.Many2ManyField{RelationModel: pool.Company(), JSON: "company_ids",
			Default: func(env models.Environment, vals models.FieldMap) interface{} {
				return pool.Company().NewSet(env).CompanyDefaultGet()
			}, Constraint: pool.User().Methods().CheckCompany()},
	})

	userModel.Methods().SelfReadableFields().DeclareMethod(
		`SelfReadableFields returns the list of its own fields that a user can read.`,
		func(rs pool.UserSet) map[string]bool {
			return map[string]bool{
				"Signature": true, "Company": true, "Login": true, "Email": true, "Name": true, "Image": true,
				"ImageMedium": true, "ImageSmall": true, "Lang": true, "TZ": true, "TZOffset": true, "Groups": true,
				"Partner": true, "LastUpdate": true, "ActionID": true,
			}
		})

	userModel.Methods().SelfWritableFields().DeclareMethod(
		`SelfWritableFields returns the list of its own fields that a user can write.`,
		func(rs pool.UserSet) map[string]bool {
			return map[string]bool{
				"Signature": true, "ActionID": true, "Company": true, "Email": true, "Name": true,
				"Image": true, "ImageMedium": true, "ImageSmall": true, "Lang": true, "TZ": true,
			}
		})

	userModel.Methods().ComputePassword().DeclareMethod(
		`ComputePassword is a technical function for the new password mechanism. It always returns an empty string`,
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			return &pool.UserData{NewPassword: ""}, []models.FieldNamer{pool.User().NewPassword()}
		})

	userModel.Methods().InversePassword().DeclareMethod(
		`InversePassword is used in the new password mechanism.`,
		func(rs pool.UserSet, vals models.FieldMapper) {
			if rs.NewPassword() == "" {
				return
			}
			if rs.ID() == rs.Env().Uid() {
				log.Panic(rs.T("Please use the change password wizard (in User Preferences or User menu) to change your own password."))
			}
			rs.SetPassword(rs.NewPassword())
		})

	userModel.Methods().ComputeShare().DeclareMethod(
		`ComputeShare checks if this is a shared user`,
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			return &pool.UserData{
				Share: !rs.HasGroup(GroupUser.ID),
			}, []models.FieldNamer{pool.User().Share()}
		})

	userModel.Methods().ComputeCompaniesCount().DeclareMethod(
		`ComputeCompaniesCount retrieves the number of companies in the system`,
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			return &pool.UserData{
				CompaniesCount: pool.Company().NewSet(rs.Env()).Sudo().SearchCount(),
			}, []models.FieldNamer{pool.User().CompaniesCount()}
		})

	userModel.Methods().OnchangeLogin().DeclareMethod(
		`OnchangeLogin matches the email if the login is an email`,
		func(rs pool.UserSet) (*pool.UserData, []models.FieldNamer) {
			if rs.Login() == "" || !emailutils.IsValidAddress(rs.Login()) {
				return &pool.UserData{}, []models.FieldNamer{}
			}
			return &pool.UserData{Email: rs.Login()}, []models.FieldNamer{pool.User().Email()}
		})

	userModel.Methods().CheckCompany().DeclareMethod(
		`CheckCompany checks that the user's company is one of its authorized companies`,
		func(rs pool.UserSet) {
			for _, company := range rs.Companies().Records() {
				if rs.Company().Equals(company) {
					return
				}
			}
			log.Panic(rs.T("The chosen company is not in the allowed companies for this user"))
		})

	userModel.Methods().Read().Extend("",
		func(rs pool.UserSet, fields []string) []models.FieldMap {
			rSet := rs
			if len(fields) > 0 && rs.ID() == rs.Env().Uid() {
				var hasUnsafeFields bool
				for _, key := range fields {
					if !rs.SelfReadableFields()[key] {
						hasUnsafeFields = true
						break
					}
				}
				if !hasUnsafeFields {
					rSet = rs.Sudo()
				}
			}
			result := rSet.Super().Read(fields)
			if !rs.CheckExecutionPermission(pool.User().Methods().Write().Underlying(), true) {
				for i, res := range result {
					if res["id"] != rs.Env().Uid() {
						if _, exists := res["password"]; exists {
							result[i]["password"] = "********"
						}
					}
				}
			}
			return result

		})

	userModel.Methods().Search().Extend("",
		func(rs pool.UserSet, cond pool.UserCondition) pool.UserSet {
			for _, field := range cond.Fields() {
				if pool.User().JSONizeFieldName(field) == "password" {
					log.Panic(rs.T("Invalid search criterion: password"))
				}
			}
			return rs.Super().Search(cond)
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
			if user.Password() == "" || user.Password() != secret {
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
