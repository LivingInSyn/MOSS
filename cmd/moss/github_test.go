package main

import (
	"os"
	"strings"
	"testing"
)

func TestGetPat(t *testing.T) {
	// Save original environment variables to restore later
	originalEnv := make(map[string]string)
	for _, envVar := range os.Environ() {
		if pair := strings.SplitN(envVar, "=", 2); len(pair) == 2 {
			originalEnv[pair[0]] = pair[1]
		}
	}

	// Clear and set up test environment variables
	os.Clearenv()
	os.Setenv("GITHUB_PAT_testorg", "github-token-direct")
	os.Setenv("GITLAB_PAT_otherorg", "gitlab-token-direct")

	// Define test cases
	tests := []struct {
		name     string
		provider string
		org      OrgConfig
		want     string
	}{
		{
			name:     "Direct environment variable lookup",
			provider: "github",
			org:      OrgConfig{Name: "testorg"},
			want:     "github-token-direct",
		},
		{
			name:     "Different organization",
			provider: "gitlab",
			org:      OrgConfig{Name: "otherorg"},
			want:     "gitlab-token-direct",
		},
		{
			name:     "Non-existent token",
			provider: "github",
			org:      OrgConfig{Name: "nonexistent"},
			want:     "",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPat(tt.provider, tt.org)
			if got != tt.want {
				t.Errorf("getPat() = %v, want %v", got, tt.want)
			}
		})
	}

	// Restore original environment
	os.Clearenv()
	for k, v := range originalEnv {
		os.Setenv(k, v)
	}
}
