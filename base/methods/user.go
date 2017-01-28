// Copyright 2016 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package methods

import (
	"fmt"
	"github.com/npiganeau/yep/pool"
	"github.com/npiganeau/yep/yep/models/types"
)

func initUsers() {
	pool.ResUsers().ExtendMethod("NameGet", "",
		func(rs pool.ResUsersSet) string {
			res := rs.Super()
			return fmt.Sprintf("%s (%s)", res, rs.Login())
		})

	pool.ResUsers().CreateMethod("ContextGet",
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
}
