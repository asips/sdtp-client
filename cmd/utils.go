package cmd

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/asips/sdtp-client/internal/log"
)

type CertInfo struct {
	DaysLeft   int
	Expiration time.Time
	Expired    bool
	Issuer     string
}

func getCertificateInfo(certFile, keyFile string) (CertInfo, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return CertInfo{}, fmt.Errorf("failed to load key pair: %w", err)
	}

	// TODO: Determine if it's safe to assume the Leaf is what we want
	leaf := cert.Leaf
	return CertInfo{
		DaysLeft:   int(time.Since(leaf.NotAfter).Hours() / 24),
		Expiration: leaf.NotAfter,
		Expired:    time.Now().After(leaf.NotAfter),
		Issuer:     leaf.Issuer.String(),
	}, nil
}

func checkCertExpr(certFile, keyFile string, checkCertDays int) bool {
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
		return true
	} else if certInfo.DaysLeft > 0 && certInfo.DaysLeft <= checkCertDays {
		log.Printf(`WARNING:    Certificate expires soon!
WARNING:
WARNING:    Expiration Date: %s
WARNING:    Days Left:       %d
WARNING:    Issuer:          %s
WARNING:
`, certInfo.Expiration.Format(time.RFC3339), certInfo.DaysLeft, certInfo.Issuer)
	} else {
		log.Printf("Certificate expires on %s (in %d days), issued by %s", certInfo.Expiration.Format(time.RFC3339), -certInfo.DaysLeft, certInfo.Issuer)
	}
	return false
}
