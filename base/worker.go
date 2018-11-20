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
		"PauseTime": models.IntegerField{},
		"IsRunning": models.BooleanField{
			ReadOnly: true},
		"MaxHistoryEntries": models.IntegerField{
			Default:  models.DefaultValue(200),
			Required: true},
		"MaxHistoryAmmount": models.IntegerField{
			Default:  models.DefaultValue(1),
			Required: true},
		"MaxHistorySelection": models.SelectionField{
			Selection: types.Selection{
				`month`: `Months`,
				`day`:   `Days`,
				`hour`:  `Hours`},
			Default:  models.DefaultValue("hour"),
			Required: true},
	})

	//will be removed after testings
	h.Worker().Methods().Println().DeclareMethod(
		``,
		func(rs h.WorkerSet, args ...interface{}) {
			fmt.Println(args...)
		})

	//will be made
	h.Worker().Methods().CLeanJobHistory().DeclareMethod(
		``,
		func(rs h.WorkerSet) {
			//h.WorkerJobHistory().Search(rs.Env(), q.WorkerJobHistory().Status().Equals("done"))
		})

	h.Worker().Methods().GetJobsInQueueCount().DeclareMethod(
		`returns the ammount of jobs currently in worker queue`,
		func(rs h.WorkerSet) *h.WorkerData {
			QSize := h.WorkerJob().Search(rs.Env(), q.WorkerJob().ParentWorkerName().Equals(rs.Name())).Len()
			return &h.WorkerData{JobsInQueue: int64(QSize)}
		})

	h.Worker().Methods().Create().Extend(
		``,
		func(set h.WorkerSet, data *h.WorkerData, namer ...models.FieldNamer) h.WorkerSet {
			rs := set.Super().Create(data, namer...)
			rs.StartWorker(data)
			return rs
		})

	h.Worker().Methods().StartWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, w *h.WorkerData) {
			models.ExecuteInNewEnvironment(rs.Env().Uid(), func(env models.Environment) {
				go rs.WorkerLoop(w)
				if _, ok := threadsChanMap[w.Name]; !ok {
					threadsChanMap[w.Name] = make(chan bool, w.MaxThreads)
				}
				for i := 0; i < int(w.MaxThreads); i++ {
					threadsChanMap[w.Name] <- true
				}
			})
		})

	h.Worker().Methods().GetWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, str string) *h.WorkerData {
			set := h.Worker().Search(rs.Env(), q.Worker().Name().Equals(str))
			if set.Len() == 0 {
				return nil
			}
			data := set.First()
			return &data
		})

	h.Worker().Methods().LoadWorkers().DeclareMethod(
		``,
		func(rs h.WorkerSet) {
			set := h.Worker().Search(rs.Env(), q.Worker().ID().Greater(-1))
			for _, s := range set.All() {
				rs.StartWorker(s)
			}
			if rs.GetWorker("Main") == nil {
				rs.Create(&h.WorkerData{
					Name: "Main",
				})
			}
		})

	h.Worker().Methods().WorkerLoop().DeclareMethod(
		``,
		func(rs h.WorkerSet, w *h.WorkerData) bool {
			for {
				select {
				case <-threadsChanMap[w.Name]:
					var hadJob bool
					models.ExecuteInNewEnvironment(security.SuperUserID, func(env2 models.Environment) {
						set := h.WorkerJob().Search(env2, q.WorkerJob().ParentWorkerName().Equals(w.Name))
						hadJob = set.Len() > 0
						if hadJob {
							res := set.Sorted(func(rs1, rs2 h.WorkerJobSet) bool {
								if rs1.CreateDate().LowerEqual(rs2.CreateDate()) {
									return true
								}
								return false
							}).All()[0]
							set.Browse([]int64{res.ID}).Unlink()
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
		``,
		func(rs h.WorkerSet, w *h.WorkerData, res *h.WorkerJobData) {
			models.ExecuteInNewEnvironment(security.SuperUserID, func(env3 models.Environment) {
				historyEntry := h.WorkerJobHistory().Search(env3, q.WorkerJobHistory().TaskUUID().Equals(res.TaskUUID))
				historyEntry.SetStatus("running")
				historyEntry.SetStartDate(dates.Now())
			})
			go models.ExecuteInNewEnvironment(security.SuperUserID, func(env2 models.Environment) {
				historyEntry := h.WorkerJobHistory().Search(env2, q.WorkerJobHistory().TaskUUID().Equals(res.TaskUUID))
				if _, ok := models.Registry.Get(res.ModelName); !ok {
					historyEntry.SetStatus("fail")
					historyEntry.SetMethodOutput(fmt.Sprintf("error: no Model known as '%s'", res.ModelName))
					threadsChanMap[w.Name] <- true
					return
				}
				rc := env2.Pool(res.ModelName)
				method, ok := rc.Model().Methods().Get(res.Method)
				if !ok {
					historyEntry.SetStatus("fail")
					historyEntry.SetMethodOutput(fmt.Sprintf("error: no method known as '%s' in model '%s'", res.Method, res.ModelName))
					threadsChanMap[w.Name] <- true
					return
				}
				json.Unmarshal([]byte(res.Method), &method)
				var params interface{}
				json.Unmarshal([]byte(res.ParamsJson), &params)
				var out []interface{}
				err := models.ExecuteInNewEnvironment(security.SuperUserID, func(env3 models.Environment) {
					if params == nil {
						out = method.CallMulti(env3.Pool(res.ModelName))
					} else {
						out = method.CallMulti(env3.Pool(res.ModelName), interfaceSlice(params)...)
					}
				})
				historyEntry.SetReturnDate(dates.Now())
				if err != nil {
					historyEntry.SetStatus("fail")
					split := strings.Split(err.Error(), "\n----------------------------------\n")
					historyEntry.SetMethodOutput(fmt.Sprintf("error: %s", split[0]))
					historyEntry.SetExcInfo(split[1])
				} else {
					historyEntry.SetStatus("done")
					outStr := ""
					for _, o := range out {
						outStr += fmt.Sprintf("%v\n", o)
					}
					historyEntry.SetMethodOutput(outStr)
				}
				threadsChanMap[w.Name] <- true
			})
		})

	h.WorkerJob().DeclareModel()
	h.WorkerJob().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{
			String: "name",
		},
		"Method":           models.CharField{},
		"ModelName":        models.CharField{},
		"ParamsJson":       models.CharField{},
		"ParentWorkerName": models.CharField{},
		"TaskUUID":         models.CharField{},
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
		``,
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
		``,
		func(rs h.WorkerJobHistorySet) {
			if rs.Status() == "pending" {
				panic("You can't requeue a job that is already queued")
			} else if rs.Status() == "" {
				panic("Please finish creating the job before trying to requeue it")
			}
			h.WorkerJob().Create(rs.Env(), &h.WorkerJobData{
				Name:             rs.Name(),
				Method:           rs.MethodName(),
				ModelName:        rs.ModelName(),
				ParamsJson:       rs.ParamsJson(),
				ParentWorkerName: rs.WorkerName(),
				TaskUUID:         rs.TaskUUID(),
			})
			rs.SetStatus("pending")
			rs.SetQueuedDate(dates.Now())
			rs.SetExcInfo("")
		})

	h.WorkerJobHistory().Methods().Create().Extend(
		``,
		func(set h.WorkerJobHistorySet, data *h.WorkerJobHistoryData, namer ...models.FieldNamer) h.WorkerJobHistorySet {
			if data.TaskUUID == "" {
				data.TaskUUID = uuid.New().String()
			}
			if data.WorkerName == "" {
				data.WorkerName = "Main"
			}
			h.WorkerJob().Create(set.Env(), &h.WorkerJobData{
				Name:             data.Name,
				Method:           data.MethodName,
				ModelName:        data.ModelName,
				ParamsJson:       data.ParamsJson,
				ParentWorkerName: data.WorkerName,
				TaskUUID:         data.TaskUUID,
			})
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
		``,
		func(rs h.JobArgsSet, params ...interface{}) h.JobArgsSet {
			paramsJson, _ := json.Marshal(params)
			rs.SetParams(string(paramsJson))
			return rs
		})

	h.JobArgs().Methods().WithWorker().DeclareMethod(
		``,
		func(rs h.JobArgsSet, workerName string) h.JobArgsSet {
			rs.SetWorkerName(workerName)
			return rs
		})

	h.JobArgs().Methods().Enqueue().DeclareMethod(
		``,
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
