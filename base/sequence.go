// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/pool/h"
)

func init() {
	h.SequenceDateRange().DeclareModel()
	h.SequenceDateRange().AddFields(map[string]models.FieldDefinition{
		"DateFrom": models.DateField{String: "From", Required: true},
		"DateTo":   models.DateField{String: "True", Required: true},
		"Sequence": models.Many2OneField{RelationModel: h.Sequence(),
			Required: true, OnDelete: models.Cascade},
		"NumberNext": models.IntegerField{String: "Next Number",
			Required: true, Default: models.DefaultValue(1), Help: "Next number of this sequence"},
		"NumberNextActual": models.IntegerField{
			Compute: h.SequenceDateRange().Methods().ComputeNumberNextActual(), String: "Next Number",
			Help:    "Next number that will be used. This number can be incremented frequently so the displayed value might already be obsolete",
			Depends: []string{"NumberNext"}},
	})

	h.SequenceDateRange().Methods().ComputeNumberNextActual().DeclareMethod(
		`ComputeNumberNextActual returns the real next number for the sequence depending on the implementation`,
		func(rs h.SequenceDateRangeSet) *h.SequenceDateRangeData {
			var res h.SequenceDateRangeData
			res.NumberNextActual = rs.NumberNext()
			return &res
		})

	h.Sequence().DeclareModel()
	h.Sequence().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{Required: true},
		"Code": models.CharField{},
		"Implementation": models.SelectionField{
			Selection: types.Selection{"standard": "Standard", "no_gap": "No Gap"}, Required: true,
			Default: models.DefaultValue("standard"),
			Help: `Two sequence object implementations are offered: Standard and 'No gap'.
The latter is slower than the former but forbids any
gap in the sequence (while they are possible in the former).`},
		"Active": models.BooleanField{Default: models.DefaultValue(true)},
		"Prefix": models.CharField{Help: "Prefix value of the record for the sequence"},
		"Suffix": models.CharField{Help: "Suffix value of the record for the sequence"},
		"NumberNext": models.IntegerField{String: "Next Number", Required: true,
			Default: models.DefaultValue(1), Help: "Next number of this sequence"},
		"NumberNextActual": models.IntegerField{
			Compute: h.Sequence().Methods().ComputeNumberNextActual(), String: "Next Number",
			Help:    "Next number that will be used. This number can be incremented frequently so the displayed value might already be obsolete",
			Depends: []string{"NumberNext"}},
		"NumberIncrement": models.IntegerField{String: "Step", Required: true,
			Default: models.DefaultValue(1), Help: "The next number of the sequence will be incremented by this number"},
		"Padding": models.IntegerField{String: "Sequence Size", Required: true,
			Default: models.DefaultValue(0),
			Help:    "Hexya will automatically adds some '0' on the left of the 'Next Number' to get the required padding size."},
		"Company": models.Many2OneField{RelationModel: h.Company(), Default: func(env models.Environment) interface{} {
			return h.Company().NewSet(env).CompanyDefaultGet()
		}},
		"UseDateRange": models.BooleanField{String: "Use subsequences per Date Range"},
		"DateRanges":   models.One2ManyField{RelationModel: h.SequenceDateRange(), ReverseFK: "Sequence"},
	})

	h.Sequence().Methods().ComputeNumberNextActual().DeclareMethod(
		`ComputeNumberNextActual returns the real next number for the sequence depending on the implementation`,
		func(rs h.SequenceSet) *h.SequenceData {
			var res h.SequenceData
			res.NumberNextActual = rs.NumberNext()
			return &res
		})

	h.Sequence().Methods().NextByID().DeclareMethod(
		`NextByID draws an interpolated string using the specified sequence.`,
		func(rs h.SequenceSet) string {
			/*
				    @api.multi
					def next_by_id(self):
						""" Draw an interpolated string using the specified sequence."""
						self.check_access_rights('read')
						return self._next()
			*/
			return ""
		})

	h.Sequence().Methods().NextByCode().DeclareMethod(
		`NextByID draw an interpolated string using a sequence with the requested code.
		If several sequences with the correct code are available to the user
		(multi-company cases), the one from the user's current company will be used.`,
		func(rs h.SequenceSet, sequenceCode string) string {
			/*
				@api.model
				def next_by_code(self, sequence_code):
					""" Draw an interpolated string using a sequence with the requested code.
						If several sequences with the correct code are available to the user
						(multi-company cases), the one from the user's current company will
						be used.

						:param dict context: context dictionary may contain a
							``force_company`` key with the ID of the company to
							use instead of the user's current company for the
							sequence selection. A matching sequence for that
							specific company will get higher priority.
					"""
					self.check_access_rights('read')
					company_ids = self.env['res.company'].search([]).ids + [False]
					seq_ids = self.search(['&', ('code', '=', sequence_code), ('company_id', 'in', company_ids)])
					if not seq_ids:
						_logger.debug("No ir.sequence has been found for code '%s'. Please make sure a sequence is set for current company." % sequence_code)
						return False
					force_company = self._context.get('force_company')
					if not force_company:
						force_company = self.env.user.company_id.id
					preferred_sequences = [s for s in seq_ids if s.company_id and s.company_id.id == force_company]
					seq_id = preferred_sequences[0] if preferred_sequences else seq_ids[0]
					return seq_id._next()
			*/
			return ""
		})
}
