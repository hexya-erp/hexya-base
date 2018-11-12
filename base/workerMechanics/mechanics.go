package workerMechanics

import (
	"errors"
	"fmt"
	"time"

	"github.com/hexya-erp/hexya/hexya/models"
	"github.com/hexya-erp/hexya/hexya/tools/strutils"
)

func CreateNewWorker(w Worker) *Worker {
	out := &Worker{
		Name:       strutils.MakeUnique(w.Name, workers.names),
		PauseTime:  1000 * time.Millisecond,
		MaxThreads: 0,
		created:    true,
	}
	if w.PauseTime != 0 {
		out.PauseTime = w.PauseTime
	}
	if w.MaxThreads != 0 {
		out.MaxThreads = w.MaxThreads
	}
	out.Threadschan = make(chan bool, out.MaxThreads)
	return out
}

func (w *Worker) Registered() bool {
	return w.registered
}

func EndRegistration(w *Worker) {
	w.registered = true
	if workers.workers == nil {
		workers.workers = make(map[string]*Worker)
	}
	workers.workers[w.Name] = w
	workers.names = append(workers.names, w.Name)
}

func (w *Worker) StartWorker(env models.Environment) error {
	switch {
	case !w.created:
		return errors.New(fmt.Sprintf("worker %s attempted to start without being created.", w.Name))
	case !w.registered:
		return errors.New(fmt.Sprintf("worker %s attempted to start without being registered.", w.Name))
	case w.running:
		return errors.New(fmt.Sprintf("worker %s is already running", w.Name))
	}
	return nil
}

func (w Worker) Get(str string) *Worker {
	for k, v := range workers.workers {
		if k == str {
			return v
		}
	}
	return nil
}
