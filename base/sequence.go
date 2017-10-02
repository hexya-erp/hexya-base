// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool"
)

func init() {
	pool.SequenceDateRange().DeclareModel()
	pool.SequenceDateRange().AddFields(map[string]models.FieldDefinition{
		"DateFrom": models.DateField{String: "From", Required: true},
		"DateTo":   models.DateField{String: "True", Required: true},
		"Sequence": models.Many2OneField{RelationModel: pool.Sequence(),
			Required: true, OnDelete: models.Cascade},
		"NumberNext": models.IntegerField{String: "Next Number",
			Required: true, Default: models.DefaultValue(1), Help: "Next number of this sequence"},
		"NumberNextActual": models.IntegerField{
			Compute: pool.SequenceDateRange().Methods().ComputeNumberNextActual(), String: "Next Number",
			Help:    "Next number that will be used. This number can be incremented frequently so the displayed value might already be obsolete",
			Depends: []string{"NumberNext"}},
	})

	pool.SequenceDateRange().Methods().ComputeNumberNextActual().DeclareMethod(
		`ComputeNumberNextActual returns the real next number for the sequence depending on the implementation`,
		func(rs pool.SequenceDateRangeSet) (*pool.SequenceDateRangeData, []models.FieldNamer) {
			var res pool.SequenceDateRangeData
			res.NumberNextActual = rs.NumberNext()
			return &res, []models.FieldNamer{pool.Sequence().NumberNextActual()}
		})

	pool.Sequence().DeclareModel()
	pool.Sequence().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{Required: true},
		"Code": models.CharField{},
		"Implementation": models.SelectionField{
			Selection: types.Selection{"standard": "Standard", "no_gap": "No Gap"}, Required: true,
			Default: models.DefaultValue("standard"),
			Help: `Two sequence object implementations are offered: Standard and 'No gap'.
The latter is slower than the former but forbids any
gap in the sequence (while they are possible in the former).`},
		"Action": models.BooleanField{Default: models.DefaultValue(true)},
		"Prefix": models.CharField{Help: "Prefix value of the record for the sequence"},
		"Suffix": models.CharField{Help: "Suffix value of the record for the sequence"},
		"NumberNext": models.IntegerField{String: "Next Number", Required: true,
			Default: models.DefaultValue(1), Help: "Next number of this sequence"},
		"NumberNextActual": models.IntegerField{
			Compute: pool.Sequence().Methods().ComputeNumberNextActual(), String: "Next Number",
			Help:    "Next number that will be used. This number can be incremented frequently so the displayed value might already be obsolete",
			Depends: []string{"NumberNext"}},
		"NumberIncrement": models.IntegerField{String: "Step", Required: true,
			Default: models.DefaultValue(1), Help: "The next number of the sequence will be incremented by this number"},
		"Padding": models.IntegerField{String: "Sequence Size", Required: true,
			Default: models.DefaultValue(0),
			Help:    "Hexya will automatically adds some '0' on the left of the 'Next Number' to get the required padding size."},
		"Company": models.Many2OneField{RelationModel: pool.Company(), Default: func(env models.Environment, maps models.FieldMap) interface{} {
			return pool.Company().NewSet(env).CompanyDefaultGet()
		}},
		"UseDateRange": models.BooleanField{String: "Use subsequences per Date Range"},
		"DateRanges":   models.One2ManyField{RelationModel: pool.SequenceDateRange(), ReverseFK: "Sequence"},
	})

	pool.Sequence().Methods().ComputeNumberNextActual().DeclareMethod(
		`ComputeNumberNextActual returns the real next number for the sequence depending on the implementation`,
		func(rs pool.SequenceSet) (*pool.SequenceData, []models.FieldNamer) {
			var res pool.SequenceData
			res.NumberNextActual = rs.NumberNext()
			return &res, []models.FieldNamer{pool.Sequence().NumberNextActual()}
		})
}
