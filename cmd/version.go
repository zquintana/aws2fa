package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	versionCmd = &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Using version:", version)

			release, _, err := getLatestRelease()
			if err != nil {
				return err
			}

			if release != nil {
				fmt.Println("There is a newer version available:", *release.TagName)
			}

			return nil
		},
	}
)
