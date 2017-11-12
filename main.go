package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/bqutil"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
	"github.com/tsaikd/gosys2bigquery/applog"
	"github.com/tsaikd/gosys2bigquery/config"
	"github.com/tsaikd/gosys2bigquery/sysinfo"
)

// command line flags
var (
	flagConfig = &cobrather.StringFlag{
		Name:    "config",
		Default: "config.yml",
		Usage:   "Path to configuration file",
		EnvVar:  "GSB_CONFIG",
	}
)

var bqclient *bqutil.Client

// Module info
var Module = &cobrather.Module{
	Use:   "gosys2bigquery",
	Short: "Log system stats to BigQuery",
	Dependencies: []*cobrather.Module{
		applog.Module,
	},
	Commands: []*cobrather.Module{
		cobrather.VersionModule,
	},
	Flags: []cobrather.Flag{
		flagConfig,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) error {
		conf, err := config.LoadFromFile(flagConfig.String())
		if err != nil {
			return err
		}
		gckeyfile := conf.BigQuery.KeyFile
		if gckeyfile == "" {
			gckeyfile = "gckeyfile.json"
		}
		bqclient, err = bqutil.NewClient(
			ctx,
			gckeyfile,
			conf.BigQuery.ProjectID,
			bqutil.OptionEnsureDatasetWhenFirstUpload(false),
		)
		if err != nil {
			return err
		}
		return sysinfo.StartLoop(ctx, bqclient, conf)
	},
	PostRunE: func(ctx context.Context, cmd *cobra.Command, args []string) error {
		if bqclient != nil {
			return bqclient.Close()
		}
		return nil
	},
}

func main() {
	Module.MustMainRun()
}
