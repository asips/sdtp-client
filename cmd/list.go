package cmd

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asips/sdtp-client/internal"
	"github.com/asips/sdtp-client/internal/log"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available items from the server based on provided tags",
	Long:  "List available items from the server based on provided tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiUrl := parseApiUrl(strApiUrl)
		doList(apiUrl, certPath, keyPath, tags, httpTimeout)

		return nil
	},
}

func init() {
	flags := listCmd.Flags()

	flags.DurationVar(&httpTimeout, "http-timeout", time.Second*60, "HTTP client timeout in seconds for list operations")
}

func doList(apiUrl *url.URL, certPath, keyPath string, tags map[string]string, timeout time.Duration) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sdtp, err := internal.NewSDTP(apiUrl, certPath, keyPath, timeout)
	if err != nil {
		log.Fatal("Failed to create SDTP client: %s", err)
	}

	files, err := sdtp.List(ctx, tags)
	if err != nil {
		log.Fatal("Failed to list files: %s", err)
	}

	for _, file := range files {
		log.Printf("ID: %d, Name: %s, Size: %d, Expires: %s", file.ID, file.Name, file.Size, file.Expires)
	}
}
