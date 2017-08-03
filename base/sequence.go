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
	pool.SequenceDateRange().AddDateField("DateFrom", models.SimpleFieldParams{String: "From", Required: true})
	pool.SequenceDateRange().AddDateField("DateTo", models.SimpleFieldParams{String: "True", Required: true})
	pool.SequenceDateRange().AddMany2OneField("Sequence", models.ForeignKeyFieldParams{RelationModel: pool.Sequence(),
		Required: true, OnDelete: models.Cascade})
	pool.SequenceDateRange().AddIntegerField("NumberNext", models.SimpleFieldParams{String: "Next Number",
		Required: true, Default: models.DefaultValue(1), Help: "Next number of this sequence"})
	pool.SequenceDateRange().AddIntegerField("NumberNextActual", models.SimpleFieldParams{Compute: "ComputeNumberNextActual",
		String: "Next Number", Help: "Next number that will be used. This number can be incremented frequently so the displayed value might already be obsolete"})

	pool.SequenceDateRange().Methods().ComputeNumberNextActual().DeclareMethod(
		`ComputeNumberNextActual returns the real next number for the sequence depending on the implementation`,
		func(rs pool.SequenceDateRangeSet) (*pool.SequenceDateRangeData, []models.FieldNamer) {
			var res pool.SequenceDateRangeData
			res.NumberNextActual = rs.NumberNext()
			return &res, []models.FieldNamer{pool.Sequence().NumberNextActual()}
		})

	pool.Sequence().DeclareModel()
	pool.Sequence().AddCharField("Name", models.StringFieldParams{Required: true})
	pool.Sequence().AddCharField("Code", models.StringFieldParams{})
	pool.Sequence().AddSelectionField("Implementation", models.SelectionFieldParams{Selection: types.Selection{"standard": "Standard", "no_gap": "No Gap"}, Required: true, Default: models.DefaultValue("standard"),
		Help: `Two sequence object implementations are offered: Standard and 'No gap'.
The latter is slower than the former but forbids any
gap in the sequence (while they are possible in the former).`})
	pool.Sequence().AddBooleanField("Action", models.SimpleFieldParams{Default: models.DefaultValue(true)})
	pool.Sequence().AddCharField("Prefix", models.StringFieldParams{Help: "Prefix value of the record for the sequence"})
	pool.Sequence().AddCharField("Suffix", models.StringFieldParams{Help: "Suffix value of the record for the sequence"})
	pool.Sequence().AddIntegerField("NumberNext", models.SimpleFieldParams{String: "Next Number", Required: true,
		Default: models.DefaultValue(1), Help: "Next number of this sequence"})
	pool.Sequence().AddIntegerField("NumberNextActual", models.SimpleFieldParams{Compute: "ComputeNumberNextActual",
		String: "Next Number", Help: "Next number that will be used. This number can be incremented frequently so the displayed value might already be obsolete"})
	pool.Sequence().AddIntegerField("NumberIncrement", models.SimpleFieldParams{String: "Step", Required: true,
		Default: models.DefaultValue(1), Help: "The next number of the sequence will be incremented by this number"})
	pool.Sequence().AddIntegerField("Padding", models.SimpleFieldParams{String: "Sequence Size", Required: true,
		Default: models.DefaultValue(0), Help: "Hexya will automatically adds some '0' on the left of the 'Next Number' to get the required padding size."})
	pool.Sequence().AddMany2OneField("Company", models.ForeignKeyFieldParams{RelationModel: pool.Company()}) // default=lambda s: s.env['res.company']._company_default_get('ir.sequence'))
	pool.Sequence().AddBooleanField("UseDateRange", models.SimpleFieldParams{String: "Use subsequences per Date Range"})
	pool.Sequence().AddOne2ManyField("DateRanges", models.ReverseFieldParams{RelationModel: pool.SequenceDateRange(), ReverseFK: "Sequence"})

	pool.Sequence().Methods().ComputeNumberNextActual().DeclareMethod(
		`ComputeNumberNextActual returns the real next number for the sequence depending on the implementation`,
		func(rs pool.SequenceSet) (*pool.SequenceData, []models.FieldNamer) {
			var res pool.SequenceData
			res.NumberNextActual = rs.NumberNext()
			return &res, []models.FieldNamer{pool.Sequence().NumberNextActual()}
		})
}
