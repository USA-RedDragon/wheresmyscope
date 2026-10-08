package mqtt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/USA-RedDragon/wheresmyscope/internal/config"
	"github.com/USA-RedDragon/wheresmyscope/internal/publicframe"
)

// TestImageURL checks the page is sent the observatory's frame only while it
// is imaging the target the frame is of, and the survey image otherwise.
func TestImageURL(t *testing.T) {
	stacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("object") != "Orion" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("ETag", `"0123456789abcdef0123"`)
		_, _ = w.Write([]byte("jpeg"))
	}))
	defer stacker.Close()
	cfg := &config.Config{MQTT: config.MQTT{Prefix: "p"}, Image: config.Image{Width: 800, Height: 480, FOV: 3.3, MinCut: 0.5, MaxCut: 99.5}}
	m := &MQTT{config: cfg, kick: make(chan struct{}, 1),
		frames: publicframe.New(stacker.URL, time.Second), publicURL: "https://wheresmyscope.example"}
	for topic, payload := range map[string]string{"p/name": "Orion", "p/ra_decimal": "5.588", "p/dec_decimal": "-5.39"} {
		m.applyState(topic, payload)
	}
	survey := m.GetState().ImageURL
	if !strings.HasPrefix(survey, "https://alaskybis.u-strasbg.fr/") {
		t.Fatalf("survey URL %q", survey)
	}
	if err := m.frames.Fetch(context.Background(), "Orion"); err != nil {
		t.Fatal(err)
	}
	if got := m.GetState().ImageURL; got != survey {
		t.Errorf("not imaging, image %q, want the survey", got)
	}
	m.applyState("p/available", "true")
	if got, want := m.GetState().ImageURL, "https://wheresmyscope.example/image.jpg?v=0123456789abcdef"; got != want {
		t.Errorf("imaging Orion, image %q, want %q", got, want)
	}
	m.applyState("p/name", "M 31")
	if got := m.GetState().ImageURL; strings.Contains(got, "wheresmyscope.example") {
		t.Errorf("target changed, still showing Orion's frame: %q", got)
	}
}
