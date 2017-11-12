package sysinfo

import (
	"context"
	"os"
	"time"

	"github.com/tsaikd/KDGoLib/bqutil"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/gosys2bigquery/applog"
	"github.com/tsaikd/gosys2bigquery/config"
	"go.uber.org/zap"
	"google.golang.org/api/googleapi"
)

// errors
var (
	ErrInvalidConfig2 = errutil.NewFactory("Invalid config %q: %+v")
)

type bqInteger int64

var hostname = func() string {
	hn, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hn
}()

// StartLoop start system stats collection loop
func StartLoop(ctx context.Context, bqclient *bqutil.Client, conf config.Config) (err error) {
	interval := time.Duration(conf.Interval) * time.Second
	if interval < time.Second {
		interval = time.Second
	}
	applog.Logger().Info("config", zap.Duration("interval", interval))

	mock := conf.BigQuery.Mock
	applog.Logger().Info("config", zap.Bool("bigquery.mock", mock))
	uploader := bqclient.UploadAsync
	if mock {
		uploader = mockUploadAsync
	}

	datasetName := conf.BigQuery.DatasetID
	if datasetName == "" {
		datasetName = "sys_log"
	}
	if !mock {
		if _, err = bqclient.EnsureDataset(datasetName); err != nil {
			if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
				applog.Logger().Error("Is BigQuery projectID correct?", zap.String("bigquery.projectid", conf.BigQuery.ProjectID))
			}
			return
		}
	}
	applog.Logger().Info("config", zap.String("bigquery.datasetid", datasetName))

	memoryTable := "memory"
	memChan := make(chan sysInfoMemory, 10)
	if conf.Memory.Enable {
		getMemory(ctx, interval, memChan)
	}
	applog.Logger().Info("config", zap.Bool("memory.enable", conf.Memory.Enable))

	fsTable := "filesystem"
	fsChan := make(chan fileSystemUsage, len(conf.FileSystem.Paths)*3+1)
	for _, path := range conf.FileSystem.Paths {
		getFileSystemUsage(ctx, interval, path, fsChan)
		applog.Logger().Info("config", zap.String("filesystem.paths", path))
	}

	dockerStatsTable := "docker_stats"
	dockerStatsChan := make(chan sysInfoDockerContainer, 100)
	if conf.Docker.Stats.Enable {
		if err = getDockerStats(ctx, interval, dockerStatsChan, conf.Docker.EndPoint); err != nil {
			return
		}
	}
	applog.Logger().Info("config", zap.Bool("docker.stats.enable", conf.Docker.Stats.Enable))

	for {
		select {
		case <-ctx.Done():
			return
		case stats := <-memChan:
			if err = uploader(datasetName, memoryTable, stats); err != nil {
				return
			}
		case stats := <-fsChan:
			if err = uploader(datasetName, fsTable, stats); err != nil {
				return
			}
		case stats := <-dockerStatsChan:
			if err = uploader(datasetName, dockerStatsTable, stats); err != nil {
				return
			}
		}
	}
}

func mockUploadAsync(datasetName string, tableName string, data interface{}) (err error) {
	applog.Logger().Sugar().Infof("dataset: %q, table: %q, data: %+v", datasetName, tableName, data)
	return nil
}
