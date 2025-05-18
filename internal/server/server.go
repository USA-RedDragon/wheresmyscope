package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/USA-RedDragon/wheresmyscope/internal/config"
	"github.com/USA-RedDragon/wheresmyscope/internal/mqtt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type contextKey uint8

const (
	MQTT_ContextKey contextKey = iota
)

func NewRouter(cfg *config.Config, mqttClient *mqtt.MQTT) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Pass the MQTT client to the context under the key "mqtt"
	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), MQTT_ContextKey, mqttClient))
			h.ServeHTTP(w, r)
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Lets get mqtt from the context
		mqtt, ok := r.Context().Value(MQTT_ContextKey).(*mqtt.MQTT)
		if !ok {
			http.Error(w, "MQTT client not found in context", http.StatusInternalServerError)
			return
		}

		// Output the current state of the scope in JSON format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false) // Disable HTML escaping
		if err := encoder.Encode(mqtt.GetState()); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
	})

	return r
}
