package coordinator

import (
	"fmt"
	"strings"

	workerruntime "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerruntime"
)

func validateManagedRunnerImages(
	catalog workerruntime.Catalog,
	formalWorkerSlugs []string,
	images map[string]string,
) error {
	for _, slug := range formalWorkerSlugs {
		reference, released := releasedReference(catalog, slug)
		configured, configuredForWorker := images[slug]
		if !configuredForWorker {
			if released {
				return fmt.Errorf("coordinator: released worker %q has no runtime image", slug)
			}
			continue
		}
		if !released {
			return fmt.Errorf("coordinator: worker %q is not released for managed runtime", slug)
		}
		if strings.TrimSpace(configured) != reference {
			return fmt.Errorf("coordinator: worker %q image does not match runtime catalog", slug)
		}
	}
	for slug, reference := range images {
		if !hasImmutableDigest(reference) {
			return fmt.Errorf("coordinator: worker %q image must use an immutable sha256 digest", slug)
		}
	}
	return nil
}

func releasedReference(catalog workerruntime.Catalog, workerSlug string) (string, bool) {
	for _, image := range catalog.ImagesFor(workerSlug) {
		if image.Enabled {
			return image.Reference, true
		}
	}
	return "", false
}

func hasImmutableDigest(reference string) bool {
	_, digest, ok := strings.Cut(strings.TrimSpace(reference), "@")
	if !ok || len(digest) != len("sha256:")+64 || !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	for _, character := range digest[len("sha256:"):] {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
