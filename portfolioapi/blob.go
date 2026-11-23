package portfolioapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// The Blob service uses the same HTTP protocol as Vercel's Blob SDK. Private
// reads bypass the CDN so a conditional write always starts from current data.
const blobAPI = "https://vercel.com/api/blob/"
const maxBlobBytes = 32 << 20

var blobClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type blobDocument struct {
	storeID string
	token   string
	key     string
}

func (doc *blobDocument) read(ctx context.Context) ([]byte, error) {
	data, _, err := doc.readVersion(ctx)
	return data, err
}

func (doc *blobDocument) readVersion(ctx context.Context) ([]byte, string, error) {
	address := "https://" + doc.storeID + ".private.blob.vercel-storage.com/" + doc.key + "?cache=0"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+doc.token)
	// Compressed responses have weak ETags, which cannot guard a Blob write.
	req.Header.Set("Accept-Encoding", "identity")
	res, err := blobClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, "", os.ErrNotExist
	}
	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("private Blob read returned HTTP %d", res.StatusCode)
	}
	revision := res.Header.Get("ETag")
	if revision == "" {
		return nil, "", errors.New("private Blob did not return an ETag")
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxBlobBytes+1))
	if len(data) > maxBlobBytes {
		return nil, "", errors.New("private Blob exceeded document size limit")
	}
	return data, revision, err
}

func (doc *blobDocument) update(ctx context.Context, edit func([]byte) ([]byte, error)) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	for attempt := 0; attempt < 5; attempt++ {
		current, revision, err := doc.readVersion(ctx)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		next, err := edit(current)
		if err != nil || next == nil {
			return err
		}
		retry, err := doc.put(ctx, next, revision)
		if err != nil || !retry {
			return err
		}
	}
	return errors.New("private Blob update conflicted repeatedly")
}

// put returns retry when a concurrent caller created or replaced the document.
func (doc *blobDocument) put(ctx context.Context, data []byte, revision string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, blobAPI+"?pathname="+url.QueryEscape(doc.key), bytes.NewReader(data))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+doc.token)
	req.Header.Set("X-API-Version", "13")
	req.Header.Set("X-Vercel-Blob-Store-ID", doc.storeID)
	req.Header.Set("X-Vercel-Blob-Access", "private")
	req.Header.Set("X-Content-Type", "application/json")
	req.Header.Set("X-Add-Random-Suffix", "0")
	if revision != "" {
		req.Header.Set("X-Allow-Overwrite", "1")
		req.Header.Set("X-If-Match", revision)
	}
	res, err := blobClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		_, _ = io.Copy(io.Discard, res.Body)
		return false, nil
	}
	var failure struct {
		Error struct {
			Code    string
			Message string
		}
	}
	_ = json.NewDecoder(io.LimitReader(res.Body, 8192)).Decode(&failure)
	conflict := res.StatusCode == http.StatusPreconditionFailed || res.StatusCode == http.StatusConflict ||
		failure.Error.Code == "precondition_failed" ||
		(revision == "" && strings.Contains(failure.Error.Message, "already exists"))
	if conflict {
		return true, nil
	}
	return false, fmt.Errorf("private Blob write returned HTTP %d (%s)", res.StatusCode, failure.Error.Code)
}
