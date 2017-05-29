// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package tests

import (
	"testing"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/pool"
	. "github.com/smartystreets/goconvey/convey"
)

var viewDef1 string = `
<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field name="Name" required="1" readonly="1"/>
			<field name="TZ" invisible="1"/>
		</group>
	</form>
</view>
`

var viewFieldInfos1 map[string]*models.FieldInfo = map[string]*models.FieldInfo{
	"name": {},
	"tz":   {},
}

var viewDef2 string = `
<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field name="Name" attrs='{"readonly": [["Function", "ilike", "manager"]], "required": [["ID", "!=", false]]}'/>
			<field name="TZ" invisible="1" attrs='{"invisble": [["Login", "!=", "john"]]}'/>
		</group>
	</form>
</view>
`

var viewFieldInfos2 map[string]*models.FieldInfo = map[string]*models.FieldInfo{
	"name": {Required: true},
	"tz":   {ReadOnly: true},
}

func TestViewModifiers(t *testing.T) {
	Convey("Testing correct modifiers injection in views", t, func() {
		models.SimulateInNewEnvironment(security.SuperUserID, func(env models.Environment) {
			Convey("'invisible', 'required' and 'readonly' field attributes should be set in modifiers", func() {
				view := pool.User().NewSet(env).ProcessView(viewDef1, viewFieldInfos1)
				So(view, ShouldEqual, `
<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field required="1" readonly="1" name="name" modifiers="{&quot;readonly&quot;:true}"/>
			<field invisible="1" name="tz" modifiers="{&quot;invisible&quot;:true}"/>
		</group>
	</form>
</view>
`)
			})
			Convey("attrs should be set in modifiers", func() {
				view := pool.User().NewSet(env).ProcessView(viewDef2, viewFieldInfos1)
				So(view, ShouldEqual, `
<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field attrs="{&quot;readonly&quot;: [[&quot;Function&quot;, &quot;ilike&quot;, &quot;manager&quot;]], &quot;required&quot;: [[&quot;ID&quot;, &quot;!=&quot;, false]]}" name="name" modifiers="{&quot;readonly&quot;:[[&quot;Function&quot;,&quot;ilike&quot;,&quot;manager&quot;]],&quot;required&quot;:[[&quot;ID&quot;,&quot;!=&quot;,false]]}"/>
			<field invisible="1" attrs="{&quot;invisble&quot;: [[&quot;Login&quot;, &quot;!=&quot;, &quot;john&quot;]]}" name="tz" modifiers="{&quot;invisible&quot;:true}"/>
		</group>
	</form>
</view>
`)
			})
			Convey("'Readonly' and 'Required' field data should be taken into account", func() {
				view := pool.User().NewSet(env).ProcessView(viewDef2, viewFieldInfos2)
				So(view, ShouldEqual, `
<view id="my_id" name="My View" model="ResUSers">
	<form>
		<group>
			<field attrs="{&quot;readonly&quot;: [[&quot;Function&quot;, &quot;ilike&quot;, &quot;manager&quot;]], &quot;required&quot;: [[&quot;ID&quot;, &quot;!=&quot;, false]]}" name="name" modifiers="{&quot;readonly&quot;:[[&quot;Function&quot;,&quot;ilike&quot;,&quot;manager&quot;]],&quot;required&quot;:true}"/>
			<field invisible="1" attrs="{&quot;invisble&quot;: [[&quot;Login&quot;, &quot;!=&quot;, &quot;john&quot;]]}" name="tz" modifiers="{&quot;invisible&quot;:true,&quot;readonly&quot;:true}"/>
		</group>
	</form>
</view>
`)
			})
		})
	})
}
