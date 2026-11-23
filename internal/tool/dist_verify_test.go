package tool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// sha256URL must read a large body to the end (a package over the tailnet takes
// minutes, not 90 s) and must report a truncated body as an error instead of a
// hash nobody can match.
func TestSha256URLSlowAndTruncated(t *testing.T) {
	body := make([]byte, 4<<20)
	for i := range body {
		body[i] = byte(i % 251)
	}
	sum := sha256.Sum256(body)
	wantHex := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		// dribble it out so the read spans many writes, like a real download
		for i := 0; i < len(body); i += 256 << 10 {
			end := i + (256 << 10)
			if end > len(body) {
				end = len(body)
			}
			if _, err := w.Write(body[i:end]); err != nil {
				return
			}
			w.(http.Flusher).Flush()
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer srv.Close()

	got, code, err := sha256URL(context.Background(), &http.Client{Timeout: 15 * time.Minute}, srv.URL+"/pkg")
	if err != nil || code != http.StatusOK {
		t.Fatalf("slow read failed: code=%d err=%v", code, err)
	}
	if got != wantHex {
		t.Fatalf("hash mismatch: got %s want %s", got, wantHex)
	}

	// truncated: the server claims more than it sends
	short := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body[:1<<20])
	}))
	defer short.Close()

	if _, _, err := sha256URL(context.Background(), &http.Client{Timeout: 15 * time.Minute}, short.URL+"/pkg"); err == nil {
		t.Fatal("truncated body must return an error, not a hash")
	}
}
