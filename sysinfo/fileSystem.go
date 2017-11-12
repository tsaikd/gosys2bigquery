package sysinfo

import (
	"context"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/tsaikd/gosys2bigquery/applog"
)

type fileSystemUsage struct {
	Timestamp time.Time
	Hostname  string
	Path      string
	Total     bqInteger
	Used      bqInteger
}

func getFileSystemUsage(ctx context.Context, interval time.Duration, path string, statsChan chan<- fileSystemUsage) {
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
				if err := sendFileSystemUsage(statsChan, path); err != nil {
					applog.Trace(err)
				}
			case <-ticker.C:
				if err := sendFileSystemUsage(statsChan, path); err != nil {
					applog.Trace(err)
				}
			}
		}
	}()
}

func sendFileSystemUsage(statsChan chan<- fileSystemUsage, path string) (err error) {
	fsUsage := sigar.FileSystemUsage{}
	if err = fsUsage.Get(path); err != nil {
		return
	}
	statsChan <- fileSystemUsage{
		Timestamp: time.Now(),
		Hostname:  hostname,
		Path:      path,
		Total:     bqInteger(fsUsage.Total << 10),
		Used:      bqInteger(fsUsage.Used << 10),
	}
	return nil
}
