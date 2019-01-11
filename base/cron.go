package base

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/models"

	"time"

	"encoding/json"

	"strings"

	"strconv"

	"sort"

	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"
)

var schedulerUpdateChan chan bool

func init() {
	h.Cron().DeclareModel()
	h.Cron().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{
			String:     "Cron",
			Unique:     true,
			Required:   true,
			Default:    models.DefaultValue("Scheduled Func"),
			Constraint: h.Cron().Methods().ConstraintCronCreation()},
		"ModelMethodStr": models.CharField{
			String:  "Method",
			Compute: h.Cron().Methods().ComputeModelMethodStr()},
		"Status": models.BooleanField{
			Default: models.DefaultValue(true)},
		"ExecuteCount": models.IntegerField{
			ReadOnly: true},
		"ExecuteNET": models.CharField{
			Compute:  h.Cron().Methods().ComputeExecuteNET(),
			ReadOnly: true},
		/* ---------------------------------------------------------------------------------------------------------- */
		"TargetWorker": models.Many2OneField{
			Required:      true,
			RelationModel: h.Worker()},
		"TargetModel": models.CharField{
			String:   "Model",
			Required: true},
		"TargetMethod": models.CharField{
			String:   "Method",
			Required: true},
		"TargetParams": models.TextField{
			String:  "Parameters",
			Default: models.DefaultValue("[]")},
		/* ---------------------------------------------------------------------------------------------------------- */
		"TimeAtDate": models.DateTimeField{
			Default: models.DefaultValue(dates.Now())},
		"Catchup": models.BooleanField{},
		/* ---------------------------------------------------------------------------------------------------------- */
		"RepeatBool": models.BooleanField{},
		"RepeatLapseAmmount": models.IntegerField{
			Default: models.DefaultValue(2)},
		"RepeatLapseSelection": models.SelectionField{
			Selection: types.Selection{
				`month`:  `Months`,
				`day`:    `Days`,
				`hour`:   `Hours`,
				`minute`: `Minutes`,
				`second`: `Seconds`,
			},
			Default:  models.DefaultValue("minute"),
			Required: true},
		"RepeatAmmountBool": models.BooleanField{},
		"RepeatAmmount": models.IntegerField{
			Default: models.DefaultValue(5)},
		/* ---------------------------------------------------------------------------------------------------------- */
		"MaskBool": models.BooleanField{},
		"Mask": models.Many2OneField{
			RelationModel: h.CronTimeMask(),
		},
	})

	h.Cron().Methods().ConstraintCronCreation().DeclareMethod(
		`ConstrantCronCreation verifies Cron's field to be valid`,
		func(rs h.CronSet) {
			var out string
			if h.Worker().NewSet(rs.Env()).GetWorker(rs.TargetWorker().Name()).Len() == 0 {
				out += fmt.Sprintf("No Worker found with name '%s'\n", rs.TargetWorker())
			}
			model, ok := models.Registry.Get(rs.TargetModel())
			if !ok {
				out += fmt.Sprintf("No Model found with name '%s'\n", rs.TargetModel())
			} else if _, ok := model.Methods().Get(rs.TargetMethod()); !ok {
				out += fmt.Sprintf("No Method found in '%s' as '%s'\n", rs.TargetModel(), rs.TargetMethod())
			}
			var js interface{}
			if err := json.Unmarshal([]byte(rs.TargetParams()), &js); err != nil {
				out += fmt.Sprintf("Parameters could not be Unmarshalled: %v\n", err)
			}
			if rs.RepeatBool() {
				if rs.RepeatLapseAmmount() < 0 {
					out += fmt.Sprintln("Lapse ammount can't be negative")
				}
				if rs.RepeatAmmount() < 0 {
					out += fmt.Sprintln("Repeat ammount can't be negative")
				}
			}
			if out != "" {
				panic(out)
			}
		})

	h.Cron().Methods().ComputeModelMethodStr().DeclareMethod(
		`ComputeModelMethodStr returns a string with format "<Model> - <Method>". Mainly Used in Cron views`,
		func(rs h.CronSet) h.CronData {
			return h.CronData{ModelMethodStr: fmt.Sprintf("%s - %s", rs.TargetModel(), rs.TargetMethod())}
		})

	h.Cron().Methods().ComputeExecuteNET().DeclareMethod(
		`ComputeExecuteNET calculates when the Cron will be called next and
				returns a CronData with its ExecuteNET filled with result string`,
		func(rs h.CronSet) h.CronData {
			etaDateTime := rs.TimeAtDate().Sub(dates.Now())
			if etaDateTime.Seconds() == 0 {
				return h.CronData{ExecuteNET: "ASAP"}
			}
			days := etaDateTime / (24 * time.Hour)
			etaDateTime = etaDateTime % (24 * time.Hour)
			hours := etaDateTime / time.Hour
			etaDateTime = etaDateTime % time.Hour
			minutes := etaDateTime / time.Minute
			etaDateTime = etaDateTime % time.Minute
			seconds := etaDateTime / time.Second
			o := []rune(fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds))
			var out []rune
			var switc bool
			for _, c := range o {
				if c >= '1' && c <= '9' {
					switc = true
				}
				if switc {
					out = append(out, c)
				} else {
					out = append(out, ' ')
				}
			}
			outStr := strings.TrimSpace(string(out))
			if outStr == "" || strings.Contains(outStr, "-") {
				outStr = "Now"
			}
			return h.CronData{ExecuteNET: string(outStr)}
		})

	h.Cron().Methods().ButtonResume().DeclareMethod(
		`ButtonResume sets rs status as active and syncronize`,
		func(rs h.CronSet) {
			rs.SetStatus(true)
			rs.ButtonRefresh()
		})

	h.Cron().Methods().ButtonSuspend().DeclareMethod(
		`ButtonSuspend sets rs status as inactive and syncronize`,
		func(rs h.CronSet) {
			rs.SetStatus(false)
			rs.ButtonRefresh()
		})

	h.Cron().Methods().ButtonRun().DeclareMethod(
		`ButtonRun pushes the pointed job it's worker's queue`,
		func(rs h.CronSet) {
			var param interface{}
			json.Unmarshal([]byte(rs.TargetParams()), &param)
			params := interfaceSlice(param)
			h.Worker().NewSet(rs.Env()).WithWorker(rs.TargetWorker().Name()).WithParams(params...).Enqueue(models.Registry.MustGet(rs.TargetModel()).Methods().MustGet(rs.TargetMethod()))
		})

	h.Cron().Methods().ButtonRefresh().DeclareMethod(
		`ButtonRefresh syncronizes the cron`,
		func(rs h.CronSet) {
			go func() {
				time.Sleep(50 * time.Millisecond)
				schedulerUpdateChan <- true
			}()
		})

	h.Cron().Methods().Create().Extend(
		`Create creates a new Cron entry`,
		func(rs h.CronSet, data *h.CronData, namer ...models.FieldNamer) h.CronSet {
			out := rs.Super().Create(data, namer...)
			rs.ButtonRefresh()
			return out
		})

	h.Cron().Methods().StartScheduler().DeclareMethod(
		`StartScheduler starts the Cron. Meant to be used once after server init`,
		func(rs h.CronSet) {
			if h.Cron().Search(rs.Env(), q.Cron().Name().Equals("Removal of old finished Jobs")).IsEmpty() {
				h.Cron().Create(rs.Env(), &h.CronData{
					Name:                 "Removal of old finished Jobs",
					TargetWorker:         h.Worker().NewSet(rs.Env()).GetWorker("Main"),
					TargetModel:          "Worker",
					TargetMethod:         "CleanJobHistory",
					TargetParams:         "[]",
					TimeAtDate:           dates.ParseDateTime("2000-01-01 00:00:00"),
					RepeatBool:           true,
					RepeatLapseAmmount:   1,
					RepeatLapseSelection: "day",
				})
			}
			go rs.SchedulerLoop(15 * time.Minute)
			time.Sleep(50 * time.Millisecond)
			schedulerUpdateChan <- true
		})

	h.Cron().Methods().SchedulerLoop().DeclareMethod(
		`SchedulerLoop starts the cron main loop. Shouldn't be directly called. Use StartScheduler instead'`,
		func(rs h.CronSet, next time.Duration) {
			def := next
			schedulerUpdateChan = make(chan bool)
			for {
				select {
				case <-time.After(next):
					next = h.Cron().NewSet(rs.Env()).Sync()
				case <-schedulerUpdateChan:
					next = h.Cron().NewSet(rs.Env()).Sync()
				case <-time.After(def):
					next = h.Cron().NewSet(rs.Env()).Sync()
				}
			}
		})

	h.Cron().Methods().CheckTimeMask().DeclareMethod(
		`CheckTimeMask returns true if the given cron entry's next call follows the entry's time mask`,
		func(rs h.CronSet, data h.CronData) bool {
			if data.MaskBool {
				if data.Mask.MonthBool() {
					curMonth := string([]byte(data.TimeAtDate.Month().String()))[:3]
					if !data.Mask.Get(curMonth).(bool) {
						return false
					}
				}
				if data.Mask.WeekDayBool() {
					curWD := string([]byte(data.TimeAtDate.Weekday().String()))[:3]
					if !data.Mask.Get(curWD).(bool) {
						return false
					}
				}
				if data.Mask.DayBool() {
					str := "," + data.Mask.DayStr() + ","
					if !strings.Contains(str, ","+strconv.Itoa(data.TimeAtDate.Day())+",") {
						return false
					}
				}
				if data.Mask.HourBool() {
					str := "," + data.Mask.HourStr() + ","
					if !strings.Contains(str, ","+strconv.Itoa(data.TimeAtDate.Hour())+",") {
						return false
					}
				}
				if data.Mask.MinuteBool() {
					str := "," + data.Mask.MinuteStr() + ","
					if !strings.Contains(str, ","+strconv.Itoa(data.TimeAtDate.Minute())+",") {
						return false
					}
				}
			}
			return true
		})

	h.Cron().Methods().Sync().DeclareMethod(
		`Sync launches all entries running late and calculates the next sync call time`,
		func(rs h.CronSet) time.Duration {
			out := float64(15 * 60)
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env models.Environment) {
				for _, rec := range h.Cron().Search(env, q.Cron().Status().Equals(true)).Records() {
					data := rec.First()
					funcExecuted := false
					for (data.TimeAtDate.LowerEqual(dates.Now().Add(time.Second)) || !rs.CheckTimeMask(data)) && data.Status {
						if !funcExecuted && rs.CheckTimeMask(data) {
							var param interface{}
							json.Unmarshal([]byte(data.TargetParams), &param)
							params := interfaceSlice(param)
							h.Worker().NewSet(env).WithWorker(data.TargetWorker.Name()).WithParams(params...).Enqueue(models.Registry.MustGet(data.TargetModel).Methods().MustGet(data.TargetMethod))
							if !data.RepeatBool {
								data.Status = false
							}
							if !data.Catchup {
								funcExecuted = true
							}
							data.ExecuteCount += 1
							if data.RepeatAmmountBool {
								data.RepeatAmmount -= 1
								if data.RepeatAmmount == 0 {
									data.Status = false
								}
							}
						}
						switch data.RepeatLapseSelection {
						case `month`:
							data.TimeAtDate = data.TimeAtDate.AddDate(0, int(data.RepeatLapseAmmount), 0)
						case `day`:
							data.TimeAtDate = data.TimeAtDate.AddDate(0, 0, int(data.RepeatLapseAmmount))
						case `hour`:
							data.TimeAtDate = data.TimeAtDate.Add(time.Duration(data.RepeatLapseAmmount) * time.Hour)
						case `minute`:
							data.TimeAtDate = data.TimeAtDate.Add(time.Duration(data.RepeatLapseAmmount) * time.Minute)
						case `second`:
							data.TimeAtDate = data.TimeAtDate.Add(time.Duration(data.RepeatLapseAmmount) * time.Second)
						}
					}
					lapse := data.TimeAtDate.Sub(dates.Now()).Seconds()
					if lapse < out {
						out = lapse
					}
					rec.Write(&h.CronData{
						TimeAtDate:    data.TimeAtDate,
						ExecuteCount:  data.ExecuteCount,
						RepeatAmmount: data.RepeatAmmount,
						Status:        data.Status},
						h.Cron().Status())
				}
			})
			return time.Duration(out) * time.Second
		})

	h.CronTimeMask().DeclareModel()

	h.CronTimeMask().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{
			Required: true,
			Unique:   true},
		"MonthBool":   models.BooleanField{},
		"Jan":         models.BooleanField{},
		"Feb":         models.BooleanField{},
		"Mar":         models.BooleanField{},
		"Apr":         models.BooleanField{},
		"May":         models.BooleanField{},
		"Jun":         models.BooleanField{},
		"Jul":         models.BooleanField{},
		"Aug":         models.BooleanField{},
		"Sep":         models.BooleanField{},
		"Oct":         models.BooleanField{},
		"Nov":         models.BooleanField{},
		"Dec":         models.BooleanField{},
		"WeekDayBool": models.BooleanField{},
		"Mon":         models.BooleanField{},
		"Tue":         models.BooleanField{},
		"Wed":         models.BooleanField{},
		"Thu":         models.BooleanField{},
		"Fri":         models.BooleanField{},
		"Sat":         models.BooleanField{},
		"Sun":         models.BooleanField{},
		"DayBool":     models.BooleanField{},
		"DayStr":      models.CharField{},
		"HourBool":    models.BooleanField{},
		"HourStr":     models.CharField{},
		"MinuteBool":  models.BooleanField{},
		"MinuteStr":   models.CharField{},
	})

	h.CronTimeMask().Methods().Create().Extend(
		`Create creates a new CronTimeMask`,
		func(rs h.CronTimeMaskSet, data *h.CronTimeMaskData, namer ...models.FieldNamer) h.CronTimeMaskSet {
			data.DayStr = rs.CompileNbStr(data.DayStr, 1, 31)
			data.HourStr = rs.CompileNbStr(data.HourStr, 0, 23)
			data.MinuteStr = rs.CompileNbStr(data.MinuteStr, 0, 59)
			out := rs.Super().Create(data, namer...)
			return out
		})

	h.CronTimeMask().Methods().CompileNbStr().DeclareMethod(
		`CompuleNbStr transforms the given NbStr to a generic readable one. errors poorly handled`,
		func(rs h.CronTimeMaskSet, str string, min, max int) string {
			spl := strings.Split(str, ",")
			intSl := []int{}
			for _, s := range spl {
				s = strings.TrimSpace(s)
				if strings.ContainsRune(s, '-') {
					sp := strings.Split(s, "-")
					int1, err1 := strconv.Atoi(sp[0])
					int2, err2 := strconv.Atoi(sp[1])
					if err1 != nil || err2 != nil {
						continue
					}
					if int2 < int1 {
						int2 = int2 + max + 1
					}
					for ; int1 <= int2; int1++ {
						if int1 > min {
							intSl = append(intSl, int1%(max+1))
						}
					}
				} else {
					i, err := strconv.Atoi(s)
					if err == nil && i > min {
						intSl = append(intSl, i%(max+1))
					}
				}
			}
			sort.Ints(intSl)
			var out []byte
			for i, n := range intSl {
				if n >= min {
					if i != 0 {
						out = append(out, ',')
					}
					out = append(out, []byte(strconv.Itoa(n))...)
				}
			}
			return string(out)
		})

}
