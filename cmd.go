package main

import (
	"github.com/spf13/cobra"
)

var (
	_configFile string
	_config     Config
)

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "./github-statistic [flags]",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			if _config.GithubToken == "" {
				panic("github token is required")
			}
			genReport(_config)
		},
	}

	// config file
	cmd.PersistentFlags().StringVar(&_configFile, "config-file", "config.yaml", "configuration file")

	// read config from yaml file
	config, err := NewFromFile(_configFile)
	if err != nil {
		panic(err)
	}
	_config = *config

	// config items
	cmd.PersistentFlags().StringVar(&_config.StartTime, "start-time", "", "start time of statistics")
	cmd.PersistentFlags().StringVar(&_config.EndTime, "end-time", "", "end time of statistics")
	cmd.PersistentFlags().IntVar(&_config.LastDays, "last-days", 14, "the past days to statistic")
	cmd.PersistentFlags().BoolVar(&_config.ToCurrent, "to-current", false, "whether to statistic to current time, otherwise to 23:59:59 of the previous day")
	cmd.PersistentFlags().StringArrayVar(&_config.Repositories, "repositories", config.Repositories, "repositories to statistic")

	return cmd
}

func main() {
	cmd := newCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
