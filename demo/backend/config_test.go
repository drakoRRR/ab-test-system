package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
		want    Config
	}{
		{
			name:    "missing SDK_API_KEY returns error",
			env:     map[string]string{"SDK_API_KEY": ""},
			wantErr: "SDK_API_KEY",
		},
		{
			name: "defaults applied when only API key is set",
			env:  map[string]string{"SDK_API_KEY": "test-key"},
			want: Config{
				APIKey:        "test-key",
				ServiceURL:    "http://localhost:8080/api/v1",
				Port:          "8081",
				ExperimentKey: "checkout-btn",
				FlagKey:       "new-checkout",
				StaticDir:     "../static",
			},
		},
		{
			name: "env vars override all defaults",
			env: map[string]string{
				"SDK_API_KEY":     "real-key",
				"SDK_SERVICE_URL": "http://api:8080/api/v1",
				"PORT":            "9000",
				"EXPERIMENT_KEY":  "btn-color",
				"FLAG_KEY":        "dark-mode",
				"STATIC_DIR":      "/var/www/static",
			},
			want: Config{
				APIKey:        "real-key",
				ServiceURL:    "http://api:8080/api/v1",
				Port:          "9000",
				ExperimentKey: "btn-color",
				FlagKey:       "dark-mode",
				StaticDir:     "/var/www/static",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := loadConfig()

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg)
		})
	}
}
