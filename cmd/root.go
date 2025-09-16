package cmd

import (
	"github.com/asips/sdtp-client/internal"
	"github.com/spf13/cobra"
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

		flags := cmd.Flags()
		certPath, err := flags.GetString("cert")
		cobra.CheckErr(err)
		keyPath, err := flags.GetString("key")
		cobra.CheckErr(err)
		checkCertDays, err := flags.GetInt("check-cert-days")
		cobra.CheckErr(err)
		checkCertExprFlag, err := flags.GetBool("check-cert-expr")
		cobra.CheckErr(err)

		if checkCertExprFlag {
			mustValidateCert(certPath, keyPath, checkCertDays)
		}
		return nil
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.StringP("api-url", "u", "https://sips-data.ssec.wisc.edu/rivet/v1", "SDTP API base url")
	flags.StringP("cert", "c", "", "Path to PEM encoded client certificate.")
	flags.StringP("key", "k", "", "Path to PEM encoded client private key")
	flags.Bool("check-cert-expr", true, "Set to false to skip checking cert expiration")
	flags.Int("check-cert-days", 30, "Number of days before cert expiration to issue a warning")

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
