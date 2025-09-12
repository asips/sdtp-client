package cmd

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	destDir   string
	tags      map[string]string
	stream    string
	shortName string
	mission   string
	ackFlag   bool
	listFlag  bool
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest data from SDTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		apiUrl, err := url.Parse(strApiUrl)
		if err != nil {
			return fmt.Errorf("invalid api-url: %w", err)
		}
		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			os.MkdirAll(destDir, 0755)
		}
		if flags.Changed("stream") {
			tags["stream"] = stream
		}
		if flags.Changed("mission") {
			tags["mission"] = mission
		}
		if flags.Changed("short-name") {
			tags["ShortName"] = shortName
		}
		if checkCertExprFlag {
			checkCert(certPath, keyPath, checkCertDays)
		}

		if listFlag {
			doList(apiUrl, certPath, keyPath, tags, httpTimeout)
		}

		return doIngest(
			apiUrl,
			certPath,
			keyPath,
			destDir,
			tags,
			ackFlag,
		)
	},
}

func init() {
	flags := ingestCmd.Flags()

	flags.StringVarP(&destDir, "dest-dir", "d", "", "Local directory to ingest data to")
	flags.StringVar(&stream, "stream", "", "SDTP 'stream' field (query parameter)")
	flags.StringVar(&shortName, "short-name", "", "SDTP 'ShortName' field (query parameter)")
	flags.StringVar(&mission, "mission", "", "SDTP 'mission' field (query parameter)")
	flags.StringToStringVarP(&tags, "tag", "t", map[string]string{}, "<key>=<value> tags to filter by. May be specified multiple times or as a comma-separated list")
	flags.BoolVarP(&ackFlag, "ack", "a", false, "Acknowledge files after successful ingest")
	flags.BoolVar(&listFlag, "list", false, "List available files, but do not download")
	flags.DurationVar(&httpTimeout, "http-timeout", time.Minute*5, "HTTP client timeout in seconds for list operations")

	flags.MarkDeprecated("list", "use 'list' sub-command instead")
}

func doIngest(apiUrl *url.URL, certFile, keyFile, destDir string, tags map[string]string, ack bool) error {
	return nil
}
