package repository

import "testing"

func TestGenerateCloneURLsCNBUsesHTTPSOnly(t *testing.T) {
	httpURL, sshURL := generateCloneURLs("cnb", "https://cnb.cool", "owner/repo")
	if httpURL != "https://cnb.cool/owner/repo" {
		t.Fatalf("httpURL = %q, want https://cnb.cool/owner/repo", httpURL)
	}
	if sshURL != "" {
		t.Fatalf("CNB does not support SSH, got sshURL %q", sshURL)
	}
}
