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

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check a new client certificate with the server",
	Long:  "check a new client certificate with the server",
	Run: func(cmd *cobra.Command, args []string) {
		apiUrl := parseApiUrl(strApiUrl)
		docheck(apiUrl)
	},
}

func docheck(apiUrl *url.URL) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sdtp, err := internal.NewSDTP(apiUrl, certPath, keyPath, httpTimeout)
	if err != nil {
		log.Fatal("Failed to create SDTP client: %s", err)
	}

	certInfo, err := getCertificateInfo(certPath, keyPath)
	if err != nil {
		log.Warn("Failed to get certification info; skipping cert expriation check: %s", err)
	}

	if certInfo.Expired {
		log.Printf(`ERROR:   Certificate Expired!
ERROR:
ERROR:   DN               %s
ERROR:   Expiration Date: %s
ERROR:   Issuer:          %s
ERROR:

`, certInfo.DN, certInfo.Expiration.Format(time.RFC3339), certInfo.Issuer)
		os.Exit(3)
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
			log.Fatal("FAILED: failed to authenticate using provided cert and key")
		case internal.ErrForbidden:
			log.Fatal("FAILED: authenticated successfully (certificate works); Not authorized to access /files endpoint")
		}
		log.Fatal("FAILED: failed for a non-auth related reason: %s", err)
	}
	log.Printf("Successfully connected to server and performed a HEAD request to the /files endpoint.")
}
