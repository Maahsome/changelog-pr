package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"changelog-pr/common"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	gitProvider string
	ghToken     string
	semVer      string
	gitCommit   string
	buildDate   string
	gitRef      string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "changelog-pr",
	Short: "Generate a changelog from PR descriptions",
	Long: `Given a previous git TAG, locate all of the PRs since that TAG, and parse the
	description of the PR for specific MD sections and build a changelog from the data.

	Currently there is only support for GitHub repositories, though adding different git
	providers should be fairly straight forward.

	Use the 'changelog-pr template' command to display the PR TEMPLATE data`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logFile, _ := cmd.Flags().GetString("log-file")
		logLevel, _ := cmd.Flags().GetString("log-level")
		ll := "Warning"
		switch strings.ToLower(logLevel) {
		case "trace":
			ll = "Trace"
		case "debug":
			ll = "Debug"
		case "info":
			ll = "Info"
		case "warning":
			ll = "Warning"
		case "error":
			ll = "Error"
		case "fatal":
			ll = "Fatal"
		}

		common.NewLogger(ll, logFile)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.changelog-pr.yaml)")
	rootCmd.PersistentFlags().StringVarP(&gitProvider, "git-provider", "g", "github", "git source provider (github)")
	rootCmd.PersistentFlags().StringP("log-file", "l", "", "Specify a log file to log events to, default to no logging")
	rootCmd.PersistentFlags().StringP("log-level", "v", "", "Specify a log level for logging, default to Warning (Trace, Debug, Info, Warning, Error, Fatal)")
	rootCmd.PersistentFlags().StringVar(&ghToken, "gh-token", "", "Specify your GitHub personal access token")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".changelog-pr" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".changelog-pr")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
