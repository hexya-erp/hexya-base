// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package base

import (
	"fmt"
	"strings"
	"time"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
)

// Sequences maps Hexya formats to Go time format
var Sequences = map[string]string{
	"year": "2006", "month": "01", "day": "02", "y": "06",
	"h24": "15", "h12": "03", "min": "04", "sec": "05",
}

// SequenceFuncs maps Hexya formats to functions that must be applied to a time.Time object
var SequenceFuncs = map[string]func(time.Time) string{
	"doy": func(t time.Time) string {
		return fmt.Sprintf("%d", t.YearDay())
	},
	"woy": func(t time.Time) string {
		_, woy := t.ISOWeek()
		return fmt.Sprintf("%d", woy)
	},
	"weekday": func(t time.Time) string {
		return fmt.Sprintf("%d", int(t.Weekday()))
	},
}

func init() {
	h.Sequence().DeclareModel()
	h.Sequence().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{Required: true},
		"Code": models.CharField{String: "Sequence Code"},
		"Implementation": models.SelectionField{
			Selection: types.Selection{"standard": "Standard", "no_gap": "No Gap"}, Required: true,
			Default: models.DefaultValue("standard"),
			Help: `Two sequence object implementations are offered: Standard and 'No gap'.
The latter is slower than the former but forbids any
gap in the sequence (while they are possible in the former).`},
		"Active": models.BooleanField{Default: models.DefaultValue(true), Required: true},
		"Prefix": models.CharField{Help: "Prefix value of the record for the sequence"},
		"Suffix": models.CharField{Help: "Suffix value of the record for the sequence"},
		"NumberNext": models.IntegerField{String: "Next Number", Required: true,
			Default: models.DefaultValue(1), Help: "Next number of this sequence"},
		"NumberNextActual": models.IntegerField{
			Compute: h.Sequence().Methods().ComputeNumberNextActual(), String: "Next Number",
			Inverse: h.Sequence().Methods().InverseNumberNextActual(),
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
		"DateRanges": models.One2ManyField{RelationModel: h.SequenceDateRange(), ReverseFK: "Sequence",
			String: "Subsequences"},
	})

	h.Sequence().Methods().ComputeNumberNextActual().DeclareMethod(
		`ComputeNumberNextActual returns the real next number for the sequence depending on the implementation`,
		func(rs h.SequenceSet) *h.SequenceData {
			var res h.SequenceData
			res.NumberNextActual = rs.NumberNext()
			return &res
		})

	h.Sequence().Methods().InverseNumberNextActual().DeclareMethod(
		`InverseNumberNextActual is the setter function for the NumberNextActual field`,
		func(rs h.SequenceSet, value int64) {
			if value == 0 {
				value = 1
			}
			rs.SetNumberNext(value)
		})

	h.Sequence().Methods().Create().Extend("",
		func(rs h.SequenceSet, vals *h.SequenceData, fieldsToReset ...models.FieldNamer) h.SequenceSet {
			seq := rs.Super().Create(vals)
			if vals.Implementation == "standard" || vals.Implementation == "" {
				numberIncrement := vals.NumberIncrement
				if numberIncrement == 0 {
					numberIncrement = 1
				}
				numberNext := vals.NumberNext
				if numberNext == 0 {
					numberNext = 1
				}
				models.CreateSequence(fmt.Sprintf("sequence_%03d", seq.ID()), numberIncrement, numberNext)
			}
			return seq
		})

	h.Sequence().Methods().Unlink().Extend("",
		func(rs h.SequenceSet) int64 {
			for _, rec := range rs.Records() {
				hexyaSeq, exists := models.Registry.GetSequence(fmt.Sprintf("sequence_%03d", rec.ID()))
				if exists {
					hexyaSeq.Drop()
				}
			}
			return rs.Super().Unlink()
		})

	h.Sequence().Methods().Write().Extend("",
		func(rs h.SequenceSet, data *h.SequenceData, fieldsToReset ...models.FieldNamer) bool {
			newImplementation := data.Implementation
			for _, seq := range rs.Records() {
				// 4 cases: we test the previous impl. against the new one.
				i := data.NumberIncrement
				if i == 0 {
					i = seq.NumberIncrement()
				}
				n := data.NumberNext
				if n == 0 {
					n = seq.NumberNext()
				}
				if seq.Implementation() == "standard" {
					hexyaSeq := models.Registry.MustGetSequence(fmt.Sprintf("sequence_%03d", seq.ID()))
					if newImplementation == "standard" || newImplementation == "" {
						// Implementation has NOT changed.
						// Only change sequence if really requested.
						if data.NumberNext != 0 || seq.NumberIncrement() != i {
							hexyaSeq.Alter(i, n)
						}
					} else {
						hexyaSeq.Drop()
						for _, subSeq := range seq.DateRanges().Records() {
							subHexyaSeq := models.Registry.MustGetSequence(fmt.Sprintf("sequence_%03d_%03d", seq.ID(), subSeq.ID()))
							subHexyaSeq.Drop()
						}
					}

					continue
				}
				if newImplementation == "no_gap" || newImplementation == "" {
					continue
				}
				models.CreateSequence(fmt.Sprintf("sequence_%03d", seq.ID()), i, n)
				for _, subSeq := range seq.DateRanges().Records() {
					models.CreateSequence(fmt.Sprintf("sequence_%03d_%03d", seq.ID(), subSeq.ID()), i, n)
				}
			}
			return rs.Super().Write(data, fieldsToReset...)
		})

	h.Sequence().Methods().NextDo().DeclareMethod(
		`NextDo returns the next sequence number formatted`,
		func(rs h.SequenceSet) string {
			rs.EnsureOne()
			if rs.Implementation() == "standard" {
				hexyaSeq := models.Registry.MustGetSequence(fmt.Sprintf("sequence_%03d", rs.ID()))
				return rs.GetNextChar(hexyaSeq.NextValue())
			}
			return rs.GetNextChar(rs.UpdateNoGap())
		})

	h.Sequence().Methods().UpdateNoGap().DeclareMethod(
		`UpdateNoGap gets the next number of a "No Gap" sequence`,
		func(rs h.SequenceSet) int64 {
			rs.EnsureOne()
			numberNext := rs.NumberNext()
			rs.Env().Cr().Execute(`SELECT number_next FROM sequence WHERE id=? FOR UPDATE NOWAIT`, rs.ID())
			rs.Env().Cr().Execute(`UPDATE sequence SET number_next=number_next + ? WHERE id=?`, rs.NumberIncrement(), rs.ID())
			rs.InvalidateCache()
			return numberNext
		})

	h.Sequence().Methods().GetNextChar().DeclareMethod(
		`GetNextChar returns the given number formatted as per the sequence data`,
		func(rs h.SequenceSet, numberNext int64) string {
			interpolate := func(format string, data map[string]string) string {
				if format == "" {
					return ""
				}
				res := format
				for k, v := range data {
					res = strings.Replace(res, fmt.Sprintf("%%(%s)", k), v, -1)
				}
				return res
			}
			interpolateMap := func() map[string]string {
				location, err := time.LoadLocation(rs.Env().Context().GetString("tz"))
				if err != nil {
					location = time.UTC
				}
				now := time.Now().In(location)
				rangeDate, effectiveDate := now, now
				if rs.Env().Context().HasKey("sequence_date") {
					effectiveDate = rs.Env().Context().GetDate("sequence_date").Time
				}
				if rs.Env().Context().HasKey("sequence_date_range") {
					rangeDate = rs.Env().Context().GetDate("sequence_date_range").Time
				}

				res := make(map[string]string)
				for key, format := range Sequences {
					res[key] = effectiveDate.Format(format)
					res["range_"+key] = rangeDate.Format(format)
					res["current_"+key] = now.Format(format)
				}
				for key, fFunc := range SequenceFuncs {
					res[key] = fFunc(effectiveDate)
					res["range_"+key] = fFunc(rangeDate)
					res["current_"+key] = fFunc(now)
				}
				return res
			}
			d := interpolateMap()
			interpolatedPrefix := interpolate(rs.Prefix(), d)
			interpolatedSuffix := interpolate(rs.Suffix(), d)
			return interpolatedPrefix +
				fmt.Sprintf(fmt.Sprintf("%%0%dd", rs.Padding()), numberNext) +
				interpolatedSuffix
		})

	h.Sequence().Methods().CreateDateRangeSeq().DeclareMethod(
		`CreateDateRangeSeq creates the date range for the given date`,
		func(rs h.SequenceSet, date dates.Date) h.SequenceDateRangeSet {
			rs.EnsureOne()
			year := date.Year()
			dateFrom := dates.ParseDate(fmt.Sprintf("%d-01-01", year))
			dateTo := dates.ParseDate(fmt.Sprintf("%d-12-31", year))
			dateRange := h.SequenceDateRange().Search(rs.Env(),
				q.SequenceDateRange().Sequence().Equals(rs).
					And().DateFrom().GreaterOrEqual(date).
					And().DateFrom().LowerOrEqual(dateTo)).
				OrderBy("DateFrom DESC").
				Limit(1)
			if !dateRange.IsEmpty() {
				dateTo = dateRange.DateFrom().AddDate(0, 0, -1)
			}
			dateRange = h.SequenceDateRange().Search(rs.Env(),
				q.SequenceDateRange().Sequence().Equals(rs).
					And().DateTo().GreaterOrEqual(dateFrom).
					And().DateTo().LowerOrEqual(date)).
				OrderBy("DateTo DESC").
				Limit(1)
			if !dateRange.IsEmpty() {
				dateTo = dateRange.DateTo().AddDate(0, 0, 1)
			}
			seqDateRange := h.SequenceDateRange().Create(rs.Env(), &h.SequenceDateRangeData{
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Sequence: rs,
			})
			return seqDateRange
		})

	h.Sequence().Methods().Next().DeclareMethod(
		`Next returns the next number (formatted) in the preferred sequence in all the ones given in self`,
		func(rs h.SequenceSet) string {
			rs.EnsureOne()
			if !rs.UseDateRange() {
				return rs.NextDo()
			}
			// Date mode
			dt := dates.Today()
			if rs.Env().Context().HasKey("sequence_date") {
				dt = rs.Env().Context().GetDate("sequence_date")
			}
			seqDate := h.SequenceDateRange().Search(rs.Env(),
				q.SequenceDateRange().Sequence().Equals(rs).
					And().DateFrom().LowerOrEqual(dt).
					And().DateTo().GreaterOrEqual(dt)).
				Limit(1)
			if seqDate.IsEmpty() {
				seqDate = rs.CreateDateRangeSeq(dt)
			}
			return seqDate.WithContext("sequence_date_range", seqDate.DateFrom()).Next()
		})

	h.Sequence().Methods().NextByID().DeclareMethod(
		`NextByID draws an interpolated string using the specified sequence.`,
		func(rs h.SequenceSet) string {
			rs.CheckExecutionPermission(h.Sequence().Methods().Read().Underlying())
			return rs.Next()
		})

	h.Sequence().Methods().NextByCode().DeclareMethod(
		`NextByCode draws an interpolated string using a sequence with the requested code.
		If several sequences with the correct code are available to the user
		(multi-company cases), the one from the user's current company will be used.

		The context may contain a 'force_company' key with the ID of the company to
		use instead of the user's current company for the sequence selection. 
		A matching sequence for that specific company will get higher priority`,
		func(rs h.SequenceSet, sequenceCode string) string {
			rs.CheckExecutionPermission(h.Sequence().Methods().Read().Underlying())
			companies := h.Company().NewSet(rs.Env()).SearchAll()
			seqs := h.Sequence().Search(rs.Env(),
				q.Sequence().Code().Equals(sequenceCode).AndCond(
					q.Sequence().Company().In(companies).Or().Company().IsNull()))
			if seqs.IsEmpty() {
				log.Debug("No Sequence has been found for this code", "code", sequenceCode, "companies", companies)
				return "False"
			}
			forceCompanyID := rs.Env().Context().GetInteger("force_company")
			if forceCompanyID == 0 {
				forceCompanyID = h.User().NewSet(rs.Env()).CurrentUser().Company().ID()
			}
			for _, seq := range seqs.Records() {
				if seq.Company().ID() == forceCompanyID {
					return seq.Next()
				}
			}
			return seqs.Records()[0].Next()
		})

	h.SequenceDateRange().DeclareModel()
	h.SequenceDateRange().AddFields(map[string]models.FieldDefinition{
		"DateFrom": models.DateField{String: "From", Required: true},
		"DateTo":   models.DateField{String: "To", Required: true},
		"Sequence": models.Many2OneField{String: "Main Sequence", RelationModel: h.Sequence(),
			Required: true, OnDelete: models.Cascade},
		"NumberNext": models.IntegerField{String: "Next Number",
			Required: true, Default: models.DefaultValue(1), Help: "Next number of this sequence"},
		"NumberNextActual": models.IntegerField{String: "Next Number",
			Compute: h.SequenceDateRange().Methods().ComputeNumberNextActual(),
			Inverse: h.SequenceDateRange().Methods().InverseNumberNextActual(),
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

	h.SequenceDateRange().Methods().InverseNumberNextActual().DeclareMethod(
		`InverseNumberNextActual is the setter function for the NumberNextActual field`,
		func(rs h.SequenceDateRangeSet, value int64) {
			if value == 0 {
				value = 1
			}
			rs.SetNumberNext(value)
		})

	h.SequenceDateRange().Methods().Next().DeclareMethod(
		`Next returns the next number (formatted) of this sequence date range.`,
		func(rs h.SequenceDateRangeSet) string {
			if rs.Sequence().Implementation() == "standard" {
				hexyaSeq := models.Registry.MustGetSequence(fmt.Sprintf("sequence_%03d_%03d", rs.Sequence().ID(), rs.ID()))
				return rs.Sequence().GetNextChar(hexyaSeq.NextValue())
			}
			return rs.Sequence().GetNextChar(rs.UpdateNoGap())
		})

	h.SequenceDateRange().Methods().Create().Extend("",
		func(rs h.SequenceDateRangeSet, data *h.SequenceDateRangeData, fieldsToReset ...models.FieldNamer) h.SequenceDateRangeSet {
			seq := rs.Super().Create(data)
			mainSeq := seq.Sequence()
			if mainSeq.Implementation() == "standard" {
				next := data.NumberNextActual
				if next == 0 {
					next = 1
				}
				models.CreateSequence(fmt.Sprintf("sequence_%03d_%03d", mainSeq.ID(), seq.ID()),
					mainSeq.NumberIncrement(), next)
			}
			return seq
		})

	h.SequenceDateRange().Methods().Unlink().Extend("",
		func(rs h.SequenceDateRangeSet) int64 {
			for _, rec := range rs.Records() {
				hexyaSeq, exists := models.Registry.GetSequence(fmt.Sprintf("sequence_%03d_%03d", rec.Sequence().ID(), rec.ID()))
				if exists {
					hexyaSeq.Drop()
				}
			}
			return rs.Super().Unlink()
		})

	h.SequenceDateRange().Methods().Write().Extend("",
		func(rs h.SequenceDateRangeSet, data *h.SequenceDateRangeData, fieldsToReset ...models.FieldNamer) bool {
			if data.NumberNext != 0 {
				seqToAlter := rs.Filtered(func(rs h.SequenceDateRangeSet) bool {
					return rs.Sequence().Implementation() == "standard"
				})
				for _, rec := range seqToAlter.Records() {
					hexyaSeq, exists := models.Registry.GetSequence(fmt.Sprintf("sequence_%03d_%03d", rec.Sequence().ID(), rec.ID()))
					if exists {
						hexyaSeq.Alter(data.NumberNext, 0)
					}
				}
			}
			return rs.Super().Write(data, fieldsToReset...)
		})

	h.SequenceDateRange().Methods().UpdateNoGap().DeclareMethod(
		`UpdateNoGap gets the next number of a "No Gap" sequence`,
		func(rs h.SequenceDateRangeSet) int64 {
			rs.EnsureOne()
			numberNext := rs.NumberNext()
			rs.Env().Cr().Execute(`SELECT number_next FROM sequence_date_range WHERE id=? FOR UPDATE NOWAIT`, rs.ID())
			rs.Env().Cr().Execute(`UPDATE sequence_date_range SET number_next=number_next + ? WHERE id=?`, rs.Sequence().NumberIncrement(), rs.ID())
			rs.InvalidateCache()
			return numberNext
		})

}
