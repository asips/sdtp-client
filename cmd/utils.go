package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/asips/sdtp-client/internal"
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

// getCertificatePrintf reads and parses a PEM encoded certificate file. There must be exactly
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

type certParserFunc = func(certFile, keyFile string) (CertInfo, error)

func mustValidateCert(certFile, keyFile string, days int) {
	info, err := getCertificateInfo(certFile, keyFile)
	if err != nil {
		log.Printf("Failed to get certificate info: %s", err)
	}
	if info.Expired {
		log.Printf("Certificate expired on %s, run 'check' for more info", info.Expiration.Format(time.RFC3339))
		os.Exit(3)
	}
	if info.DaysLeft > 0 && info.DaysLeft <= days {
		log.Printf("WARNING!! Certificate expiring in %d days; run 'check' for more info", info.DaysLeft)
	}
}

func parseApiUrl(strUrl string) *url.URL {
	u, err := url.Parse(strUrl)
	if err != nil {
		log.Fatal("invalid api-url: %s", err)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		log.Fatal("api-url must not contain query or fragment")
	}
	return u
}

type mockSDTP struct {
	err     error
	listing []internal.FileInfo
}

func (m *mockSDTP) Check(ctx context.Context) error {
	return m.err
}
func (m *mockSDTP) List(ctx context.Context, tags map[string]string) ([]internal.FileInfo, error) {
	return m.listing, m.err
}

func (m *mockSDTP) Download(ctx context.Context, file internal.FileInfo, destDir string) error {
	return m.err
}
func (m *mockSDTP) Ack(ctx context.Context, file internal.FileInfo) error {
	return m.err
}
func (m *mockSDTP) Register(ctx context.Context) error {
	return m.err
}

func createMockSDTP(t *testing.T) *mockSDTP {
	t.Helper()

	return &mockSDTP{err: nil}
}
