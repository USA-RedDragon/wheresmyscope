package config_test

import (
	"reflect"
	"testing"

	configulator "github.com/USA-RedDragon/configulator/v2"
	"github.com/USA-RedDragon/wheresmyscope/internal/config"
)

// TestProductionEnv sets every variable the home-cluster deployment
// (apps/wheresmyscope/values.yaml and its wheresmyscope Secret) passes,
// under the same names, and checks each lands in its field. A renamed tag
// or section that silently drops one of them fails here.
func TestProductionEnv(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"MQTT_PREFIX":              "observatory/mercury/target",
		"MQTT_BROKER":              "mqtt://192.168.254.11:1883",
		"MQTT_USERNAME":            "wheresmyscope",
		"MQTT_PASSWORD":            "dummy",
		"IMAGE_HEIGHT":             "480",
		"IMAGE_WIDTH":              "800",
		"PUBLIC_FRAME_STACKER_URL": "http://astro-stacker.astro-processing.svc.cluster.local:8080",
	}

	cfg, err := configulator.New(config.ConfigSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Separator: "_"}).
		WithEnviron(func(k string) (string, bool) { v, ok := env[k]; return v, ok }).
		Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	checks := []struct {
		env  string
		got  any
		want any
	}{
		{"MQTT_PREFIX", cfg.MQTT.Prefix, "observatory/mercury/target"},
		{"MQTT_BROKER", cfg.MQTT.Broker, "mqtt://192.168.254.11:1883"},
		{"MQTT_USERNAME", cfg.MQTT.Username, "wheresmyscope"},
		{"MQTT_PASSWORD", cfg.MQTT.Password, "dummy"},
		{"IMAGE_HEIGHT", cfg.Image.Height, 480},
		{"IMAGE_WIDTH", cfg.Image.Width, 800},
		{"PUBLIC_FRAME_STACKER_URL", cfg.PublicFrame.StackerURL, env["PUBLIC_FRAME_STACKER_URL"]},
	}
	if len(checks) != len(env) {
		t.Fatalf("%d checks for %d variables", len(checks), len(env))
	}
	for _, c := range checks {
		if _, ok := env[c.env]; !ok {
			t.Errorf("check for unset variable %s", c.env)
		}
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s: got %#v, want %#v", c.env, c.got, c.want)
		}
	}

	// Untouched fields keep their defaults.
	if cfg.PublicFrame.PublicURL != "https://wheresmyscope.mcswain.dev" ||
		cfg.PublicFrame.IntervalSeconds != 60 || cfg.PublicFrame.TimeoutSeconds != 10 {
		t.Errorf("public-frame defaults: %+v", cfg.PublicFrame)
	}
	if cfg.MQTT.ClientID != "wheresmyscope" || cfg.Port != 8080 ||
		!reflect.DeepEqual(cfg.CORSAllowedOrigins, []string{"https://*", "http://*"}) {
		t.Errorf("defaults: %+v", cfg)
	}
}

// TestV1SectionNamesIgnored pins the v1 name, built from the Go field
// name, as no longer read: PUBLICFRAME_STACKER_URL must not set the URL.
func TestV1SectionNamesIgnored(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"MQTT_BROKER":             "mqtt://localhost:1883",
		"PUBLICFRAME_STACKER_URL": "http://astro-stacker:8080",
	}
	cfg, err := configulator.New(config.ConfigSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Separator: "_"}).
		WithEnviron(func(k string) (string, bool) { v, ok := env[k]; return v, ok }).
		Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.PublicFrame.StackerURL != "" {
		t.Error("PUBLICFRAME_STACKER_URL was read; the section is PUBLIC_FRAME_")
	}
}

// TestLoadValidates checks Load still runs Validate, as v1 did.
func TestLoadValidates(t *testing.T) {
	t.Parallel()

	_, err := configulator.New(config.ConfigSchema()).
		WithEnvironmentVariables(&configulator.EnvironmentVariableOptions{Separator: "_"}).
		WithEnviron(func(string) (string, bool) { return "", false }).
		Load()
	if err == nil {
		t.Fatal("Load without an MQTT broker succeeded")
	}
}
