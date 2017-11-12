package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromYAML(t *testing.T) {
	require := require.New(t)
	require.NotNil(require)

	conf, err := LoadFromYAML([]byte(strings.TrimSpace(`
interval: 5
bigquery:
  mock: true
  keyfile: gckeyfile.json
  projectid: example-project-2017
  datasetid: sys_log
memory:
  enable: true
filesystem:
  paths:
    - "/"
    - "/tmp"
docker:
  endpoint: "unix:///var/run/docker.sock"
  keepcontainernameprefixslash: true
  stats:
    enable: true
	`)))
	require.NoError(err)

	require.EqualValues(5, conf.Interval)
	require.True(conf.BigQuery.Mock)
	require.Equal("gckeyfile.json", conf.BigQuery.KeyFile)
	require.Equal("example-project-2017", conf.BigQuery.ProjectID)
	require.Equal("sys_log", conf.BigQuery.DatasetID)
	require.True(conf.Memory.Enable)
	require.Len(conf.FileSystem.Paths, 2)
	require.Equal("unix:///var/run/docker.sock", conf.Docker.EndPoint)
	require.True(conf.Docker.KeepContainerNamePrefixSlash)
	require.True(conf.Docker.Stats.Enable)
}
