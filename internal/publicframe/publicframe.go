// Package publicframe fetches the observatory's newest sub of a target, as
// the astro-stacker renders it for the public (small, stretched and
// watermarked), so the page can show it instead of a survey image.
//
// It polls in the background and keeps the frame in memory: page requests
// never reach the stacker, and the stacker is asked again only with the
// ETag of what is held, so an unchanged frame costs a 304.
package publicframe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// maxBytes bounds a frame; the stacker's are about 100 KB.
const maxBytes = 4 << 20

// Frame is one fetched frame.
type Frame struct {
	Object       string
	ETag         string
	LastModified string
	Body         []byte
}

// Version is a short form of the frame's ETag, for cache-busting URLs.
func (f *Frame) Version() string {
	v := strings.Trim(strings.TrimPrefix(f.ETag, "W/"), `"`)
	if len(v) > 16 {
		v = v[:16]
	}
	return v
}

// Fetcher holds the frame of the target being imaged.
type Fetcher struct {
	url    string
	client *http.Client

	mu    sync.RWMutex
	frame *Frame
	// ok is whether the last fetch for frame's object succeeded; after an
	// error the frame is kept, for a 304 to bring back, but not shown.
	ok bool
}

// New fetches from the stacker at base (e.g.
// http://astro-stacker.astro-processing:8080), each request bounded by
// timeout.
func New(base string, timeout time.Duration) *Fetcher {
	return &Fetcher{url: strings.TrimSuffix(base, "/") + "/api/v1/public-light.jpg", client: &http.Client{Timeout: timeout}}
}

// Current is the frame to show for object: only one fetched for it by a
// fetch that succeeded.
func (f *Fetcher) Current(object string) (*Frame, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if !f.ok || f.frame == nil || object == "" || f.frame.Object != object {
		return nil, false
	}
	return f.frame, true
}

// Latest is the last frame fetched, whatever its state, for serving a URL
// the page was already given.
func (f *Fetcher) Latest() *Frame {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.frame
}

// Fetch asks the stacker for object's frame, sending the ETag of the one
// held. A 404 means it has none; any other failure keeps the frame held.
func (f *Fetcher) Fetch(ctx context.Context, object string) error {
	f.mu.RLock()
	held := f.frame
	f.mu.RUnlock()
	if held != nil && held.Object != object {
		held = nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url+"?"+url.Values{"object": {object}}.Encode(), nil)
	if err != nil {
		return err
	}
	if held != nil {
		req.Header.Set("If-None-Match", held.ETag)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		f.fail()
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		if held == nil {
			f.fail()
			return errors.New("304 without a frame held")
		}
		f.set(held, true)
		return nil
	case http.StatusOK:
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
		if err != nil {
			f.fail()
			return err
		}
		if len(body) > maxBytes {
			f.fail()
			return fmt.Errorf("frame over %d bytes", maxBytes)
		}
		etag := resp.Header.Get("ETag")
		if etag == "" {
			f.fail()
			return errors.New("frame without an ETag")
		}
		f.set(&Frame{Object: object, ETag: etag, LastModified: resp.Header.Get("Last-Modified"), Body: body}, true)
		slog.Info("Fetched public frame", "object", object, "etag", etag, "bytes", len(body))
		return nil
	case http.StatusNotFound:
		f.set(nil, false)
		return nil
	default:
		f.fail()
		return fmt.Errorf("stacker answered %s", resp.Status)
	}
}

func (f *Fetcher) set(frame *Frame, ok bool) {
	f.mu.Lock()
	f.frame, f.ok = frame, ok
	f.mu.Unlock()
}

// fail keeps the frame held: the stacker restarting shouldn't swap the
// page back to the survey image.
func (f *Fetcher) fail() {}

// ServeHTTP serves the frame held, the bytes as fetched, with its ETag.
// The page's URL carries the frame's version, so it may be cached.
func (f *Fetcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	frame := f.Latest()
	if frame == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("ETag", frame.ETag)
	if frame.LastModified != "" {
		w.Header().Set("Last-Modified", frame.LastModified)
	}
	if v := r.URL.Query().Get("v"); v != "" && v == frame.Version() {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	if r.Header.Get("If-None-Match") == frame.ETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", fmt.Sprint(len(frame.Body)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(frame.Body)
	}
}
