// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import "github.com/hexya-erp/hexya/hexya/models/security"

var (
	GroupERPManager, GroupSystem, GroupUser, GroupMultiCompany, GroupMultiCurrency,
	GroupTechnicalFeatures, GroupPartnerManager, GroupPortal, GroupPublic *security.Group
)

func init() {
	GroupERPManager = security.Registry.NewGroup("group_erp_manager", "Access Rights")
	GroupSystem = security.Registry.NewGroup("group_system", "Settings", GroupERPManager)
	GroupUser = security.Registry.NewGroup("group_user", "Employee")
	GroupMultiCompany = security.Registry.NewGroup("group_multi_company", "Multi Companies")
	GroupMultiCurrency = security.Registry.NewGroup("group_multi_currency", "Multi Currencies")
	GroupTechnicalFeatures = security.Registry.NewGroup("group_no_one", "Technical Features")
	GroupPartnerManager = security.Registry.NewGroup("group_partner_manager", "Contact Creation")
	GroupPortal = security.Registry.NewGroup("group_portal", "Portal")
	GroupPublic = security.Registry.NewGroup("group_public", "Public")
}
