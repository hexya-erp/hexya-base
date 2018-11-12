package base

import (
	"fmt"

	"github.com/hexya-erp/hexya/hexya/models"

	"time"

	"github.com/hexya-erp/hexya/pool/h"
)

var schedulerUpdateChan chan bool

func init() {
	h.Cron().DeclareModel()
	h.Cron().AddFields(map[string]models.FieldDefinition{
		"Name": models.CharField{
			String: "Cron",
			Unique: true},
		"TargetWorker": models.CharField{
			String:  "Worker",
			Default: models.DefaultValue("Main")},
		"TargetModel": models.CharField{
			String:   "Model",
			Required: true},
		"TargetMethod": models.CharField{
			String:   "Method",
			Required: true},
		"ModelMethodStr": models.CharField{
			String:  "Method",
			Compute: h.Cron().Methods().ComputeModelMethodStr()},
	})

	h.Cron().Methods().ComputeModelMethodStr().DeclareMethod(
		``,
		func(rs h.CronSet) h.CronData {
			return h.CronData{ModelMethodStr: fmt.Sprintf("%s - %s", rs.TargetModel(), rs.TargetMethod())}
		})

	h.Cron().Methods().StartScheduler().DeclareMethod(
		``,
		func(rs h.CronSet) {
			schedulerUpdateChan = make(chan bool)
			go rs.SchedulerLoop(15 * time.Minute)
		})

	h.Cron().Methods().SchedulerLoop().DeclareMethod(
		``,
		func(rs h.CronSet, next time.Duration) {
			for {
				select {
				case <-time.After(next):
					fmt.Println("time trigger")
				case <-schedulerUpdateChan:
					fmt.Println("update")
				}
			}
		})
}
