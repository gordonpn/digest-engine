package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name, port, timeout string
		wantPort            int
		wantTimeout         time.Duration
		invalid             bool
	}{
		{"valid", "9090", "2s", 9090, 2 * time.Second, false},
		{"secret in port", "private-credential", "2s", 0, 0, true},
		{"port zero", "0", "2s", 0, 0, true},
		{"port too large", "65536", "2s", 0, 0, true},
		{"empty port", "", "2s", 0, 0, true},
		{"invalid timeout", "8080", "private-credential", 0, 0, true},
		{"zero timeout", "8080", "0s", 0, 0, true},
		{"negative timeout", "8080", "-1s", 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PORT", tc.port)
			t.Setenv("DIGEST_ENGINE_SHUTDOWN_TIMEOUT", tc.timeout)
			cfg, err := Load()
			if tc.invalid {
				if err == nil {
					t.Fatal("expected configuration error")
				}
				if strings.Contains(err.Error(), "private-credential") {
					t.Fatal("configuration value leaked")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Port != tc.wantPort || cfg.ShutdownTimeout != tc.wantTimeout {
				t.Fatalf("unexpected config: %+v", cfg)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{"PORT", "DIGEST_ENGINE_SHUTDOWN_TIMEOUT"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 || cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
