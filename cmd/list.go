package cmd

import (
	"context"
	"encoding/json"
	"fmt"
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
	Long: `List available items from the server based on provided tags.

Listed files will be printed as JSON objects, one per line to stdout. Any log messages
go to stderr.
`,
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

	if len(files) == 0 {
		log.Printf("No files found")
		return
	}

	log.Printf("Found %d files:", len(files))
	for _, file := range files {
		dat, err := json.Marshal(file)
		if err != nil {
			log.Printf("failed to marshal to json: %s", err)
			continue
		}
		fmt.Fprintf(os.Stdout, "%s\n", string(dat))
	}
}
