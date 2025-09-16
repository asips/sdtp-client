package cmd

import (
	"context"
	"encoding/json"
	"fmt"
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
		flags := cmd.Flags()
		certPath, err := flags.GetString("cert")
		cobra.CheckErr(err)
		keyPath, err := flags.GetString("key")
		cobra.CheckErr(err)
		httpTimeout, err := flags.GetDuration("http-timeout")
		cobra.CheckErr(err)
		checkCertDays, err := flags.GetInt("check-cert-days")
		cobra.CheckErr(err)

		mustValidateCert(certPath, keyPath, checkCertDays)

		strApiUrl, err := flags.GetString("api-url")
		cobra.CheckErr(err)
		apiUrl := parseApiUrl(strApiUrl)
		sdtp, err := internal.NewDefaultSDTP(apiUrl, certPath, keyPath, httpTimeout)
		if err != nil {
			log.Fatal("Failed to create SDTP client: %s", err)
		}

		tags, err := flags.GetStringToString("tag")
		cobra.CheckErr(err)

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		_, err = doList(ctx, sdtp, tags)
		if err != nil {
			log.Fatal("Failed to list files: %s", err)
		}

		return nil
	},
}

func init() {
	flags := listCmd.Flags()

	flags.Duration("http-timeout", time.Second*60, "HTTP client timeout in seconds for list operations")
	flags.StringToStringP("tag", "t", map[string]string{}, "<key>=<value> tags to filter by. May be specified multiple times or as a comma-separated list")
}

func doList(ctx context.Context, sdtp internal.SDTPClient, tags map[string]string) (int, error) {
	files, err := sdtp.List(ctx, tags)
	if err != nil {
		return 0, fmt.Errorf("Failed to list files: %s", err)
	}

	if len(files) == 0 {
		log.Printf("No files found")
		return 0, nil
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

	return len(files), nil
}
