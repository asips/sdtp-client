package cmd

import (
	"fmt"
	"net/url"

	"github.com/asips/sdtp-client/internal"
	"github.com/asips/sdtp-client/internal/log"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available items from the server based on provided tags",
	Long:  "List available items from the server based on provided tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiUrl, err := url.Parse(strApiUrl)
		if err != nil {
			return fmt.Errorf("invalid api-url: %w", err)
		}
		doList(apiUrl, certPath, keyPath, tags)

		return nil
	},
}

func doList(apiUrl *url.URL, certPath, keyPath string, tags map[string]string) {
	sdtp, err := internal.NewSDTP(apiUrl, certPath, keyPath)
	if err != nil {
		log.Fatal("Failed to create SDTP client: %s", err)
	}

	files, err := sdtp.List(tags)
	if err != nil {
		log.Fatal("Failed to list files: %s", err)
	}

	for _, file := range files {
		log.Printf("ID: %d, Name: %s, Size: %d, Expires: %s", file.ID, file.Name, file.Size, file.Expires)
	}
}
