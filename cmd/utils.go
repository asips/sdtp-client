package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asips/sdtp-client/internal/log"
)

type CertInfo struct {
	DaysLeft   int
	Expiration time.Time
	Expired    bool
	Issuer     string
}

// getCertificateInfo reads and parses a PEM encoded certificate file. There must be exactly
// one certificate in the file, i.e., it must not be a certificate chain.
func getCertificateInfo(certFile, keyFile string) (CertInfo, error) {
	chain, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return CertInfo{}, fmt.Errorf("failed to load certificate: %w", err)
	}
	cert := chain.Leaf
	return CertInfo{
		DaysLeft:   -int(time.Since(cert.NotAfter).Hours() / 24),
		Expiration: cert.NotAfter,
		Expired:    time.Now().After(cert.NotAfter),
		Issuer:     cert.Issuer.String(),
	}, nil
}

func ensureCertExpr(certFile, keyFile string, checkCertDays int) {
	certInfo, err := getCertificateInfo(certFile, keyFile)
	if err != nil {
		log.Warn("Failed to get certification info; skipping cert expriation check: %s", err)
	}

	if certInfo.Expired {
		log.Printf(`ERROR:   Certificate Expired!
ERROR:
ERROR:   Expiration Date: %s
ERROR:   Issuer:          %s
ERROR:
`, certInfo.Expiration.Format(time.RFC3339), certInfo.Issuer)
		os.Exit(3)
	} else if certInfo.DaysLeft > 0 && certInfo.DaysLeft <= checkCertDays {
		log.Printf(`WARNING:    Certificate expires soon!
WARNING:
WARNING:    Expiration Date: %s
WARNING:    Days Left:       %d
WARNING:    Issuer:          %s
WARNING:
`, certInfo.Expiration.Format(time.RFC3339), certInfo.DaysLeft, certInfo.Issuer)
	} else {
		log.Printf("Certificate OK! expires on %s (in %d days), issued by %s", certInfo.Expiration.Format(time.RFC3339), certInfo.DaysLeft, certInfo.Issuer)
	}
}

// ExitHandler returns a context.Context that will be canceled if any of the
// provided signals are received.
func exitHandler(sig ...os.Signal) context.Context {
	if len(sig) == 0 {
		sig = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sig...)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		log.Info("canceled %s", <-ch)
		cancel()
	}()
	return ctx
}

func parseApiUrl(strUrl string) *url.URL {
	u, err := url.Parse(strApiUrl)
	if err != nil {
		log.Fatal("invalid api-url: %s", err)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		log.Fatal("api-url must not contain query or fragment")
	}
	return u
}
