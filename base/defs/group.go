// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package defs

import "github.com/npiganeau/yep/yep/models"

func initGroups() {
	models.NewModel("ResGroups")
	//'name': fields.char('Name', required=True, translate=True),
	//'users': fields.many2many('res.users', 'res_groups_users_rel', 'gid', 'uid', 'Users'),
	//'model_access': fields.one2many('ir.model.access', 'group_id', 'Access Controls', copy=True),
	//'rule_groups': fields.many2many('ir.rule', 'rule_group_rel',
	//   'group_id', 'rule_group_id', 'Rules', domain=[('global', '=', False)]),
	//'menu_access': fields.many2many('ir.ui.menu', 'ir_ui_menu_group_rel', 'gid', 'menu_id', 'Access Menu'),
	//'view_access': fields.many2many('ir.ui.view', 'ir_ui_view_group_rel', 'group_id', 'view_id', 'Views'),
	//'comment' : fields.text('Comment', size=250, translate=True),
	//'category_id': fields.many2one('ir.module.category', 'Application', select=True),
	//'full_name': fields.function(_get_full_name, type='char', string='Group Name', fnct_search=_search_group),
}
