// Package setstatus provides functionality to set GitHub commit statuses.
package setstatus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// validStates contains the allowed values for commit status state.
var validStates = map[string]bool{
	"error":   true,
	"failure": true,
	"pending": true,
	"success": true,
}

// shaPattern validates a git commit SHA (7–40 hex characters).
var shaPattern = regexp.MustCompile(`^[a-fA-F0-9]{7,40}$`)

// repoPattern validates owner/repo format.
var repoPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

// statusRequest represents the JSON body for the GitHub commit status API.
type statusRequest struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url,omitempty"`
	Description string `json:"description,omitempty"`
	Context     string `json:"context,omitempty"`
}

// Run sets a commit status via the GitHub API.
// sha is the commit SHA, state is one of error/failure/pending/success,
// context is a label for the status, description is a short human-readable message,
// targetURL is a URL associated with the status, and repo is owner/repo (optional).
func Run(sha, state, context, description, targetURL, repo string) error {
	// Validate SHA
	if !shaPattern.MatchString(sha) {
		return fmt.Errorf("invalid commit SHA %q: must be 7–40 hexadecimal characters", sha)
	}

	// Validate state
	if !validStates[state] {
		return fmt.Errorf("invalid state %q: must be one of error, failure, pending, success", state)
	}

	// Validate repo if provided
	if repo != "" && !repoPattern.MatchString(repo) {
		return fmt.Errorf("invalid repository format %q: must be owner/repo", repo)
	}

	// Validate targetURL if provided - basic check for prohibited characters
	if strings.ContainsAny(targetURL, "\x00\n\r") {
		return fmt.Errorf("invalid target URL: contains prohibited characters")
	}

	// Resolve repo if not provided
	if repo == "" {
		var err error
		repo, err = getCurrentRepo()
		if err != nil {
			return fmt.Errorf("could not determine repository: %w (use --repo to specify)", err)
		}
	}

	return postStatus(repo, sha, state, context, description, targetURL)
}

// getCurrentRepo returns the current repository in owner/repo format using gh CLI.
func getCurrentRepo() (string, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh repo view failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("gh repo view failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// postStatus calls the GitHub API to create a commit status.
func postStatus(repo, sha, state, context, description, targetURL string) error {
	req := statusRequest{
		State:       state,
		Context:     context,
		Description: description,
		TargetURL:   targetURL,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	endpoint := fmt.Sprintf("repos/%s/statuses/%s", repo, sha)
	args := []string{"api", "-X", "POST", endpoint, "--input", "-"}

	cmd := exec.Command("gh", args...)
	cmd.Stdin = bytes.NewReader(body)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set commit status: %w\nStderr: %s", err, stderr.String())
	}

	fmt.Printf("✅ Commit status set: %s (%s)\n", state, context)
	return nil
}
