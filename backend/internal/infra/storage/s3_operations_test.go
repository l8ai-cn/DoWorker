package storage

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestGetInternalURL_UsesRunnerEndpoint verifies that GetInternalURL — used by
// the extension importer to hand runner pods a presigned skill/resource
// download link — targets STORAGE_RUNNER_ENDPOINT rather than the backend's
// own (often container-unreachable) STORAGE_ENDPOINT.
func TestGetInternalURL_UsesRunnerEndpoint(t *testing.T) {
	cfg := S3Config{
		Endpoint:       "localhost:10004",
		RunnerEndpoint: "host.docker.internal:10004",
		Region:         "us-east-1",
		Bucket:         "agentsmesh",
		AccessKey:      "minioadmin",
		SecretKey:      "minioadmin",
		UseSSL:         false,
		UsePathStyle:   true,
	}

	s, err := NewS3Storage(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	url, err := s.GetInternalURL(context.Background(), "skills/1/demo/abc.tar.gz", 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(url, "http://host.docker.internal:10004/") {
		t.Errorf("expected presigned URL to target runner endpoint, got %q", url)
	}
}

// TestGetInternalURL_FallsBackToInternalEndpoint verifies that when no
// RunnerEndpoint is configured, GetInternalURL falls back to the same
// endpoint the backend itself uses (e.g. production S3/OSS where runner
// pods and the backend share network reachability).
func TestGetInternalURL_FallsBackToInternalEndpoint(t *testing.T) {
	cfg := S3Config{
		Endpoint:     "localhost:10004",
		Region:       "us-east-1",
		Bucket:       "agentsmesh",
		AccessKey:    "minioadmin",
		SecretKey:    "minioadmin",
		UseSSL:       false,
		UsePathStyle: true,
	}

	s, err := NewS3Storage(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	url, err := s.GetInternalURL(context.Background(), "skills/1/demo/abc.tar.gz", 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(url, "http://localhost:10004/") {
		t.Errorf("expected presigned URL to fall back to internal endpoint, got %q", url)
	}
}
