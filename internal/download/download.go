// Package download retrieves files over HTTP(S) with a simple progress
// indicator.
package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// DefaultTimeout is applied to individual HTTP requests when no context
// deadline is set.
const DefaultTimeout = 5 * time.Minute

// ProgressFunc is called periodically with the number of bytes transferred
// and the total size in bytes (0 when the server does not advertise a
// Content-Length).
type ProgressFunc func(current, total int64)

// File downloads src to dest, replacing the destination atomically.
func File(ctx context.Context, src, dest string, progress ProgressFunc) error {
	if _, err := url.Parse(src); err != nil {
		return fmt.Errorf("invalid url %q: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create download dir: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "armada/0.1")

	client := &http.Client{Timeout: DefaultTimeout}
	if _, deadlined := ctx.Deadline(); deadlined {
		client.Timeout = 0
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", src, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %d", src, resp.StatusCode)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), filepath.Base(dest)+".*.part")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	total := resp.ContentLength
	var reader io.Reader = resp.Body
	if progress != nil {
		reader = &progressReader{r: resp.Body, total: total, progress: progress}
	}

	if _, err := io.Copy(tmp, reader); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dest, err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("rename %s: %w", dest, err)
	}
	if progress != nil {
		// Flush with the final size so the caller sees 100%.
		info, _ := os.Stat(dest)
		if info != nil {
			progress(info.Size(), info.Size())
		}
	}
	return nil
}

type progressReader struct {
	r        io.Reader
	total    int64
	current  int64
	progress ProgressFunc
	lastTick time.Time
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.current += int64(n)
		if time.Since(p.lastTick) > 100*time.Millisecond {
			p.progress(p.current, p.total)
			p.lastTick = time.Now()
		}
	}
	return n, err
}
