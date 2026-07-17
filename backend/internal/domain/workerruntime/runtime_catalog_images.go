package workerruntime

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed runtime_catalog.lock.json
var runtimeCatalogLockJSON []byte

type runtimeCatalogLock struct {
	SchemaVersion int                   `json:"schema_version"`
	Revision      string                `json:"revision"`
	Images        []CatalogRuntimeImage `json:"images"`
}

func loadRuntimeCatalogLock() runtimeCatalogLock {
	lock, err := parseRuntimeCatalogLock(runtimeCatalogLockJSON)
	if err != nil {
		panic(fmt.Sprintf("invalid embedded runtime catalog lock: %v", err))
	}
	return lock
}

func LoadCatalog(filePath string) (Catalog, error) {
	if strings.TrimSpace(filePath) == "" {
		return DefaultCatalog(), nil
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return Catalog{}, fmt.Errorf("read runtime catalog: %w", err)
	}
	lock, err := parseRuntimeCatalogLock(raw)
	if err != nil {
		return Catalog{}, err
	}
	return catalogFromLock(lock), nil
}

func parseRuntimeCatalogLock(raw []byte) (runtimeCatalogLock, error) {
	var lock runtimeCatalogLock
	if err := json.Unmarshal(raw, &lock); err != nil {
		return runtimeCatalogLock{}, fmt.Errorf("decode runtime catalog lock: %w", err)
	}
	if lock.SchemaVersion != 1 {
		return runtimeCatalogLock{}, fmt.Errorf("runtime catalog schema_version must be 1")
	}
	if strings.TrimSpace(lock.Revision) == "" {
		return runtimeCatalogLock{}, fmt.Errorf("runtime catalog revision is required")
	}
	if len(lock.Images) == 0 {
		return runtimeCatalogLock{}, fmt.Errorf("runtime catalog must declare an image")
	}
	ids := make(map[int64]struct{}, len(lock.Images))
	slugs := make(map[string]struct{}, len(lock.Images))
	workerTypes := make(map[string]struct{}, len(lock.Images))
	for index, image := range lock.Images {
		if image.ID <= 0 || strings.TrimSpace(image.Slug) == "" ||
			strings.TrimSpace(image.Name) == "" || len(image.WorkerTypeSlugs) == 0 {
			return runtimeCatalogLock{}, fmt.Errorf("runtime catalog image %d is incomplete", index)
		}
		if _, exists := ids[image.ID]; exists {
			return runtimeCatalogLock{}, fmt.Errorf("runtime catalog repeats image id %d", image.ID)
		}
		if _, exists := slugs[image.Slug]; exists {
			return runtimeCatalogLock{}, fmt.Errorf("runtime catalog repeats image slug %q", image.Slug)
		}
		digest, err := immutableImageDigest(image.Reference)
		if err != nil {
			return runtimeCatalogLock{}, err
		}
		if digest != image.Digest {
			return runtimeCatalogLock{}, fmt.Errorf(
				"runtime catalog image %q digest does not match reference",
				image.Slug,
			)
		}
		for _, workerType := range image.WorkerTypeSlugs {
			if strings.TrimSpace(workerType) == "" {
				return runtimeCatalogLock{}, fmt.Errorf(
					"runtime catalog image %q has empty worker type",
					image.Slug,
				)
			}
			if _, exists := workerTypes[workerType]; exists {
				return runtimeCatalogLock{}, fmt.Errorf(
					"runtime catalog repeats worker type %q",
					workerType,
				)
			}
			workerTypes[workerType] = struct{}{}
		}
		ids[image.ID] = struct{}{}
		slugs[image.Slug] = struct{}{}
	}
	return lock, nil
}

func DefaultCatalogRevision() string {
	return loadRuntimeCatalogLock().Revision
}

func immutableImageDigest(reference string) (string, error) {
	_, digest, ok := strings.Cut(strings.TrimSpace(reference), "@")
	if !ok || !validImageDigest(digest) {
		return "", fmt.Errorf(
			"runtime image reference %q must end with an immutable sha256 digest",
			reference,
		)
	}
	return digest, nil
}

func validImageDigest(digest string) bool {
	if len(digest) != len("sha256:")+64 || !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	for _, character := range digest[len("sha256:"):] {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
