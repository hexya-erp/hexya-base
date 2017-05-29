// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	_ "github.com/hexya-erp/hexya-base/web"
	"github.com/hexya-erp/hexya/hexya/tests"
)

func TestMain(m *testing.M) {
	tests.RunTests(m, "web")
}
