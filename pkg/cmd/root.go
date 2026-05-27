package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"pm-cli/pkg/config"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:           "pm",
	Short:         "PontoMais CLI",
	Long:          "A small CLI to interact with PontoMais API.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to config file")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	_ = config.Init(cfgFile)
}



