package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/asips/sdtp-client/internal"
	"github.com/stretchr/testify/assert"
)

func Test_doCheck(t *testing.T) {

	t.Run("Expired cert is fatal", func(t *testing.T) {
		fakeCertParser := func(certPath, keyPath string) (CertInfo, error) {
			return CertInfo{
				DN:         "CN=Test,O=Example,C=US",
				Issuer:     "CN=Example CA,O=Example,C=US",
				Expiration: time.Now().Add(5 * 24 * time.Hour),
				DaysLeft:   5,
				Expired:    true,
			}, nil
		}

		err := doCheck(t.Context(), nil, "path/to/cert", "path/to/key", 10, fakeCertParser)

		assert.Equal(t, errCertExpired, err)
	})

	t.Run("Valid cert", func(t *testing.T) {
		fakeCertParser := func(certPath, keyPath string) (CertInfo, error) {
			return CertInfo{
				DN:         "CN=Test,O=Example,C=US",
				Issuer:     "CN=Example CA,O=Example,C=US",
				Expiration: time.Now().Add(5 * 24 * time.Hour),
				DaysLeft:   5,
				Expired:    false,
			}, nil
		}
		sdtp := createMockSDTP(t)

		sdtp.err = internal.ErrNotAuthorized
		err := doCheck(t.Context(), sdtp, "path/to/cert", "path/to/key", 10, fakeCertParser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate")

		sdtp.err = internal.ErrForbidden
		err = doCheck(t.Context(), sdtp, "path/to/cert", "path/to/key", 10, fakeCertParser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Not authorized to access")

		sdtp.err = fmt.Errorf("some other error")
		err = doCheck(t.Context(), sdtp, "path/to/cert", "path/to/key", 10, fakeCertParser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some other error")
	})
}
