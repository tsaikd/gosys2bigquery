package sysinfo

import (
	"context"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/tsaikd/gosys2bigquery/applog"
)

type sysInfoMemory struct {
	Timestamp  time.Time
	Hostname   string
	Total      bqInteger
	Used       bqInteger
	ActualFree bqInteger
	ActualUsed bqInteger
}

func getMemory(ctx context.Context, interval time.Duration, statsChan chan<- sysInfoMemory) {
	ticker := time.NewTicker(interval)
	firstTick := time.NewTimer(0)

	go func() {
		defer ticker.Stop()
		defer firstTick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-firstTick.C:
				if err := sendMemoryInfo(statsChan); err != nil {
					applog.Trace(err)
				}
			case <-ticker.C:
				if err := sendMemoryInfo(statsChan); err != nil {
					applog.Trace(err)
				}
			}
		}
	}()
}

func sendMemoryInfo(statsChan chan<- sysInfoMemory) (err error) {
	mem := sigar.Mem{}
	if err = mem.Get(); err != nil {
		return
	}
	statsChan <- sysInfoMemory{
		Timestamp:  time.Now(),
		Hostname:   hostname,
		Total:      bqInteger(mem.Total),
		Used:       bqInteger(mem.Used),
		ActualFree: bqInteger(mem.ActualFree),
		ActualUsed: bqInteger(mem.ActualUsed),
	}
	return nil
}
