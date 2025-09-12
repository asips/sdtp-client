package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/asips/sdtp-client/internal"
	"github.com/spf13/cobra"
)

var (
	strApiUrl         string
	certPath          string
	keyPath           string
	destDir           string
	tags              map[string]string
	stream            string
	shortName         string
	mission           string
	ackFlag           bool
	listFlag          bool
	checkCertExprFlag bool
	checkCertDays     int
)

var rootCmd = &cobra.Command{
	Use:     "sdtp",
	Short:   "Atmosphere SIPS SDTP Client",
	Version: internal.Version + " (" + internal.GitSHA + ")",
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
			if expired := checkCertExpr(certPath, keyPath, checkCertDays); expired {
				os.Exit(3)
			}
		}

		if listFlag {
			doList(apiUrl, certPath, keyPath, tags)
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
	flags := rootCmd.Flags()

	flags.StringVarP(&destDir, "dest-dir", "d", "", "Local directory to ingest data to")
	flags.StringVar(&stream, "stream", "", "SDTP 'stream' field (query parameter)")
	flags.StringVar(&shortName, "short-name", "", "SDTP 'ShortName' field (query parameter)")
	flags.StringVar(&mission, "mission", "", "SDTP 'mission' field (query parameter)")
	flags.StringToStringVarP(&tags, "tag", "t", map[string]string{}, "<key>=<value> tags to filter by. May be specified multiple times or as a comma-separated list")
	flags.BoolVarP(&ackFlag, "ack", "a", false, "Acknowledge files after successful ingest")
	flags.BoolVar(&listFlag, "list", false, "List available files, but do not download")

	flags = rootCmd.PersistentFlags()
	flags.StringVarP(&strApiUrl, "api-url", "u", "https://sips-data.ssec.wisc.edu/rivet/api/v1", "SDTP API base url")
	flags.StringVarP(&certPath, "cert", "c", "", "Path to client certificate")
	flags.StringVarP(&keyPath, "key", "k", "", "Path to client private key")
	flags.BoolVar(&checkCertExprFlag, "check-cert-expr", true, "Set to false to skip checking cert expiration")
	flags.IntVar(&checkCertDays, "check-cert-days", 30, "Number of days before cert expiration to issue a warning")

	rootCmd.MarkFlagRequired("cert")
	rootCmd.MarkFlagRequired("key")
	// flags.MarkDeprecated("short-name", "use --tag ShortName=<value> instead")
	// flags.MarkDeprecated("mission", "use --tag ShortName=<value> instead")
	// flags.MarkDeprecated("stream", "use --tag ShortName=<value> instead")
	flags.MarkDeprecated("list", "use 'list' sub-command instead")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(registerCmd)
}

func doIngest(apiUrl *url.URL, certFile, keyFile, destDir string, tags map[string]string, ack bool) error {
	return nil
}

func Execute() error {
	return rootCmd.Execute()
}
