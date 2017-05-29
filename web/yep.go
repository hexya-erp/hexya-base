// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	_ "github.com/hexya-erp/hexya-base/base"
	_ "github.com/hexya-erp/hexya-base/web/controllers"
	_ "github.com/hexya-erp/hexya-base/web/defs"
	"github.com/hexya-erp/hexya/hexya/server"
)

const (
	MODULE_NAME = "web"
)

func init() {
	server.RegisterModule(&server.Module{
		Name:     MODULE_NAME,
		PostInit: func() {},
	})
}
