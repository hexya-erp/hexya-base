// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package methods

import (
	"fmt"

	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/models/security"
	"github.com/npiganeau/yep/yep/models/types"
)

// BaseAuthBackend is the authentication backend of the Base module
// Users are authenticated against the ResUsers model in the database
type BaseAuthBackend struct{}

// Authenticate the user defined by login and secret.
func (bab *BaseAuthBackend) Authenticate(login, secret string, context *types.Context) (uid int64, err error) {
	models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
		uid, err = pool.ResUsers().NewSet(env).WithNewContext(context).Authenticate(login, secret)
	})
	return
}

func initUsers() {
	pool.ResUsers().ExtendMethod("NameGet", "",
		func(rs pool.ResUsersSet) string {
			res := rs.Super().NameGet()
			return fmt.Sprintf("%s (%s)", res, rs.Login())
		})

	pool.ResUsers().AddMethod("ContextGet",
		`UsersContextGet returns a context with the user's lang, tz and uid
		This method must be called on a singleton.`,
		func(rs pool.ResUsersSet) *types.Context {
			rs.EnsureOne()
			res := types.NewContext()
			res = res.WithKey("lang", rs.Lang())
			res = res.WithKey("tz", rs.TZ())
			res = res.WithKey("uid", rs.ID())
			return res
		})

	pool.ResUsers().AddMethod("Authenticate",
		"Authenticate the user defined by login and secret",
		func(rs pool.ResUsersSet, login, secret string) (uid int64, err error) {
			user := rs.Search(pool.ResUsers().Login().Equals(login))
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

	security.AuthenticationRegistry.RegisterBackend(new(BaseAuthBackend))
}
