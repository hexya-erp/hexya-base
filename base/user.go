// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/actions"
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/operator"
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
					ul := pool.UserChangePasswordWizardLine().Create(env, &pool.UserChangePasswordWizardLineData{
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
		"Wizard":      models.Many2OneField{RelationModel: pool.UserChangePasswordWizard(), Required: true},
		"User":        models.Many2OneField{RelationModel: pool.User(), OnDelete: models.Cascade, Required: true},
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
			if rs.Company().Intersect(rs.Companies()).IsEmpty() {
				log.Panic(rs.T("The chosen company is not in the allowed companies for this user"))
			}
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

	userModel.Methods().Create().Extend("",
		func(rs pool.UserSet, vals *pool.UserData) pool.UserSet {
			user := rs.Super().Create(vals)
			user.Partner().SetActive(user.Active())
			if !user.Partner().Company().IsEmpty() {
				user.Partner().SetCompany(user.Company())
			}
			return user
		})

	userModel.Methods().Write().Extend("",
		func(rs pool.UserSet, data *pool.UserData, fieldsToUnset ...models.FieldNamer) bool {
			if val, exists := data.Get(pool.User().Active(), fieldsToUnset...); exists && !val.(bool) {
				for _, user := range rs.Records() {
					if user.ID() == security.SuperUserID {
						log.Panic(rs.T("You cannot deactivate the admin user."))
					}
					if user.ID() == rs.Env().Uid() {
						log.Panic(rs.T("You cannot deactivate the user you're currently logged in as."))
					}
				}
			}
			rSet := rs
			if rs.ID() == rs.Env().Uid() {
				var hasUnsafeFields bool
				for key := range data.FieldsSet(fieldsToUnset...) {
					if !rs.SelfWritableFields()[string(key)] {
						hasUnsafeFields = true
						break
					}
				}
				if !hasUnsafeFields {
					if comp, exists := data.Get(pool.User().Company(), fieldsToUnset...); exists {
						if comp.(pool.CompanySet).Intersect(pool.User().NewSet(rs.Env()).CurrentUser().Companies()).IsEmpty() {
							data, fieldsToUnset = data.Remove(rs, pool.User().Company(), fieldsToUnset...)
						}
					}
					// safe fields only, so we write as super-user to bypass access rights
					rSet = rs.Sudo()
				}
			}
			res := rSet.Super().Write(data, fieldsToUnset...)
			if _, ok := data.Get(pool.User().Groups(), fieldsToUnset...); ok {
				// We get groups before removing all memberships otherwise we might get stuck with permissions if we
				// are modifying our own user memberships.
				rs.SyncMemberships()
			}
			if comp, exists := data.Get(pool.User().Company(), fieldsToUnset...); exists {
				for _, user := range rs.Records() {
					// if partner is global we keep it that way
					if !user.Partner().Company().Equals(comp.(pool.CompanySet)) {
						user.Partner().SetCompany(user.Company())
					}
				}
			}
			return res
		})

	userModel.Methods().Unlink().Extend("",
		func(rs pool.UserSet) int64 {
			for _, id := range rs.Ids() {
				if id == security.SuperUserID {
					log.Panic(rs.T("You can not remove the admin user as it is used internally for resources created by Hexya"))
				}
			}
			return rs.Super().Unlink()
		})

	userModel.Methods().SearchByName().Extend("",
		func(rs pool.UserSet, name string, op operator.Operator, additionalCond pool.UserCondition, limit int) pool.UserSet {
			if name == "" {
				return rs.Super().SearchByName(name, op, additionalCond, limit)
			}
			var users pool.UserSet
			if op == operator.Equals || op == operator.IContains {
				users = pool.User().Search(rs.Env(), pool.User().Login().Equals(name).AndCond(additionalCond)).Limit(limit)
			}
			if users.IsEmpty() {
				users = pool.User().Search(rs.Env(), pool.User().Name().AddOperator(op, name).AndCond(additionalCond)).Limit(limit)
			}
			return users
		})

	userModel.Methods().Copy().Extend("",
		func(rs pool.UserSet, overrides *pool.UserData, fieldsToReset ...models.FieldNamer) pool.UserSet {
			rs.EnsureOne()
			_, eName := overrides.Get(pool.User().Name(), fieldsToReset...)
			_, ePartner := overrides.Get(pool.User().Partner(), fieldsToReset...)
			if !eName && !ePartner {
				overrides.Name = rs.T("%s (copy)", rs.Name())
				fieldsToReset = append(fieldsToReset, pool.User().Name())
			}
			_, eLogin := overrides.Get(pool.User().Login(), fieldsToReset...)
			if !eLogin {
				overrides.Login = rs.T("%s (copy)", rs.Login())
				fieldsToReset = append(fieldsToReset, pool.User().Login())
			}
			return rs.Super().Copy(overrides, fieldsToReset...)
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

	userModel.Methods().ActionGet().DeclareMethod(
		`ActionGet returns the action for the preferences popup`,
		func(rs pool.UserSet) *actions.Action {
			return actions.Registry.GetById("base_action_res_users_my")
		})

	userModel.Methods().UpdateLastLogin().DeclareMethod(
		`UpdateLastLogin updates the last login date of the user`,
		func(rs pool.UserSet) {
			// only create new records to avoid any side-effect on concurrent transactions
			// extra records will be deleted by the periodical garbage collection
			pool.UserLog().Create(rs.Env(), &pool.UserLogData{})
		})

	userModel.Methods().CheckCredentials().DeclareMethod(
		`CheckCredentials checks that the user defined by its login and secret is allowed to log in.
		It returns the uid of the user on success and an error otherwise.`,
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

	userModel.Methods().Authenticate().DeclareMethod(
		"Authenticate the user defined by login and secret",
		func(rs pool.UserSet, login, secret string) (uid int64, err error) {
			uid, err = rs.CheckCredentials(login, secret)
			if err != nil {
				rs.UpdateLastLogin()
			}
			return
		})

	userModel.Methods().ChangePassword().DeclareMethod(
		`ChangePassword changes current user password. Old password must be provided explicitly
        to prevent hijacking an existing user session, or for cases where the cleartext
        password is not used to authenticate requests. It returns true or panics.`,
		func(rs pool.UserSet, oldPassword, newPassword string) bool {
			currentUser := pool.User().NewSet(rs.Env()).CurrentUser()
			uid, err := rs.CheckCredentials(currentUser.Login(), oldPassword)
			if err != nil || rs.Env().Uid() != uid {
				log.Panic("Invalid password", "user", currentUser.Login(), "uid", uid)
			}
			currentUser.SetPassword(newPassword)
			return true
		})

	userModel.Methods().PreferenceSave().DeclareMethod(
		`PreferenceSave is called when validating the preferences popup`,
		func(rs pool.UserSet) *actions.Action {
			return &actions.Action{
				Type: actions.ActionClient,
				Tag:  "reload_context",
			}
		})

	userModel.Methods().PreferenceChangePassword().DeclareMethod(
		`PreferenceChangePassword is called when clicking 'Change Password' in the preferences popup`,
		func(rs pool.UserSet) *actions.Action {
			return &actions.Action{
				Type:   actions.ActionClient,
				Tag:    "change_password",
				Target: "new",
			}
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

	userModel.Methods().IsAdmin().DeclareMethod(
		`IsAdmin returns true if this user is the administrator or member of the 'Access Rights' group`,
		func(rs pool.UserSet) bool {
			rs.EnsureOne()
			return rs.IsSuperUser() || rs.HasGroup(GroupERPManager.ID)
		})

	userModel.Methods().IsSuperUser().DeclareMethod(
		`IsSuperUser returns true if this user is the administrator`,
		func(rs pool.UserSet) bool {
			rs.EnsureOne()
			return rs.ID() == security.SuperUserID
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

	userModel.Methods().GetCompany().DeclareMethod(
		`GetCompany returns the current user's company.`,
		func(rs pool.UserSet) pool.CompanySet {
			return pool.User().NewSet(rs.Env()).CurrentUser().Company()
		})

	userModel.Methods().GetCompanyCurrency().DeclareMethod(
		`GetCompanyCurrency returns the currency of the current user's company.`,
		func(rs pool.UserSet) pool.CurrencySet {
			return pool.User().NewSet(rs.Env()).CurrentUser().Company().Currency()
		})

	userModel.Methods().CurrentUser().DeclareMethod(
		`CurrentUser returns a UserSet with the currently logged in user.`,
		func(rs pool.UserSet) pool.UserSet {
			return pool.User().Browse(rs.Env(), []int64{rs.Env().Uid()})
		})

	security.AuthenticationRegistry.RegisterBackend(new(BaseAuthBackend))

}
