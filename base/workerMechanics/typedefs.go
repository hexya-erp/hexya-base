package workerMechanics

import (
	"time"

	"github.com/hexya-erp/hexya/hexya/models"
)

type JobPreArgs struct {
	WorkerName string
	RecCol     *models.RecordCollection
	Params     []interface{}
}

var workers workerList

type Worker struct {
	Name        string
	PauseTime   time.Duration
	MaxThreads  int
	Threadschan chan bool
	created     bool
	registered  bool
	running     bool
}

type workerList struct {
	workers map[string]*Worker
	names   []string
}
