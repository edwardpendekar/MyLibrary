// Package storage abstracts "where uploaded binaries live" behind a single
// interface so the rest of the app (handlers, services) never talks to the
// filesystem or S3 directly. Swapping STORAGE_PROVIDER=s3 in config is enough
// to move covers/PDFs to S3-compatible storage without touching call sites.
package storage

import (
	"context"
	"io"
)

// ObjectMeta is returned after a successful Save.
type ObjectMeta struct {
	Key      string // storage key/path, persisted in files.stored_path
	URL      string // URL the frontend can use to fetch the object
	Size     int64
	MimeType string
}

type Provider interface {
	// Save streams reader into storage under a key derived from folder+filename
	// and returns where it ended up. Implementations must generate collision-safe keys.
	Save(ctx context.Context, folder, filename string, reader io.Reader, size int64, contentType string) (*ObjectMeta, error)

	// Open returns a reader for a previously stored key. Caller must Close it.
	Open(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the object at key. Not finding it is not an error.
	Delete(ctx context.Context, key string) error

	// URL returns a client-fetchable URL for key without touching storage.
	URL(key string) string
}
