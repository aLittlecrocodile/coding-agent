package main

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		def      string
		want     string
		setValue string
	}{
		{
			name:     "returns existing env var",
			key:      "TEST_VAR",
			def:      "default",
			want:     "custom",
			setValue: "custom",
		},
		{
			name:     "returns default when not set",
			key:      "NON_EXISTENT_VAR",
			def:      "default",
			want:     "default",
			setValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				os.Setenv(tt.key, tt.setValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.def)
			if got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
