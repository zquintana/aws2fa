package cmd

import (
	"fmt"
	v "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"strings"
)

var (
	version    = "0.0+dev"
	versionCmd = &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Using version:", version)

			release, _, err := getLatestRelease()
			if err != nil {
				return err
			}

			if isNewerVersion(getCurrentVersion(), *release.TagName) {
				fmt.Println("There is a newer version available:", *release.TagName)
			}

			return nil
		},
	}
)

func getCurrentVersion() string {
	return parseVersion(version)
}

func parseVersion(v string) string {
	i := strings.IndexByte(v, '-')
	if i > 0 {
		return v[:i]
	}

	return v
}

func isNewerVersion(currentVersion string, tagName string) bool {
	c, err := v.NewVersion(currentVersion)
	if err != nil {
		panic(err)
	}

	t, err := v.NewVersion(tagName)
	if err != nil {
		panic(err)
	}

	return t.GreaterThan(c)
}
