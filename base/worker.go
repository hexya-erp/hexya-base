package base

import (
	"fmt"
	"time"

	"reflect"

	"encoding/json"

	"strings"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	"github.com/hexya-erp/hexya/pool/h"
	"github.com/hexya-erp/hexya/pool/q"

	"github.com/google/uuid"
)

func init() {
	h.Worker().DeclareModel()
	h.Worker().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{
			String:   "Worker",
			Unique:   true,
			Required: true},
		"MaxThreads": models.IntegerField{
			String:   "Threads",
			Required: true,
			Default:  models.DefaultValue(1)},
		"JobsInQueue": models.IntegerField{
			String:   "Jobs in Queue",
			Compute:  h.Worker().Methods().GetJobsInQueueCount(),
			ReadOnly: true},
		"PauseTime": models.IntegerField{
			Default:  models.DefaultValue(1),
			Required: true},
		"IsRunning": models.BooleanField{
			ReadOnly: true},
		"MaxHistoryEntries": models.IntegerField{
			Default:  models.DefaultValue(500),
			Required: true},
		"MaxHistoryAmmount": models.IntegerField{
			Default:  models.DefaultValue(3),
			Required: true},
		"MaxHistorySelection": models.SelectionField{
			Selection: types.Selection{
				`month`: `Months`,
				`day`:   `Days`,
				`hour`:  `Hours`},
			Default:  models.DefaultValue("days"),
			Required: true},
	})

	h.Worker().Methods().CleanJobHistory().DeclareMethod(
		`CleanJobHistory goes though the job list and removes any entries marked as done and following the Worker rules.
				If worker rules as default: removes all jobs marked as done, older than 3 days or until the total ammount is below 500`,
		func(rs h.WorkerSet) string {
			var deleteLen int
			for _, workerData := range h.Worker().NewSet(rs.Env()).SearchAll().All() {
				query := q.WorkerJobHistory().Status().Equals("done").And().WorkerName().Equals(workerData.Name)
				set := h.WorkerJobHistory().Search(rs.Env(), query).Sorted(func(rs1, rs2 h.WorkerJobHistorySet) bool {
					if rs1.CreateDate().LowerEqual(rs2.CreateDate()) {
						return true
					}
					return false
				})
				rs.PollCancel()
				deltaDur := time.Duration(workerData.MaxHistoryAmmount)
				switch workerData.MaxHistorySelection {
				case "month":
					deltaDur = deltaDur * 30 * 24 * time.Hour
				case "day":
					deltaDur = deltaDur * 24 * time.Hour
				case "hour":
					deltaDur = deltaDur * time.Hour
				}
				setLen := set.Len()
				i := -1
				toDelete := set.Filtered(func(set h.WorkerJobHistorySet) bool {
					i++
					if int64(setLen-i) > workerData.MaxHistoryEntries || set.CreateDate().Add(deltaDur).LowerEqual(dates.Now()) {
						return true
					}
					return false
				})
				deleteLen += toDelete.Len()
				rs.PollCancel()
				toDelete.Unlink()
			}
			return fmt.Sprintf("%d Records Unlinked.", deleteLen)
		})

	h.Worker().Methods().GetJobsInQueueCount().DeclareMethod(
		`GetJobsInQueueCount returns the ammount of jobs currently in a worker's queue`,
		func(rs h.WorkerSet) *h.WorkerData {
			QSize := h.WorkerJobHistory().Search(rs.Env(), q.WorkerJobHistory().WorkerName().Equals(rs.Name()).And().Status().Equals("pending")).Len()
			return &h.WorkerData{JobsInQueue: int64(QSize)}
		})

	h.Worker().Methods().Create().Extend(
		`Create creates and initialize a new Worker`,
		func(set h.WorkerSet, data *h.WorkerData, namer ...models.FieldNamer) h.WorkerSet {
			rs := set.Super().Create(data, namer...)
			rs.StartWorker()
			return rs
		})

	h.Worker().Methods().StartWorker().DeclareMethod(
		`StartWorker starts the goroutine of a specified worker`,
		func(rs h.WorkerSet) {
			models.ExecuteInNewEnvironment(rs.Env().Uid(), func(env models.Environment) {
				w := rs.First()
				go rs.WorkerLoop(&w)
				if _, ok := threadsChanMap[w.Name]; !ok {
					threadsChanMap[w.Name] = make(chan bool, w.MaxThreads)
				}
				for i := 0; i < int(w.MaxThreads); i++ {
					threadsChanMap[w.Name] <- true
				}
			})
		})

	h.Worker().Methods().GetWorker().DeclareMethod(
		`GetWorker returns the worker correspoding to the given name`,
		func(rs h.WorkerSet, str string) h.WorkerSet {
			return h.Worker().Search(rs.Env(), q.Worker().Name().Equals(str))
		})

	h.Worker().Methods().LoadWorkers().DeclareMethod(
		`LoadWorkers reads into database and starts every worker. it also creates the Main worker if it doesn't exist
				Meant to be called only once after server start`,
		func(rs h.WorkerSet) {
			set := h.Worker().Search(rs.Env(), q.Worker().ID().Greater(-1))
			for _, s := range set.Records() {
				s.StartWorker()
			}
			if rs.GetWorker("Main").Len() == 0 {
				rs.Create(&h.WorkerData{
					Name: "Main",
				})
			}
			cancelChans = make(map[string]chan bool)
		})

	h.Worker().Methods().WorkerLoop().DeclareMethod(
		`WorkerLoop is the neverending loop of a worker's goroutine. Should not be directly called. Usage of StartWorker is advised`,
		func(rs h.WorkerSet, w *h.WorkerData) bool {
			for {
				select {
				case <-threadsChanMap[w.Name]:
					var hadJob bool
					models.ExecuteInNewEnvironment(security.SuperUserID, func(env2 models.Environment) {
						set := h.WorkerJobHistory().Search(env2, q.WorkerJobHistory().WorkerName().Equals(w.Name).And().Status().Equals("pending"))
						hadJob = set.Len() > 0
						if hadJob {
							res := set.Sorted(func(rs1, rs2 h.WorkerJobHistorySet) bool {
								if rs1.CreateDate().LowerEqual(rs2.CreateDate()) {
									return true
								}
								return false
							}).Records()[0]
							h.Worker().NewSet(env2).Execute(w, res)
						}
					})
					if !hadJob {
						threadsChanMap[w.Name] <- true
						time.Sleep(time.Duration(w.PauseTime) * time.Second)
					}
				default:
					time.Sleep(time.Duration(w.PauseTime) * time.Second)
				}
			}
		})

	h.Worker().Methods().Execute().DeclareMethod(
		`Execute launches a Worker method specified by the given res`,
		func(rs h.WorkerSet, w *h.WorkerData, res h.WorkerJobHistorySet) {
			writeIn := h.WorkerJobHistoryData{
				StartDate: dates.Now(),
				Status:    "running"}

			if _, ok := models.Registry.Get(res.ModelName()); !ok {
				writeIn.Status = "fail"
				writeIn.MethodOutput = fmt.Sprintf("error: no Model known as '%s'", res.ModelName())
				res.Write(&writeIn)
				threadsChanMap[w.Name] <- true
				return
			}
			rc := res.Env().Pool(res.ModelName())
			method, ok := rc.Model().Methods().Get(res.MethodName())
			if !ok {
				writeIn.Status = "fail"
				writeIn.MethodOutput = fmt.Sprintf("error: no method known as '%s' in model '%s'", res.MethodName(), res.ModelName)
				res.Write(&writeIn)
				threadsChanMap[w.Name] <- true
				return
			}
			res.Write(&writeIn)
			go models.ExecuteInNewEnvironment(security.SuperUserID, func(env2 models.Environment) {
				json.Unmarshal([]byte(res.MethodName()), &method)
				var params interface{}
				json.Unmarshal([]byte(res.ParamsJson()), &params)
				var out []interface{}
				cancelChans[res.TaskUUID()] = make(chan bool)
				err := models.ExecuteInNewEnvironment(security.SuperUserID, func(env3 models.Environment) {
					if params == nil {
						out = method.CallMulti(env3.Pool(res.ModelName()).WithContext("cancelChanId", res.TaskUUID()))
					} else {
						out = method.CallMulti(env3.Pool(res.ModelName()).WithContext("cancelChanId", res.TaskUUID()), interfaceSlice(params)...)
					}
				})
				delete(cancelChans, res.TaskUUID())
				writeOut := h.WorkerJobHistoryData{ReturnDate: dates.Now()}
				if err != nil {
					writeOut.Status = "fail"
					split := strings.Split(err.Error(), "\n----------------------------------\n")
					if split[0] == "ABORT" {
						writeOut.Status = "abort"
					} else {
						writeOut.MethodOutput = fmt.Sprintf("error: %s", split[0])
						writeOut.ExcInfo = split[1]
					}
				} else {
					writeOut.Status = "done"
					outStr := ""
					for _, o := range out {
						outStr += fmt.Sprintf("%v\n", o)
					}
					writeOut.MethodOutput = outStr
				}
				h.WorkerJobHistory().Browse(env2, []int64{res.ID()}).Write(&writeOut)
				threadsChanMap[w.Name] <- true
			})
		})

	h.WorkerJobHistory().DeclareModel()
	h.WorkerJobHistory().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{
			String:   "Method Name",
			Default:  models.DefaultValue("Custom Job"),
			Required: true,
		},
		"WorkerName": models.CharField{
			String:  "Worker",
			Default: models.DefaultValue("Main"),
		},
		"ModelName": models.CharField{
			String:   "Model",
			Required: true,
		},
		"MethodName": models.CharField{
			String:   "Method",
			Required: true,
		},
		"ParamsJson": models.TextField{
			String: "Parameters",
			Help:   "The parameters given to the method. With JSON formating.",
		},
		"MethodOutput": models.TextField{
			String:   "Return Value",
			ReadOnly: true,
		},
		"ExcInfo": models.TextField{
			String:   "Exception Information",
			ReadOnly: true,
		},
		"Status": models.SelectionField{
			String: "Job Status",
			Selection: types.Selection{
				"pending": "Pending",
				"cancel":  "Canceled",
				"running": "Running",
				"abort":   "Aborted",
				"done":    "Done",
				"fail":    "Failed",
			},
			ReadOnly: true,
		},
		"QueuedDate": models.DateTimeField{
			String:   "Queued at",
			ReadOnly: true,
		},
		"StartDate": models.DateTimeField{
			String:   "Started at",
			ReadOnly: true,
		},
		"ReturnDate": models.DateTimeField{
			String:   "finished at",
			ReadOnly: true,
		},
		"TaskUUID": models.CharField{
			ReadOnly: true,
		},
	})
	h.WorkerJobHistory().Fields().CreateDate().SetReadOnly(true)

	h.WorkerJobHistory().Methods().ButtonDone().DeclareMethod(
		`ButtonDone Sets the current rs as done`,
		func(rs h.WorkerJobHistorySet) {
			switch rs.Status() {
			case "":
				panic("Please finish creating the job before trying to mark it as done")
			case "done":
				panic("The Job is already marked as done")
			case "pending":
				panic("You can't mark a pending job as done. please cancel it first")
			default:
				//confirmation box
				rs.SetStatus("done")
			}
		})

	h.WorkerJobHistory().Methods().Requeue().DeclareMethod(
		`Requeue pushes the given rs back to the worker queue`,
		func(rs h.WorkerJobHistorySet) {
			if rs.Status() == "pending" {
				panic("You can't requeue a job that is already queued")
			} else if rs.Status() == "" {
				panic("Please finish creating the job before trying to requeue it")
			}
			rs.Write(&h.WorkerJobHistoryData{
				Status:     "pending",
				QueuedDate: dates.Now(),
				ExcInfo:    "",
			})
		})

	h.WorkerJobHistory().Methods().ButtonCancel().DeclareMethod(
		`ButtonCancel cancels the given job`,
		func(rs h.WorkerJobHistorySet) {
			switch rs.Status() {
			case "pending":
				rs.SetStatus("cancel")
			case "running":
				rs.CancelJob()
			}
		})

	h.WorkerJobHistory().Methods().CancelJob().DeclareMethod(
		`CancelJob marks the given rs to be cancelled. Usage of method h.Worker.PollCancel is required in the worker method`,
		func(rs h.WorkerJobHistorySet) {
			cancelChans[rs.TaskUUID()] <- true
		})

	h.WorkerJobHistory().Methods().Create().Extend(
		`Create creates a new Job entry`,
		func(set h.WorkerJobHistorySet, data *h.WorkerJobHistoryData, namer ...models.FieldNamer) h.WorkerJobHistorySet {
			if data.TaskUUID == "" {
				data.TaskUUID = uuid.New().String()
			}
			if data.WorkerName == "" {
				data.WorkerName = "Main"
			}
			data.Status = "pending"
			data.QueuedDate = dates.Now()
			return set.Super().Create(data, namer...)
		})

	h.JobArgs().DeclareModel()
	h.JobArgs().AddFields(map[string]models.FieldDefinition{
		"WorkerName": models.CharField{},
		"ModelName":  models.CharField{},
		"Methoder":   models.CharField{},
		"Params":     models.CharField{},
	})

	h.JobArgs().Methods().WithParams().DeclareMethod(
		`WithParams gives the current JobArg some parameters. (no parameters by default)`,
		func(rs h.JobArgsSet, params ...interface{}) h.JobArgsSet {
			paramsJson, _ := json.Marshal(params)
			rs.SetParams(string(paramsJson))
			return rs
		})

	h.JobArgs().Methods().WithWorker().DeclareMethod(
		`WithWorker sets the JobsArgs' Worker to a worker corresponding to the given name. ('Main' by default)`,
		func(rs h.JobArgsSet, workerName string) h.JobArgsSet {
			rs.SetWorkerName(workerName)
			return rs
		})

	h.JobArgs().Methods().Enqueue().DeclareMethod(
		`Enqueue creates and pushes to queue the given JobArg created using WithParams and WithWorker`,
		func(rs h.JobArgsSet, method models.Methoder) {
			h.WorkerJobHistory().Create(rs.Env(), &h.WorkerJobHistoryData{
				Name:       method.Underlying().Name(),
				Status:     "pending",
				WorkerName: rs.WorkerName(),
				ModelName:  rs.ModelName(),
				MethodName: method.Underlying().Name(),
				ParamsJson: rs.Params(),
				QueuedDate: dates.Now(),
			})
		})
}

var cancelChans map[string]chan bool

var threadsChanMap map[string]chan bool

func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("interfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}
