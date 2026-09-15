package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalDisk stores files under a base directory on the local filesystem and
// serves them back through a public HTTP prefix (see router: static file route).
type LocalDisk struct {
	baseDir   string
	publicURL string // e.g. "http://localhost:8080/uploads" or "/uploads" behind nginx
}

func NewLocalDisk(baseDir, publicURL string) *LocalDisk {
	return &LocalDisk{baseDir: baseDir, publicURL: strings.TrimRight(publicURL, "/")}
}

func (l *LocalDisk) Save(_ context.Context, folder, filename string, reader io.Reader, _ int64, contentType string) (*ObjectMeta, error) {
	ext := filepath.Ext(filename)
	key := filepath.ToSlash(filepath.Join(folder, uuid.NewString()+ext))
	fullPath := filepath.Join(l.baseDir, filepath.FromSlash(key))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, reader)
	if err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return &ObjectMeta{
		Key:      key,
		URL:      l.URL(key),
		Size:     written,
		MimeType: contentType,
	}, nil
}

func (l *LocalDisk) Open(_ context.Context, key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(l.baseDir, filepath.FromSlash(key)))
}

func (l *LocalDisk) Delete(_ context.Context, key string) error {
	err := os.Remove(filepath.Join(l.baseDir, filepath.FromSlash(key)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (l *LocalDisk) URL(key string) string {
	return l.publicURL + "/" + key
}
