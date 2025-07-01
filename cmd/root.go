package cmd

import (
	"github.com/KhashayarKhm/anti-stale/utils"
	"github.com/spf13/cobra"
)

var (
	version = "development"
	rootCmd = &cobra.Command{
		Use:     "anti-stale",
		Short:   "check and find staled issues or pull requests and send comment to un-stale them",
		Version: version,
	}

	configPath   string
	globalConfig = &utils.Config{}
	logger       = &utils.Log{Level: utils.LogLevelInfo}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file (default: anti-stale.json)")
	rootCmd.PersistentFlags().IntVar(&logger.Level, "log-level", 1, "Debug: 0, Info: 1, Warn: 2, Error: 3")
}

func initConfig() {
	globalConfig.MustLoad(logger, configPath)
}
