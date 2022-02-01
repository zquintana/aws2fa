package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	versionCmd = &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Using version:", version)
		},
	}
)
