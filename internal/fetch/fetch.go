// Package fetch is the outward side of the tool: it reaches the network and
// opens archives, and hands the rest plain bytes or a reader.
package fetch

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Client fetches URLs. One instance is shared, because a redirect chain and a
// connection pool are worth reusing across the builders that run in a row.
type Client struct {
	http *http.Client
}

// New returns a client with a timeout long enough for the Ukrainian corpus,
// which is the largest thing this downloads.
func New() *Client {
	return &Client{http: &http.Client{Timeout: 10 * time.Minute}}
}

// Bytes downloads a URL into memory.
func (c *Client) Bytes(url string) ([]byte, error) {
	body, err := c.open(url)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return io.ReadAll(body)
}

// Stream downloads a URL and hands the caller the body to read. The caller
// closes it.
func (c *Client) Stream(url string) (io.ReadCloser, error) { return c.open(url) }

// Bzip2 downloads and decompresses in process, which is what removes the
// bunzip2 dependency.
func (c *Client) Bzip2(url string) (io.ReadCloser, error) {
	body, err := c.open(url)
	if err != nil {
		return nil, err
	}
	return readCloser{Reader: bzip2.NewReader(body), closer: body}, nil
}

// Binary downloads an archive and writes one file out of it to dest,
// executable. Both shapes upstream uses are handled, zip and tar.gz.
func (c *Client) Binary(url, member, dest string) error {
	data, err := c.Bytes(url)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	switch {
	case strings.HasSuffix(url, ".zip"):
		return extractZip(data, member, dest)
	case strings.HasSuffix(url, ".tar.gz"), strings.HasSuffix(url, ".tgz"):
		return extractTarGz(data, member, dest)
	default:
		return fmt.Errorf("stet: unknown archive type: %s", url)
	}
}

func (c *Client) open(url string) (io.ReadCloser, error) {
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("stet: fetching %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("stet: fetching %s: %s", url, resp.Status)
	}
	return resp.Body, nil
}

func extractZip(data []byte, member, dest string) error {
	r, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		if path.Base(f.Name) != member {
			continue
		}
		src, err := f.Open()
		if err != nil {
			return err
		}
		defer src.Close()
		return writeExecutable(dest, src)
	}
	return fmt.Errorf("stet: %s not found in archive", member)
}

func extractTarGz(data []byte, member, dest string) error {
	gz, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if path.Base(header.Name) == member {
			return writeExecutable(dest, tr)
		}
	}
	return fmt.Errorf("stet: %s not found in archive", member)
}

// writeExecutable renames into place, so an interrupted download leaves no half
// a binary for the next run to trust.
func writeExecutable(dest string, src io.Reader) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".stet-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dest)
}

type readCloser struct {
	io.Reader
	closer io.Closer
}

func (r readCloser) Close() error { return r.closer.Close() }
