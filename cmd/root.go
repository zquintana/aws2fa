package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"os"
	"path"
)

var (
	profile         string
	durationSeconds int32
	verbose         bool
	rootCmd         = &cobra.Command{
		Use: "aws2fa",
	}
)

func readAwsCredentials() (*ini.File, error) {
	return readAwsConfigFile("credentials")
}

func readAwsConfig() (*ini.File, error) {
	return readAwsConfigFile("config")
}

func getAwsConfigFilePath(name string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".aws", name), nil
}

func readAwsConfigFile(name string) (*ini.File, error) {
	configPath, err := getAwsConfigFilePath(name)
	if err != nil {
		return nil, err
	}

	return ini.Load(configPath)
}

func getCurrentProfileName() string {
	if len(profile) < 1 {
		return "default"
	} else {
		return profile
	}
}

func log(a ...interface{}) {
	if verbose != true {
		return
	}

	fmt.Println(a...)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "P", "", "AWS Profile")
	rootCmd.PersistentFlags().Int32VarP(&durationSeconds, "duration-seconds", "D", int32(24*60*60), "Duration in seconds")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "")

	rootCmd.AddCommand(mfaAuthCmd)
	rootCmd.AddCommand(opCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(selfUpdateCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
