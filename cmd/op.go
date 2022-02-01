package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
)

var opCmd = &cobra.Command{
	Use:   "op",
	Short: "Authenticate with 1Password managed token",
	Long: "Requires valid 1Password session to work. Tag your 1Password entries with " +
		"\"aws-profile: PROFILE_NAME\" to automatically retrieve token for authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
		if err != nil {
			return err
		}

		data, err := retrieveItems()
		if err != nil {
			return err
		}

		item, err := searchForProfileEntry(data)
		if err != nil {
			return err
		}

		totp, err := getTotp(item)
		if err != nil {
			return err
		}

		return perform2faLogin(cfg, totp)
	},
}

type Overview struct {
	Tags []string `json:"tags"`
}

type SimpleItem struct {
	Uuid      string   `json:"uuid"`
	VaultUuid string   `json:"vaultUuid"`
	Overview  Overview `json:"overview"`
}

func retrieveItems() ([]byte, error) {
	cmd := exec.Command("op", "list", "items")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, errors.New("unable to list items, make sure 1password cli is installed and you are signed in")
	}

	return stdout, nil
}

func getTotp(item *SimpleItem) (string, error) {
	cmd := exec.Command("op", "get", "totp", item.Uuid)
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdout)), nil
}

func searchForProfileEntry(itemsData []byte) (*SimpleItem, error) {
	var items []interface{}
	if err := json.Unmarshal(itemsData, &items); err != nil {
		return nil, err
	}

	query, err := gojq.Parse(fmt.Sprintf(".[] | select(.overview.tags | length > 0) | select(.overview.tags[] == \"aws-profile: %s\")", getCurrentProfileName()))
	if err != nil {
		return nil, err
	}

	iter := query.Run(items)
	if item, ok := iter.Next(); ok {
		b, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		var opItem SimpleItem
		if err = json.Unmarshal(b, &opItem); err != nil {
			return nil, err
		}

		return &opItem, nil
	}

	return nil, errors.New("unable to find item")
}
