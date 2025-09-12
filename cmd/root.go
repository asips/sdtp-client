package cmd

import (
	"time"

	"github.com/asips/sdtp-client/internal"
	"github.com/spf13/cobra"
)

var (
	strApiUrl         string
	certPath          string
	keyPath           string
	checkCertExprFlag bool
	checkCertDays     int
	httpTimeout       time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "sdtp",
	Short: "Atmosphere SIPS SDTP Client",
	Long: `sdtp is a command line client for download files from an SDTP Provider.

Authentication is done using x509 client certificates, therefore you must provide a
valid client certificate and private key. The certificate must be signed by a CA trusted
by the SDTP sever to successfully authenticate (connection will fail otherwise).

References:
- Project Repository,
  https://github.com/asips/sdtp-client
- Science Data Transfer Protocol (SDTP) Interface Control Document (ICD), 
  https://www.earthdata.nasa.gov/s3fs-public/2023-11/423-ICD-027_SDTP_ICD_Original.pdf
`,
	Version: internal.Version + " (" + internal.GitSHA + ")",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkCertExprFlag {
			checkCert(certPath, keyPath, checkCertDays)
		}
		return nil
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&strApiUrl, "api-url", "u", "https://sips-data.ssec.wisc.edu/rivet/v1", "SDTP API base url")
	flags.StringVarP(&certPath, "cert", "c", "", "Path to PEM encoded client certificate.")
	flags.StringVarP(&keyPath, "key", "k", "", "Path to PEM encoded client private key")
	flags.BoolVar(&checkCertExprFlag, "check-cert-expr", true, "Set to false to skip checking cert expiration")
	flags.IntVar(&checkCertDays, "check-cert-days", 30, "Number of days before cert expiration to issue a warning")

	rootCmd.MarkPersistentFlagRequired("cert")
	rootCmd.MarkPersistentFlagRequired("key")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(checkCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
