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
		{
			name:     "returns empty string when env var is empty",
			key:      "EMPTY_VAR",
			def:      "default",
			want:     "default",
			setValue: "",
		},
		{
			name:     "returns empty string as default",
			key:      "ANOTHER_VAR",
			def:      "",
			want:     "",
			setValue: "",
		},
		{
			name:     "returns value when default is empty",
			key:      "SET_VAR",
			def:      "",
			want:     "value",
			setValue: "value",
		},
		{
			name:     "handles whitespace in value",
			key:      "WHITESPACE_VAR",
			def:      "default",
			want:     "  has spaces  ",
			setValue: "  has spaces  ",
		},
		{
			name:     "handles special characters",
			key:      "SPECIAL_VAR",
			def:      "default",
			want:     "test:value:123",
			setValue: "test:value:123",
		},
		{
			name:     "handles numeric value",
			key:      "PORT_VAR",
			def:      "8080",
			want:     "9090",
			setValue: "9090",
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
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.def, got, tt.want)
			}
		})
	}
}

func TestGetEnvConcurrent(t *testing.T) {
	// Test that getEnv is safe for concurrent use
	done := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				getEnv("TEST_VAR", "default")
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestGetEnvWithLongValue(t *testing.T) {
	longBytes := make([]byte, 10000)
	for i := range longBytes {
		longBytes[i] = 'a'
	}
	longValue := string(longBytes)

	os.Setenv("LONG_VAR", longValue)
	defer os.Unsetenv("LONG_VAR")

	got := getEnv("LONG_VAR", "default")
	if got != longValue {
		t.Errorf("getEnv() with long value returned incorrect length")
	}
}
