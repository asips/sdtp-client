package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func createMockClient(fn RoundTripFunc) *DefaultSDTPClient {
	return &DefaultSDTPClient{
		client: &http.Client{
			Transport: RoundTripFunc(fn),
		},
		apiUrl: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
			Path:   "/sdtp",
		},
	}
}

func TestList(t *testing.T) {

	t.Run("nominal", func(t *testing.T) {
		body := `{"files": [
		{"id":1,"name":"file1.txt","size":1234,"tags":{"stream":"test"}},
		{"id":2,"name":"file2.txt","size":1234,"tags":{"stream":"test"}},
		{"id":3,"name":"file3.txt","size":1234,"tags":{"stream":"test"}}
		]}`
		sdtp := createMockClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
			}
		})

		files, err := sdtp.List(t.Context(), map[string]string{})

		assert.NoError(t, err)
		assert.Len(t, files, 3)
	})

	tests := []struct {
		Status int
		Err    error
	}{
		{http.StatusForbidden, ErrForbidden},
		{http.StatusUnauthorized, ErrNotAuthorized},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status=%d", tt.Status), func(t *testing.T) {
			sdtp := createMockClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: tt.Status,
					Body:       http.NoBody,
				}
			})

			_, err := sdtp.List(t.Context(), map[string]string{})

			assert.Equal(t, tt.Err, err)
		})
	}
}

func TestDownload(t *testing.T) {
	t.Run("nominal", func(t *testing.T) {
		tmpdir := t.TempDir()
		body := `xxx`
		sdtp := createMockClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
			}
		})

		err := sdtp.Download(t.Context(), FileInfo{
			ID:       1,
			Name:     "file1.txt",
			Checksum: "md5:f561aaf6ef0bf14d4208bb46a4ccb3ad",
		}, tmpdir)

		if assert.NoError(t, err) {
			data, err := os.ReadFile(tmpdir + "/file1.txt")
			assert.NoError(t, err)
			assert.Equal(t, body, string(data))
		}
	})

	tests := []struct {
		Status int
		Err    error
	}{
		{http.StatusForbidden, ErrForbidden},
		{http.StatusUnauthorized, ErrNotAuthorized},
		{http.StatusNotFound, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status=%d", tt.Status), func(t *testing.T) {
			sdtp := createMockClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: tt.Status,
					Body:       http.NoBody,
				}
			})

			_, err := sdtp.List(t.Context(), map[string]string{})

			assert.Equal(t, tt.Err, err)
		})
	}
}

func TestAck(t *testing.T) {
	tests := []struct {
		Status int
		Err    error
	}{
		{http.StatusForbidden, ErrForbidden},
		{http.StatusUnauthorized, ErrNotAuthorized},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusOK, nil},
		{http.StatusNoContent, nil},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status=%d", tt.Status), func(t *testing.T) {
			sdtp := createMockClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: tt.Status,
					Body:       http.NoBody,
				}
			})

			err := sdtp.Ack(t.Context(), FileInfo{
				ID:       1,
				Name:     "file1.txt",
				Checksum: "md5:f561aaf6ef0bf14d4208bb46a4ccb3ad",
			})

			assert.Equal(t, tt.Err, err)
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		Status int
		Err    error
	}{
		{http.StatusForbidden, ErrForbidden},
		{http.StatusUnauthorized, ErrNotAuthorized},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusOK, nil},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status=%d", tt.Status), func(t *testing.T) {
			sdtp := createMockClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: tt.Status,
					Body:       http.NoBody,
				}
			})

			err := sdtp.Check(t.Context())

			assert.Equal(t, tt.Err, err)
		})
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		Status int
		Err    error
	}{
		{http.StatusForbidden, ErrForbidden},
		{http.StatusUnauthorized, ErrNotAuthorized},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusConflict, ErrExists},
		{http.StatusOK, nil},
		{http.StatusCreated, nil},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status=%d", tt.Status), func(t *testing.T) {
			sdtp := createMockClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: tt.Status,
					Body:       http.NoBody,
				}
			})

			err := sdtp.Register(t.Context())

			assert.Equal(t, tt.Err, err)
		})
	}
}
