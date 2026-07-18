package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestInternalPresignPutURLSignsContentLength(t *testing.T) {
	s, err := NewS3Storage(S3Config{
		Endpoint:       "localhost:10004",
		RunnerEndpoint: "host.docker.internal:10004",
		Region:         "us-east-1",
		Bucket:         "agentsmesh",
		AccessKey:      "minioadmin",
		SecretKey:      "minioadmin",
		UsePathStyle:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	signed, err := s.InternalPresignPutURL(
		context.Background(),
		"workspace-artifacts/video.mp4",
		"video/mp4",
		32<<20,
		5*time.Minute,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := url.Parse(signed)
	if err != nil {
		t.Fatalf("parse presigned URL: %v", err)
	}
	if !strings.Contains(parsed.Query().Get("X-Amz-SignedHeaders"), "content-length") {
		t.Fatalf("expected content-length to be signed, got %q", signed)
	}
	if parsed.Host != "host.docker.internal:10004" {
		t.Fatalf("expected runner-reachable upload URL, got %q", signed)
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

func TestS3ExistsReturnsNonNotFoundHeadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)
	s, err := NewS3Storage(S3Config{
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
		Region:       "us-east-1",
		Bucket:       "agentsmesh",
		AccessKey:    "test",
		SecretKey:    "test",
		UsePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	exists, err := s.Exists(context.Background(), "skills/direct/test/package.tar.gz")

	if err == nil {
		t.Fatal("expected HeadObject authorization error")
	}
	if exists {
		t.Fatal("unauthorized object must not be reported as existing")
	}
}

func TestS3ExistsReturnsFalseForNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)
	s, err := NewS3Storage(S3Config{
		Endpoint:     strings.TrimPrefix(server.URL, "http://"),
		Region:       "us-east-1",
		Bucket:       "agentsmesh",
		AccessKey:    "test",
		SecretKey:    "test",
		UsePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	exists, err := s.Exists(context.Background(), "skills/direct/test/missing.tar.gz")

	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("missing object reported as existing")
	}
}
