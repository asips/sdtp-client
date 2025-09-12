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
	ErrNotAuthorized = fmt.Errorf("not authorized")
	ErrNotFound      = fmt.Errorf("not found")
	ErrForbidden     = fmt.Errorf("forbidden")
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

type SDTP struct {
	client *http.Client
	apiUrl *url.URL
}

func NewSDTP(apiUrl *url.URL, certFile, keyFile string, timeout time.Duration) (*SDTP, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
		Timeout: timeout,
	}

	return &SDTP{client, apiUrl}, nil
}

func (s *SDTP) mustNewReq(ctx context.Context, method, url string) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to create request: %v", err))
	}
	return req
}

func (s *SDTP) List(ctx context.Context, tags map[string]string) ([]FileInfo, error) {
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

func (s *SDTP) Download(ctx context.Context, file FileInfo, destDir string) error {
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

	if !dest.Matches(file.Checksum) {
		os.Remove(destPath)
		return fmt.Errorf("checksum mismatch for %s", destPath)
	}
	if err := os.Rename(destPath, path.Join(destDir, file.Name)); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", destPath, file.Name, err)
	}
	return nil
}

func (s *SDTP) Ack(ctx context.Context, file FileInfo) error {
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
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return nil
}

func (s *SDTP) Register(ctx context.Context) error {
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
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return nil
}

func (s *SDTP) Check(ctx context.Context) error {
	epUrl := fmt.Sprintf("%s/files", s.apiUrl)

	req := s.mustNewReq(ctx, http.MethodHead, epUrl)
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
	return nil
}

type writer struct {
	w io.Writer
	h hash.Hash
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writer) Matches(checksum string) bool {
	sum := fmt.Sprintf("%x", w.h.Sum(nil))
	return checksum == sum
}

func newWriter(destPath, checksum string) (*writer, error) {
	alg, _, found := strings.Cut(checksum, ":")
	if !found {
		return nil, fmt.Errorf("invalid checksum format")
	}

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open destination %s: %w", destPath, err)
	}
	defer dest.Close()

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
	return &writer{dest, hash}, nil
}
