package config_test

import (
	"errors"
	"testing"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/wheresmyscope/internal/config"
)

func TestLogLevelConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logLevel config.LogLevel
		valid    bool
	}{
		{"debug level", config.LogLevelDebug, true},
		{"info level", config.LogLevelInfo, true},
		{"warn level", config.LogLevelWarn, true},
		{"error level", config.LogLevelError, true},
		{"invalid level", "invalid", false},
	}

	defConfig, err := configulator.New(config.ConfigSchema()).Default()
	if err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	// The broker has no default and is required.
	defConfig.MQTT.Broker = "mqtt://localhost:1883"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defConfig
			cfg.LogLevel = tt.logLevel
			err := cfg.Validate()
			if tt.valid {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			} else if !errors.Is(err, config.ErrInvalidLogLevel) {
				t.Errorf("Validate() error = %v, want %v", err, config.ErrInvalidLogLevel)
			}
		})
	}
}

func TestPortValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		port  int
		valid bool
	}{
		{"valid port", 8080, true},
		{"port too low", 0, false},
		{"port too high", 70000, false},
	}

	defConfig, err := configulator.New(config.ConfigSchema()).Default()
	if err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	// The broker has no default and is required.
	defConfig.MQTT.Broker = "mqtt://localhost:1883"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defConfig
			cfg.Port = tt.port
			err := cfg.Validate()
			if tt.valid {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			} else if !errors.Is(err, config.ErrInvalidPort) {
				t.Errorf("Validate() error = %v, want %v", err, config.ErrInvalidPort)
			}
		})
	}
}

func TestPublicFrameValidation(t *testing.T) {
	t.Parallel()

	defConfig, err := configulator.New(config.ConfigSchema()).Default()
	if err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}
	defConfig.MQTT.Broker = "mqtt://localhost:1883"
	defConfig.PublicFrame.StackerURL = "http://astro-stacker.astro-processing:8080"
	if err := defConfig.Validate(); err != nil {
		t.Errorf("defaults with a stacker URL: %v", err)
	}
	cfg := defConfig
	cfg.PublicFrame.TimeoutSeconds = 0
	if err := cfg.Validate(); !errors.Is(err, config.ErrPublicFrameTiming) {
		t.Errorf("zero timeout: %v", err)
	}
	cfg = defConfig
	cfg.PublicFrame.PublicURL = ""
	if err := cfg.Validate(); !errors.Is(err, config.ErrNoPublicURL) {
		t.Errorf("no public URL: %v", err)
	}
}
