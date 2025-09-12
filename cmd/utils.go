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
	DN         string
}

var oid = map[string]string{
	"2.5.4.3":                    "CN",
	"2.5.4.4":                    "SN",
	"2.5.4.5":                    "serialNumber",
	"2.5.4.6":                    "C",
	"2.5.4.7":                    "L",
	"2.5.4.8":                    "ST",
	"2.5.4.9":                    "streetAddress",
	"2.5.4.10":                   "O",
	"2.5.4.11":                   "OU",
	"2.5.4.12":                   "title",
	"2.5.4.17":                   "postalCode",
	"2.5.4.42":                   "GN",
	"2.5.4.43":                   "initials",
	"2.5.4.44":                   "generationQualifier",
	"2.5.4.46":                   "dnQualifier",
	"2.5.4.65":                   "pseudonym",
	"0.9.2342.19200300.100.1.25": "DC",
	"1.2.840.113549.1.9.1":       "emailAddress",
	"0.9.2342.19200300.100.1.1":  "userid",
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
		DN:         cert.Subject.ToRDNSequence().String(),
	}, nil
}

func checkCert(certFile, keyFile string, days int) {
	info, err := getCertificateInfo(certPath, keyPath)
	if err != nil {
		log.Warn("Failed to get certificate info: %s", err)
	}
	if info.Expired {
		log.Printf("Certificate expired on %s, run 'check' for more info", info.Expiration.Format(time.RFC3339))
		os.Exit(3)
	}
	if info.DaysLeft > 0 && info.DaysLeft <= days {
		log.Info("WARNING!! Certificate expiring in %s days; run 'check' for more info", info.DaysLeft)
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
