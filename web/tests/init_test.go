// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	_ "github.com/npiganeau/yep-base/web"
	"github.com/npiganeau/yep/yep/tests"
)

func TestMain(m *testing.M) {
	tests.RunTests(m, "web")
}
