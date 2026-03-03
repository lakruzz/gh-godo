package setstatus

import (
	"strings"
	"testing"
)

func TestRunValidation(t *testing.T) {
	tests := []struct {
		name        string
		sha         string
		state       string
		context     string
		description string
		targetURL   string
		repo        string
		wantErr     string
	}{
		{
			name:    "invalid sha - too short",
			sha:     "abc",
			state:   "success",
			context: "ci/build",
			repo:    "owner/repo",
			wantErr: "invalid commit SHA",
		},
		{
			name:    "invalid sha - non-hex",
			sha:     "xyz12345",
			state:   "success",
			context: "ci/build",
			repo:    "owner/repo",
			wantErr: "invalid commit SHA",
		},
		{
			name:    "invalid state",
			sha:     "abc1234",
			state:   "unknown",
			context: "ci/build",
			repo:    "owner/repo",
			wantErr: "invalid state",
		},
		{
			name:    "invalid repo format",
			sha:     "abc1234",
			state:   "success",
			context: "ci/build",
			repo:    "notavalidrepo",
			wantErr: "invalid repository format",
		},
		{
			name:      "invalid target url with null byte",
			sha:       "abc1234",
			state:     "success",
			context:   "ci/build",
			repo:      "owner/repo",
			targetURL: "http://example.com/\x00path",
			wantErr:   "invalid target URL",
		},
		{
			name:      "invalid target url with newline",
			sha:       "abc1234",
			state:     "success",
			context:   "ci/build",
			repo:      "owner/repo",
			targetURL: "http://example.com/\npath",
			wantErr:   "invalid target URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.sha, tt.state, tt.context, tt.description, tt.targetURL, tt.repo)
			if err == nil {
				t.Errorf("Run() expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Run() error = %q, want error containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidStates(t *testing.T) {
	states := []string{"error", "failure", "pending", "success"}
	for _, s := range states {
		if !validStates[s] {
			t.Errorf("expected %q to be a valid state", s)
		}
	}
	if validStates["unknown"] {
		t.Error("expected 'unknown' to be an invalid state")
	}
}

func TestSHAPattern(t *testing.T) {
	valid := []string{
		"abc1234",
		"abc1234def5678901234567890abcdef12345678",
		"ABCDEF1234567",
	}
	invalid := []string{
		"abc",       // too short
		"xyz12345",  // non-hex
		"",          // empty
		"abc1234 ",  // trailing space
		"abc1234\n", // newline
	}

	for _, sha := range valid {
		if !shaPattern.MatchString(sha) {
			t.Errorf("expected SHA %q to be valid", sha)
		}
	}
	for _, sha := range invalid {
		if shaPattern.MatchString(sha) {
			t.Errorf("expected SHA %q to be invalid", sha)
		}
	}
}

func TestRepoPattern(t *testing.T) {
	valid := []string{
		"owner/repo",
		"my-org/my-repo",
		"user123/project.name",
	}
	invalid := []string{
		"notavalidrepo",
		"owner/",
		"/repo",
		"owner/repo/extra",
		"",
	}

	for _, repo := range valid {
		if !repoPattern.MatchString(repo) {
			t.Errorf("expected repo %q to be valid", repo)
		}
	}
	for _, repo := range invalid {
		if repoPattern.MatchString(repo) {
			t.Errorf("expected repo %q to be invalid", repo)
		}
	}
}
