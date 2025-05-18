package config_test

import (
	"errors"
	"testing"

	"github.com/USA-RedDragon/configulator"
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

	defConfig, err := configulator.New[config.Config]().Default()
	if err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}

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

	defConfig, err := configulator.New[config.Config]().Default()
	if err != nil {
		t.Fatalf("failed to create default config: %v", err)
	}

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
