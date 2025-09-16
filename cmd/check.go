package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asips/sdtp-client/internal"
	"github.com/asips/sdtp-client/internal/log"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check a new client certificate with the server",
	Long:  "check a new client certificate with the server",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		certPath, err := flags.GetString("cert")
		cobra.CheckErr(err)
		keyPath, err := flags.GetString("key")
		cobra.CheckErr(err)
		checkCertDays, err := flags.GetInt("check-cert-days")
		cobra.CheckErr(err)
		httpTimeout, err := flags.GetDuration("http-timeout")
		cobra.CheckErr(err)
		apiUrlStr, err := flags.GetString("api-url")
		cobra.CheckErr(err)
		apiUrl := parseApiUrl(apiUrlStr)
		sdtp, err := internal.NewDefaultSDTP(apiUrl, certPath, keyPath, httpTimeout)
		if err != nil {
			log.Fatal("Failed to create SDTP client: %s", err)
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		err = doCheck(ctx, sdtp, certPath, keyPath, checkCertDays, getCertificateInfo)
		if err == errCertExpired {
			os.Exit(3)
		} else if err != nil {
			os.Exit(1)
		}
	},
}

var errCertExpired = fmt.Errorf("certificate expired")

func doCheck(ctx context.Context, sdtp internal.SDTPClient, certPath, keyPath string, checkCertDays int, certParser certParserFunc) error {

	certInfo, err := certParser(certPath, keyPath)
	if err != nil {
		log.Printf("Failed to get certification info; skipping cert expriation check: %s", err)
	}

	if certInfo.Expired {
		log.Printf(`ERROR:   Certificate Expired!
ERROR:
ERROR:   DN               %s
ERROR:   Expiration Date: %s
ERROR:   Issuer:          %s
ERROR:

`, certInfo.DN, certInfo.Expiration.Format(time.RFC3339), certInfo.Issuer)
		return errCertExpired
	} else if certInfo.DaysLeft > 0 && certInfo.DaysLeft <= checkCertDays {
		log.Printf(`WARNING:    Certificate expires soon!
WARNING:
WARNING:	DN:              %s
WARNING:    Expiration Date: %s
WARNING:    Days Left:       %d
WARNING:    Issuer:          %s
WARNING:

`, certInfo.DN, certInfo.Expiration.Format(time.RFC3339), certInfo.DaysLeft, certInfo.Issuer)
	} else {
		log.Printf(`Certificate Ok!

    DN:              %s
    Expiration Date: %s
    Days Left:       %d
    Issuer:          %s

`, certInfo.DN, certInfo.Expiration.Format(time.RFC3339), certInfo.DaysLeft, certInfo.Issuer)
	}

	err = sdtp.Check(ctx)
	if err != nil {
		switch err {
		case internal.ErrNotAuthorized:
			return fmt.Errorf("failed to authenticate using provided cert and key")
		case internal.ErrForbidden:
			return fmt.Errorf("authenticated successfully (certificate works); Not authorized to access /files endpoint")
		}
		return fmt.Errorf("failed for a non-auth related reason: %s", err)
	}
	log.Printf("Successfully connected to server and performed a HEAD request to the /files endpoint.")

	return nil
}
