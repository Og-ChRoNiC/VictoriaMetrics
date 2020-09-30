package common

import (
	"runtime"
	"sync"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
)

// ScheduleUnmarshalWork schedules uw to run in the worker pool.
//
// It is expected that StartUnmarshalWorkers is already called.
func ScheduleUnmarshalWork(uw UnmarshalWork) {
	unmarshalWorkCh <- uw
}

// UnmarshalWork is a unit of unmarshal work.
type UnmarshalWork interface {
	// Unmarshal must implement CPU-bound unmarshal work.
	Unmarshal()
}

// StartUnmarshalWorkers starts unmarshal workers.
func StartUnmarshalWorkers() {
	if unmarshalWorkCh != nil {
		logger.Panicf("BUG: it looks like startUnmarshalWorkers() has been alread called without stopUnmarshalWorkers()")
	}
	gomaxprocs := runtime.GOMAXPROCS(-1)
	unmarshalWorkCh = make(chan UnmarshalWork, 2*gomaxprocs)
	unmarshalWorkersWG.Add(gomaxprocs)
	for i := 0; i < gomaxprocs; i++ {
		go func() {
			defer unmarshalWorkersWG.Done()
			for uw := range unmarshalWorkCh {
				uw.Unmarshal()
			}
		}()
	}
}

// StopUnmarshalWorkers stops unmarshal workers.
//
// No more calles to ScheduleUnmarshalWork are allowed after callsing stopUnmarshalWorkers
func StopUnmarshalWorkers() {
	close(unmarshalWorkCh)
	unmarshalWorkersWG.Wait()
	unmarshalWorkCh = nil
}

var (
	unmarshalWorkCh    chan UnmarshalWork
	unmarshalWorkersWG sync.WaitGroup
)
