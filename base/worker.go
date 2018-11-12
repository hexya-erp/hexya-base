package base

import (
	"fmt"
	"time"

	"errors"
	"reflect"

	"encoding/json"

	"strconv"

	"strings"

	"github.com/hexya-erp/hexya-base/base/workerMechanics"
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
			String: "Worker",
			Unique: true},
		"MaxThreads": models.IntegerField{
			String: "Threads"},
		"JobsInQueue": models.IntegerField{
			String:  "Jobs in Queue",
			Compute: h.Worker().Methods().GetJobsInQueueCount()},
		"PauseTime": models.IntegerField{},
		"IsRunning": models.BooleanField{},
	})

	h.Worker().Methods().DummyFunc().DeclareMethod(
		``,
		func(rs h.WorkerSet, str string) {
			fmt.Println("lel ", str)
			time.Sleep(2 * time.Second)
		})

	h.Worker().Methods().DummyFuncPanic().DeclareMethod(
		``,
		func(rs h.WorkerSet) {
			panic("PAAAAAAANIC")
		})

	h.Worker().Methods().DummyFuncSingleReturn().DeclareMethod(
		``,
		func(rs h.WorkerSet) string {
			return "hoy"
		})

	h.Worker().Methods().DummyFuncMult().DeclareMethod(
		``,
		func(rs h.WorkerSet, a, b float64) (float64, float64, float64) {
			return a, b, float64(a * b)
		})

	h.Worker().Methods().GetJobsInQueueCount().DeclareMethod(
		`returns the ammount of jobs currently in worker queue`,
		func(rs h.WorkerSet) *h.WorkerData {
			QSize := h.WorkerJob().Search(rs.Env(), q.WorkerJob().ParentWorkerName().Equals(rs.Name())).Len()
			if QSize == 0 {
				for i := 0; i < 20; i++ {
					switch {
					case i%3 == 0:
						rs.Enqueue(h.Worker().Methods().DummyFuncSingleReturn())
					case i%5 == 0:
						rs.Enqueue(h.Worker().Methods().DummyFuncPanic())
					case i%7 == 0:
						rs.WithParams(i, i+2).Enqueue(h.Worker().Methods().DummyFuncMult())
					default:
						rs.WithWorker("Main").WithParams(strconv.Itoa(i)).Enqueue(h.Worker().Methods().DummyFunc())
					}
				}
			}
			return &h.WorkerData{JobsInQueue: int64(QSize)}
		})

	h.Worker().Methods().CreateAndRegisterNewWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, wo workerMechanics.Worker) *workerMechanics.Worker {
			w := rs.CreateNewWorker(wo)
			rs.RegisterWorker(w)
			return w
		})

	h.Worker().Methods().CreateRegisterStartNewWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, wo workerMechanics.Worker) *workerMechanics.Worker {
			w := rs.CreateAndRegisterNewWorker(wo)
			rs.StartWorker(w)
			return w
		})

	h.Worker().Methods().CreateNewWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, w workerMechanics.Worker) *workerMechanics.Worker {
			var out *workerMechanics.Worker
			models.ExecuteInNewEnvironment(rs.Env().Uid(), func(env models.Environment) {
				out = workerMechanics.CreateNewWorker(w)
			})
			return out
		})

	h.Worker().Methods().RegisterWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, w *workerMechanics.Worker) {
			models.ExecuteInNewEnvironment(rs.Env().Uid(), func(env models.Environment) {
				err := registerWorker(w, rs)
				if err != nil {
					log.Error(fmt.Sprintf("%v", err))
				}
			})
		})

	h.Worker().Methods().StartWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, w *workerMechanics.Worker) {
			models.ExecuteInNewEnvironment(rs.Env().Uid(), func(env models.Environment) {
				err := w.StartWorker(rs.Env())
				if err != nil {
					log.Error(fmt.Sprintf("%v", err))
				} else {
					go rs.WorkerLoop(w)
					for i := 0; i < w.MaxThreads; i++ {
						w.Threadschan <- true
					}
				}
			})
		})

	h.Worker().Methods().GetWorker().DeclareMethod(
		``,
		func(rs h.WorkerSet, str string) *workerMechanics.Worker {
			return workerMechanics.Worker{}.Get(str)
		})

	h.Worker().Methods().LoadWorkers().DeclareMethod(
		``,
		func(rs h.WorkerSet) {
			set := h.Worker().Search(rs.Env(), q.Worker().ID().Greater(-1))
			for _, s := range set.All() {
				neww := workerMechanics.CreateNewWorker(workerMechanics.Worker{
					Name:       s.Name,
					PauseTime:  time.Duration(s.PauseTime) * time.Second,
					MaxThreads: int(s.MaxThreads),
				})
				workerMechanics.EndRegistration(neww)
				go rs.WorkerLoop(neww)
				for i := 0; i < neww.MaxThreads; i++ {
					neww.Threadschan <- true
				}
			}
			if rs.GetWorker("Main") == nil {
				rs.CreateRegisterStartNewWorker(workerMechanics.Worker{
					Name:       "Main",
					PauseTime:  1 * time.Second,
					MaxThreads: 1,
				})
				rs.WithWorker("Main").WithParams("hey").Enqueue(h.Worker().Methods().DummyFunc())

			}
		})

	h.Worker().Methods().WorkerLoop().DeclareMethod(
		``,
		func(rs h.WorkerSet, w *workerMechanics.Worker) bool {
			for {
				select {
				case <-w.Threadschan:
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
						w.Threadschan <- true
						time.Sleep(w.PauseTime)
					}
				default:
					time.Sleep(w.PauseTime)
				}
			}
		})

	h.Worker().Methods().Execute().DeclareMethod(
		``,
		func(rs h.WorkerSet, w *workerMechanics.Worker, res *h.WorkerJobData) {
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
					w.Threadschan <- true
					return
				}
				rc := env2.Pool(res.ModelName)
				method, ok := rc.Model().Methods().Get(res.Method)
				if !ok {
					historyEntry.SetStatus("fail")
					historyEntry.SetMethodOutput(fmt.Sprintf("error: no method known as '%s' in model '%s'", res.Method, res.ModelName))
					w.Threadschan <- true
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
				w.Threadschan <- true
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

func registerWorker(w *workerMechanics.Worker, rs h.WorkerSet) error {
	if w.Registered() == false {
		if rs.Search(q.Worker().Name().Equals(w.Name)).Len() > 0 {
			return errors.New(fmt.Sprintf(`Could not register worker "%s". another worker with this name is already registered`, w.Name))
		}
		if w.PauseTime.Seconds() < 1 {
			w.PauseTime = 1 * time.Second
		}
		rs.Create(&h.WorkerData{
			Name:       w.Name,
			MaxThreads: int64(w.MaxThreads),
			PauseTime:  int64(w.PauseTime / time.Second),
		})
		workerMechanics.EndRegistration(w)
	}
	return nil
}

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
