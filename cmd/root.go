package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"changelog-pr/common"

	"github.com/spf13/viper"
)

var (
	cfgFile     string
	gitProvider string
	ghToken     string
	ghHost      string
	glToken     string
	glHost      string
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

		if os.Args[1] != "version" {
			if len(gitProvider) > 0 {
				viper.Set("gitprovider", gitProvider)
				verr := viper.WriteConfig()
				if verr != nil {
					logrus.WithError(verr).Info("Failed to write config")
				}
			} else {
				gitProviderFromConfig := viper.GetString("gitprovider")
				if len(gitProviderFromConfig) > 0 {
					gitProvider = gitProviderFromConfig
				} else {
					gitProvider = "gitlab"
				}
			}

			if len(glHost) > 0 {
				viper.Set("gitlabhost", glHost)
				verr := viper.WriteConfig()
				if verr != nil {
					logrus.WithError(verr).Info("Failed to write config")
				}
			} else {
				glHostFromConfig := viper.GetString("gitlabhost")
				if len(glHostFromConfig) > 0 {
					glHost = glHostFromConfig
				} else {
					glHost = "gitlab.com"
				}
			}

			if len(glToken) > 0 {
				viper.Set("gitlabtoken", glToken)
				verr := viper.WriteConfig()
				if verr != nil {
					logrus.WithError(verr).Info("Failed to write config")
				}
			} else {
				glTokenFromConfig := viper.GetString("gitlabtoken")
				if len(glTokenFromConfig) > 0 {
					glToken = glTokenFromConfig
				} else {
					glTokenFromEnv := os.Getenv("GITLAB_TOKEN")
					if len(glTokenFromEnv) > 0 {
						glToken = glTokenFromEnv
					} else {
						if gitProvider == "gitlab" {
							logrus.Fatal("Please provide a gitlab-token via --gitlab-token or GITLAB_TOKEN environment variable")
						}
					}
				}
			}

			if len(ghHost) > 0 {
				viper.Set("githubhost", ghHost)
				verr := viper.WriteConfig()
				if verr != nil {
					logrus.WithError(verr).Info("Failed to write config")
				}
			} else {
				ghHostFromConfig := viper.GetString("githubhost")
				if len(ghHostFromConfig) > 0 {
					ghHost = ghHostFromConfig
				} else {
					ghHost = "github.com"
				}
			}

			if len(ghToken) > 0 {
				viper.Set("githubtoken", ghToken)
				verr := viper.WriteConfig()
				if verr != nil {
					logrus.WithError(verr).Info("Failed to write config")
				}
			} else {
				ghTokenFromConfig := viper.GetString("githubtoken")
				if len(ghTokenFromConfig) > 0 {
					ghToken = ghTokenFromConfig
				} else {
					ghTokenFromEnv := os.Getenv("GITHUB_TOKEN")
					if len(ghTokenFromEnv) > 0 {
						ghToken = ghTokenFromEnv
					} else {
						if gitProvider == "github" {
							logrus.Fatal("Please provide a github-token via --github-token or GITHUB_TOKEN environment variable")
						}
					}
				}
			}
		}
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
	rootCmd.PersistentFlags().StringVarP(&gitProvider, "git-provider", "g", "", "git source provider (github)")
	rootCmd.PersistentFlags().StringP("log-file", "l", "", "Specify a log file to log events to, default to no logging")
	rootCmd.PersistentFlags().StringP("log-level", "v", "", "Specify a log level for logging, default to Warning (Trace, Debug, Info, Warning, Error, Fatal)")
	rootCmd.PersistentFlags().StringVar(&ghToken, "github-token", "", "Specify your GitHub personal access token")
	rootCmd.PersistentFlags().StringVar(&ghHost, "github-host", "", "Specify your GitHub Host")
	rootCmd.PersistentFlags().StringVar(&glToken, "gitlab-token", "", "Specify your GitLab personal access token")
	rootCmd.PersistentFlags().StringVar(&glHost, "gitlab-host", "", "Specify your GitLab Host")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		workDir := fmt.Sprintf("%s/.config/changelog-pr", home)
		if _, err := os.Stat(workDir); err != nil {
			if os.IsNotExist(err) {
				mkerr := os.MkdirAll(workDir, os.ModePerm)
				if mkerr != nil {
					common.Logger.Fatal("Error creating ~/.config/changelog-pr directory", mkerr)
				}
			}
		}
		if stat, err := os.Stat(workDir); err == nil && stat.IsDir() {
			configFile := fmt.Sprintf("%s/%s", workDir, "config.yaml")
			createRestrictedConfigFile(configFile)
			viper.SetConfigFile(configFile)
		} else {
			common.Logger.Info("The ~/.config/changelog-pr path is a file and not a directory, please remove the 'changelog-pr' file.")
			os.Exit(1)
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		common.Logger.Warn("Failed to read viper config file.")
	}
}

func createRestrictedConfigFile(fileName string) {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			file, ferr := os.Create(fileName)
			if ferr != nil {
				common.Logger.Info("Unable to create the configfile.")
				os.Exit(1)
			}
			mode := int(0600)
			if cherr := file.Chmod(os.FileMode(mode)); cherr != nil {
				common.Logger.Info("Chmod for config file failed, please set the mode to 0600.")
			}
		}
	}
}
