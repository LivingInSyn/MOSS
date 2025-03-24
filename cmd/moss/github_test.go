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

	// Clear environment variables we'll be testing
	os.Unsetenv("GITHUB_PAT_testorg_CLOUD")
	os.Unsetenv("GITHUB_PAT_testorg_ONPREM")
	os.Unsetenv("GITLAB_PAT_testorg_CLOUD")

	// Set up test environment variables
	os.Setenv("GITHUB_PAT_testorg_CLOUD", "github-cloud-token")
	os.Setenv("GITHUB_PAT_testorg_ONPREM", "github-onprem-token")
	os.Setenv("GITLAB_PAT_testorg_CLOUD", "gitlab-cloud-token")

	// Define test cases
	tests := []struct {
		name     string
		provider string
		org      OrgConfig
		want     string
	}{
		{
			name:     "GitHub cloud organization",
			provider: "github",
			org:      OrgConfig{Name: "testorg", Type: "cloud"},
			want:     "github-cloud-token",
		},
		{
			name:     "GitHub onprem organization",
			provider: "github",
			org:      OrgConfig{Name: "testorg", Type: "onprem"},
			want:     "github-onprem-token",
		},
		{
			name:     "GitLab cloud organization",
			provider: "gitlab",
			org:      OrgConfig{Name: "testorg", Type: "cloud"},
			want:     "gitlab-cloud-token",
		},
		{
			name:     "Missing token",
			provider: "bitbucket",
			org:      OrgConfig{Name: "testorg", Type: "cloud"},
			want:     "",
		},
		{
			name:     "Case insensitive provider",
			provider: "Github", // Mixed case
			org:      OrgConfig{Name: "testorg", Type: "cloud"},
			want:     "github-cloud-token",
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
