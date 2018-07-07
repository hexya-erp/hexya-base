hexya.define('web_kanban.compatibility', function (require) {
"use strict";

var kanban_widgets = require('web_kanban.widgets');
var KanbanRecord = require('web_kanban.Record');
var KanbanColumn = require('web_kanban.Column');
var KanbanView = require('web_kanban.KanbanView');

return;
hexyaerp = window.hexyaerp || {};
hexyaerp.web_kanban = hexyaerp.web_kanban || {};
hexyaerp.web_kanban.AbstractField = kanban_widgets.AbstractField;
hexyaerp.web_kanban.KanbanGroup = KanbanColumn;
hexyaerp.web_kanban.KanbanRecord = KanbanRecord;
hexyaerp.web_kanban.KanbanView = KanbanView;

});
