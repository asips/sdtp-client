package internal

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var (
	ErrNotAuthorized = fmt.Errorf("unable to authenticate with the provided certificate")
	ErrNotFound      = fmt.Errorf("not found")
	ErrForbidden     = fmt.Errorf("authenticated, but no permissions to the resource")
	ErrExists        = fmt.Errorf("already exists")
)

// file returned to the client
type FileInfo struct {
	ID       int64             `json:"fileid"`
	Name     string            `json:"name"`
	Checksum string            `json:"checksum"`
	Size     int64             `json:"size"`
	Expires  string            `json:"expires"`
	Tags     map[string]string `json:"tags"`
	Extra    map[string]any    `json:"extra"`
}

// SDTPClient is the interface for interacting with the SDTP server
//
// TODO: This interface is too heavy
type FileListor interface {
	List(ctx context.Context, tags map[string]string) ([]FileInfo, error)
}
type FileDownloader interface {
	Download(ctx context.Context, file FileInfo, destDir string) error
	Ack(ctx context.Context, file FileInfo) error
}

type SDTPClient interface {
	FileListor
	FileDownloader
	Register(ctx context.Context) error
	Check(ctx context.Context) error
}

type DefaultSDTPClient struct {
	client *http.Client
	apiUrl *url.URL
}

func NewDefaultSDTP(apiUrl *url.URL, certFile, keyFile string, timeout time.Duration) (*DefaultSDTPClient, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	return &DefaultSDTPClient{
		apiUrl: apiUrl,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
					// disable TLS 1.3 to avoid Apache SSL error "Re-negotiation handshake failed"
					MaxVersion: tls.VersionTLS12,
					// set to avoid Apahce SSL error "SSL Library Error: error:0A000153:SSL routines::no renegotiation"
					Renegotiation: tls.RenegotiateOnceAsClient,
				},
			},
			Timeout: timeout,
		}}, nil
}

func (s *DefaultSDTPClient) mustNewReq(ctx context.Context, method, url string) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create request: %v", err))
	}
	return req
}

func (s *DefaultSDTPClient) List(ctx context.Context, tags map[string]string) ([]FileInfo, error) {
	qry := url.Values{}
	for k, v := range tags {
		qry.Set(k, v)
	}
	epUrl := fmt.Sprintf("%s/files?%s", s.apiUrl.String(), qry.Encode())

	req := s.mustNewReq(ctx, http.MethodGet, epUrl)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to setup request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrNotAuthorized
	case http.StatusForbidden:
		return nil, ErrForbidden
	case http.StatusNotFound:
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	var listResp struct {
		Files []FileInfo `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return listResp.Files, nil
}

func (s *DefaultSDTPClient) Download(ctx context.Context, file FileInfo, destDir string) error {
	epUrl := fmt.Sprintf("%s/files/%d", s.apiUrl, file.ID)

	req := s.mustNewReq(ctx, http.MethodGet, epUrl)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to setup request: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrNotAuthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	destPath := path.Join(destDir, "."+file.Name)
	dest, err := newWriter(destPath, file.Checksum)
	if err != nil {
		return fmt.Errorf("failed to create dest: %w", err)
	}

	if _, err = io.Copy(dest, resp.Body); err != nil {
		return fmt.Errorf("failed to write to %s: %w", destPath, err)
	}
	dest.Close()

	if !dest.ChecksumMatches() {
		os.Remove(destPath)
		return fmt.Errorf("checksum mismatch for %s; got %s, wanted %s", file.Name, dest.Computed(), file.Checksum)
	}
	if err := os.Rename(destPath, path.Join(destDir, file.Name)); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", destPath, file.Name, err)
	}
	return nil
}

func (s *DefaultSDTPClient) Ack(ctx context.Context, file FileInfo) error {
	epUrl := fmt.Sprintf("%s/files/%d", s.apiUrl, file.ID)

	req := s.mustNewReq(ctx, http.MethodDelete, epUrl)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to setup request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrNotAuthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusNoContent, http.StatusOK:
		// Ok response should be a 204, but let's call 200 ok too
		return nil
	}

	return fmt.Errorf("request failed: %s", resp.Status)
}

func (s *DefaultSDTPClient) Register(ctx context.Context) error {
	epUrl := fmt.Sprintf("%s/register", s.apiUrl)

	req := s.mustNewReq(ctx, http.MethodPut, epUrl)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to setup request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrNotAuthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		return ErrExists
	case http.StatusOK, http.StatusCreated:
		return nil
	}
	return fmt.Errorf("request failed: %s", resp.Status)
}

func (s *DefaultSDTPClient) Check(ctx context.Context) error {
	epUrl := fmt.Sprintf("%s/files", s.apiUrl)

	req := s.mustNewReq(ctx, http.MethodGet, epUrl)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to setup request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrNotAuthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		return ErrExists
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return nil
}

var _ SDTPClient = (*DefaultSDTPClient)(nil)

type writer struct {
	w            io.WriteCloser
	h            hash.Hash
	expectedCsum string
}

func (w *writer) Close() error {
	return w.w.Close()
}

func (w *writer) Write(p []byte) (int, error) {
	_, err := w.h.Write(p)
	if err != nil {
		return 0, fmt.Errorf("checksum err: %w", err)
	}
	return w.w.Write(p)
}

func (w *writer) ChecksumMatches() bool {
	return w.expectedCsum == w.Computed()
}

func (w *writer) Computed() string {
	return strings.ToLower(fmt.Sprintf("%x", w.h.Sum(nil)))
}

func newWriter(destPath, checksum string) (*writer, error) {
	alg, checksumVal, found := strings.Cut(checksum, ":")
	if !found {
		return nil, fmt.Errorf("invalid checksum format")
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open destination %s: %w", destPath, err)
	}

	var hash hash.Hash
	switch strings.ToLower(alg) {
	case "sha256":
		hash = sha256.New()
	case "sha384":
		hash = sha512.New384()
	case "sha512":
		hash = sha512.New()
	case "md5":
		hash = md5.New()
	default:
		return nil, fmt.Errorf("%s checksum not supported", alg)
	}
	return &writer{dest, hash, checksumVal}, nil
}
