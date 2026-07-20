package workspace

import "testing"

// --- Test extractRepoName ---

func TestExtractRepoNameSSH(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{"git@github.com:user/repo.git", "user-repo"},
		{"git@github.com:org/project.git", "org-project"},
		{"git@gitlab.com:team/service.git", "team-service"},
	}

	for _, tc := range testCases {
		result := extractRepoName(tc.url)
		if result != tc.expected {
			t.Errorf("extractRepoName(%s): got %v, want %v", tc.url, result, tc.expected)
		}
	}
}

func TestExtractRepoNameHTTPS(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "user-repo"},
		{"https://github.com/user/repo", "user-repo"},
		{"https://gitlab.com/org/project.git", "org-project"},
	}

	for _, tc := range testCases {
		result := extractRepoName(tc.url)
		if result != tc.expected {
			t.Errorf("extractRepoName(%s): got %v, want %v", tc.url, result, tc.expected)
		}
	}
}

func TestExtractRepoNameLocalPaths(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"/tmp/owner/repo.git", "owner-repo"},
		{`C:\Users\runner\origin.git`, "runner-origin"},
		{"C:/Users/runner/origin.git", "runner-origin"},
		{`\\server\share\origin.git`, "share-origin"},
	}

	for _, tc := range testCases {
		result := extractRepoName(tc.path)
		if result != tc.expected {
			t.Errorf("extractRepoName(%s): got %v, want %v", tc.path, result, tc.expected)
		}
	}
}

func TestExtractRepoNameInvalid(t *testing.T) {
	result := extractRepoName("")
	if result != "" {
		t.Errorf("extractRepoName(empty): got %v, want empty", result)
	}
}

func TestExtractRepoNameSinglePart(t *testing.T) {
	result := extractRepoName("repo")
	if result != "" {
		t.Errorf("extractRepoName(repo): got %v, want empty", result)
	}
}

func TestExtractRepoNameSSHVariants(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{"git@github.com:user/repo", "user-repo"},
		{"git@bitbucket.org:team/project.git", "team-project"},
	}

	for _, tc := range testCases {
		result := extractRepoName(tc.url)
		if result != tc.expected {
			t.Errorf("extractRepoName(%s): got %v, want %v", tc.url, result, tc.expected)
		}
	}
}

func BenchmarkExtractRepoName(b *testing.B) {
	urls := []string{
		"git@github.com:user/repo.git",
		"https://github.com/user/repo.git",
		"https://github.com/org/project",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractRepoName(urls[i%len(urls)])
	}
}
