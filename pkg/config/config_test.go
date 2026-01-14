package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool //nolint:govet // test struct
	}{
		{
			name: "valid configuration",
			env: map[string]string{
				"POD_LABELS":      "app=test",
				"TRAEFIK_API_URL": "http://localhost:8080/api",
				"POD_NAMESPACE":   "default",
			},
			wantErr: false,
		},
		{
			name: "missing POD_LABELS",
			env: map[string]string{
				"TRAEFIK_API_URL": "http://localhost:8080/api",
				"POD_NAMESPACE":   "default",
			},
			wantErr: true,
		},
		{
			name: "missing TRAEFIK_API_URL",
			env: map[string]string{
				"POD_LABELS":    "app=test",
				"POD_NAMESPACE": "default",
			},
			wantErr: true,
		},
		{
			name: "invalid UPDATE_INTERVAL",
			env: map[string]string{
				"POD_LABELS":      "app=test",
				"TRAEFIK_API_URL": "http://localhost:8080/api",
				"POD_NAMESPACE":   "default",
				"UPDATE_INTERVAL": "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid UPDATE_INTERVAL",
			env: map[string]string{
				"POD_LABELS":      "app=test",
				"TRAEFIK_API_URL": "http://localhost:8080/api",
				"POD_NAMESPACE":   "default",
				"UPDATE_INTERVAL": "5s",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			cfg, err := LoadFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if cfg.PodLabels != tt.env["POD_LABELS"] {
					t.Errorf("PodLabels = %v, want %v", cfg.PodLabels, tt.env["POD_LABELS"])
				}
				if cfg.TraefikAPIURL != tt.env["TRAEFIK_API_URL"] {
					t.Errorf("TraefikAPIURL = %v, want %v", cfg.TraefikAPIURL, tt.env["TRAEFIK_API_URL"])
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool //nolint:govet // test struct
	}{
		{
			name: "valid config",
			cfg: &Config{
				PodLabels:      "app=test",
				TraefikAPIURL:  "http://localhost:8080/api",
				PodNamespace:   "default",
				BackendPort:    3333,
				UpdateInterval: time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing PodLabels",
			cfg: &Config{
				TraefikAPIURL:  "http://localhost:8080/api",
				PodNamespace:   "default",
				BackendPort:    3333,
				UpdateInterval: time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid BackendPort",
			cfg: &Config{
				PodLabels:      "app=test",
				TraefikAPIURL:  "http://localhost:8080/api",
				PodNamespace:   "default",
				BackendPort:    99999,
				UpdateInterval: time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid UpdateInterval",
			cfg: &Config{
				PodLabels:      "app=test",
				TraefikAPIURL:  "http://localhost:8080/api",
				PodNamespace:   "default",
				BackendPort:    3333,
				UpdateInterval: 500 * time.Millisecond,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
