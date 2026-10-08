package publicframe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	var status atomic.Int32
	status.Store(http.StatusOK)
	var conditional atomic.Int32
	stacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/public-light.jpg" || r.URL.Query().Get("object") != "Orion" {
			http.NotFound(w, r)
			return
		}
		switch s := int(status.Load()); s {
		case http.StatusOK:
			if r.Header.Get("If-None-Match") == `"abc123"` {
				conditional.Add(1)
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("ETag", `"abc123"`)
			w.Header().Set("Last-Modified", "Wed, 07 Oct 2026 11:22:11 GMT")
			_, _ = w.Write([]byte("jpeg"))
		case -1:
			time.Sleep(200 * time.Millisecond)
		default:
			w.WriteHeader(s)
		}
	}))
	defer stacker.Close()
	f := New(stacker.URL, 100*time.Millisecond)
	ctx := context.Background()

	if err := f.Fetch(ctx, "Orion"); err != nil {
		t.Fatal(err)
	}
	fr, ok := f.Current("Orion")
	if !ok || string(fr.Body) != "jpeg" || fr.Version() != "abc123" {
		t.Fatalf("after fetch: %+v %v", fr, ok)
	}
	if _, ok := f.Current("M 31"); ok {
		t.Error("frame shown for another target")
	}

	// Unchanged: a conditional request, the frame kept.
	if err := f.Fetch(ctx, "Orion"); err != nil || conditional.Load() != 1 {
		t.Fatalf("refetch: %v, %d conditional", err, conditional.Load())
	}
	if _, ok := f.Current("Orion"); !ok {
		t.Error("frame dropped on 304")
	}

	// Errors and timeouts keep the frame shown.
	for _, s := range []int32{http.StatusInternalServerError, -1} {
		status.Store(s)
		if err := f.Fetch(ctx, "Orion"); err == nil {
			t.Errorf("status %d: no error", s)
		}
		if _, ok := f.Current("Orion"); !ok {
			t.Errorf("status %d: frame dropped", s)
		}
	}
	status.Store(http.StatusOK)

	// The stacker has none: nothing to show.
	status.Store(http.StatusNotFound)
	if err := f.Fetch(ctx, "Orion"); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.Current("Orion"); ok || f.Latest() != nil {
		t.Error("frame kept after 404")
	}
}

func TestServe(t *testing.T) {
	f := New("http://unused", time.Second)
	rec := httptest.NewRecorder()
	f.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/image.jpg", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("no frame: %d", rec.Code)
	}
	f.set(&Frame{Object: "Orion", ETag: `"abc"`, Body: []byte("jpeg")}, true)
	rec = httptest.NewRecorder()
	f.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/image.jpg?v=abc", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "jpeg" || rec.Header().Get("Content-Type") != "image/jpeg" ||
		rec.Header().Get("Cache-Control") != "public, max-age=86400, immutable" {
		t.Errorf("served %d %q %v", rec.Code, rec.Body.String(), rec.Header())
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/image.jpg?v=old", nil)
	req.Header.Set("If-None-Match", `"abc"`)
	rec = httptest.NewRecorder()
	f.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified || rec.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("conditional: %d %v", rec.Code, rec.Header())
	}
}
