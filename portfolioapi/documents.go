package portfolioapi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// document updates compare against the contents that the callback saw, and no-op on a nil write.
type document interface {
	read(context.Context) ([]byte, error)
	update(context.Context, func(current []byte) (next []byte, err error)) error
}

func openStatsDocument(env func(string) string, statsFile string) (document, error) {
	if id := strings.TrimPrefix(env("BLOB_STORE_ID"), "store_"); id != "" {
		token := env("BLOB_READ_WRITE_TOKEN")
		if token == "" {
			return nil, errors.New("BLOB_READ_WRITE_TOKEN is required for the private Blob store")
		}
		return &blobDocument{storeID: id, token: token, key: "agentic-stats.json"}, nil
	}
	if env("VERCEL") == "1" {
		return nil, errors.New("configure a private Blob store for the deployed API")
	}
	statsPath, err := filepath.Abs(statsFile)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(statsPath), 0o700); err != nil {
		return nil, err
	}
	return &fileDocument{path: statsPath}, nil
}

type fileDocument struct {
	path  string
	mutex sync.Mutex
}

func (doc *fileDocument) read(_ context.Context) ([]byte, error) {
	return os.ReadFile(doc.path)
}

func (doc *fileDocument) update(ctx context.Context, edit func([]byte) ([]byte, error)) error {
	doc.mutex.Lock()
	defer doc.mutex.Unlock()
	current, err := doc.read(ctx)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	next, err := edit(current)
	if err != nil || next == nil {
		return err
	}
	path, err := filepath.EvalSymlinks(doc.path)
	if errors.Is(err, os.ErrNotExist) {
		path = doc.path
	} else if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".portfolio-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(next)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
