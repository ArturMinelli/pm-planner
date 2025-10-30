package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ()

// rootCmd is the base command for pm-cli
var rootCmd = &cobra.Command{
	Use:   "pm",
	Short: "PontoMais CLI",
	Long:  "A small CLI to interact with PontoMais API.",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Config via Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".pm"))
	}

	// Read config if present; ignore if not found
	_ = viper.ReadInConfig()
}



