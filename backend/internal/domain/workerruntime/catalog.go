package workerruntime

import "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"

type CatalogRuntimeImage struct {
	ID              int64    `json:"id"`
	Slug            string   `json:"slug"`
	Name            string   `json:"name"`
	Reference       string   `json:"reference"`
	Digest          string   `json:"digest"`
	WorkerTypeSlugs []string `json:"worker_type_slugs"`
	Enabled         bool     `json:"enabled"`
}

type CatalogComputeTarget struct {
	ID                int64
	Slug              string
	Name              string
	Kind              workerspec.ComputeTargetKind
	SupportsPooled    bool
	SupportsDedicated bool
	Enabled           bool
	DisabledReason    string
}

type CatalogResourceProfile struct {
	ID        int64
	Slug      string
	Name      string
	Resources workerspec.ResourceRequestsLimits
	Enabled   bool
}

type Catalog struct {
	revision string
	images   []CatalogRuntimeImage
	targets  []CatalogComputeTarget
	profiles []CatalogResourceProfile
}

func DefaultCatalog() Catalog {
	return catalogFromLock(loadRuntimeCatalogLock())
}

func catalogFromLock(lock runtimeCatalogLock) Catalog {
	images := make([]CatalogRuntimeImage, len(lock.Images))
	for index, image := range lock.Images {
		images[index] = cloneCatalogImage(image)
	}
	return Catalog{
		revision: lock.Revision,
		images:   images,
		targets: []CatalogComputeTarget{
			{
				ID:             1,
				Slug:           "organization-runner-pool",
				Name:           "Organization runner pool",
				Kind:           workerspec.ComputeTargetKindRunnerPool,
				SupportsPooled: true,
				Enabled:        true,
			},
			{
				ID:                2,
				Slug:              "managed-kubernetes",
				Name:              "Managed Kubernetes",
				Kind:              workerspec.ComputeTargetKindKubernetes,
				SupportsDedicated: true,
				DisabledReason:    "Dedicated managed Kubernetes provisioning is not configured",
			},
		},
		profiles: []CatalogResourceProfile{
			{
				ID:   1,
				Slug: "standard",
				Name: "Standard",
				Resources: workerspec.ResourceRequestsLimits{
					CPURequestMilliCPU:  200,
					CPULimitMilliCPU:    1000,
					MemoryRequestBytes:  256 << 20,
					MemoryLimitBytes:    1 << 30,
					StorageRequestBytes: 10 << 30,
					StorageLimitBytes:   10 << 30,
				},
				Enabled: true,
			},
			{
				ID:   2,
				Slug: "large",
				Name: "Large",
				Resources: workerspec.ResourceRequestsLimits{
					CPURequestMilliCPU:  1000,
					CPULimitMilliCPU:    2000,
					MemoryRequestBytes:  1 << 30,
					MemoryLimitBytes:    4 << 30,
					StorageRequestBytes: 50 << 30,
					StorageLimitBytes:   50 << 30,
				},
				Enabled: true,
			},
		},
	}
}

func (catalog Catalog) Revision() string { return catalog.revision }

func (catalog Catalog) ImagesFor(workerTypeSlug string) []CatalogRuntimeImage {
	var matches []CatalogRuntimeImage
	for _, image := range catalog.images {
		if containsWorkerType(image.WorkerTypeSlugs, workerTypeSlug) {
			matches = append(matches, cloneCatalogImage(image))
		}
	}
	return matches
}

func (catalog Catalog) Images() []CatalogRuntimeImage {
	images := make([]CatalogRuntimeImage, len(catalog.images))
	for index, image := range catalog.images {
		images[index] = cloneCatalogImage(image)
	}
	return images
}

func (catalog Catalog) Target(id int64) *CatalogComputeTarget {
	for _, target := range catalog.targets {
		if target.ID == id {
			copy := target
			return &copy
		}
	}
	return nil
}

func (catalog Catalog) Targets() []CatalogComputeTarget {
	return append([]CatalogComputeTarget{}, catalog.targets...)
}

func (catalog Catalog) Profile(id int64) *CatalogResourceProfile {
	for _, profile := range catalog.profiles {
		if profile.ID == id {
			copy := cloneCatalogProfile(profile)
			return &copy
		}
	}
	return nil
}

func (catalog Catalog) Profiles() []CatalogResourceProfile {
	profiles := make([]CatalogResourceProfile, len(catalog.profiles))
	for index, profile := range catalog.profiles {
		profiles[index] = cloneCatalogProfile(profile)
	}
	return profiles
}

func containsWorkerType(workerTypes []string, wanted string) bool {
	for _, workerType := range workerTypes {
		if workerType == wanted {
			return true
		}
	}
	return false
}

func cloneCatalogImage(image CatalogRuntimeImage) CatalogRuntimeImage {
	image.WorkerTypeSlugs = append([]string{}, image.WorkerTypeSlugs...)
	return image
}

func cloneCatalogProfile(profile CatalogResourceProfile) CatalogResourceProfile {
	if profile.Resources.GPURequest != nil {
		value := *profile.Resources.GPURequest
		profile.Resources.GPURequest = &value
	}
	if profile.Resources.GPULimit != nil {
		value := *profile.Resources.GPULimit
		profile.Resources.GPULimit = &value
	}
	return profile
}
