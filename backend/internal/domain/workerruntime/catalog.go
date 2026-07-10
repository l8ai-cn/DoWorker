package workerruntime

import "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"

const DefaultCatalogRevision = "runtime-catalog-2026-07-10"

type CatalogRuntimeImage struct {
	ID              int64
	Slug            string
	Name            string
	Reference       string
	Digest          string
	WorkerTypeSlugs []string
	Enabled         bool
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
	return Catalog{
		revision: DefaultCatalogRevision,
		images: []CatalogRuntimeImage{
			{
				ID:              1,
				Slug:            "codex-cli-stable",
				Name:            "Codex CLI",
				Reference:       "repo.aiedulab.cn:8443/agentsmesh/runner-codex-cli@sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1",
				Digest:          "sha256:963c99fb047c0a4fed518eb9949e805fd31329a8395526fbb1fe34d8254ebea1",
				WorkerTypeSlugs: []string{"codex-cli"},
				Enabled:         true,
			},
			{
				ID:              2,
				Slug:            "claude-code-stable",
				Name:            "Claude Code",
				Reference:       "repo.aiedulab.cn:8443/agentsmesh/runner-claude-code@sha256:a9a02976dec14907be8eb6a7f68cd1adc5158099645244be733546b0f3e7041f",
				Digest:          "sha256:a9a02976dec14907be8eb6a7f68cd1adc5158099645244be733546b0f3e7041f",
				WorkerTypeSlugs: []string{"claude-code"},
				Enabled:         true,
			},
			{
				ID:              3,
				Slug:            "gemini-cli-stable",
				Name:            "Gemini CLI",
				Reference:       "repo.aiedulab.cn:8443/agentsmesh/runner-gemini-cli@sha256:852dba55bcc3213c72a7ee94e9c2da29a44e2ba0d5a9c0a8c15fea5adb8c6cd4",
				Digest:          "sha256:852dba55bcc3213c72a7ee94e9c2da29a44e2ba0d5a9c0a8c15fea5adb8c6cd4",
				WorkerTypeSlugs: []string{"gemini-cli"},
				Enabled:         true,
			},
		},
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
					CPURequestMilliCPU: 200,
					CPULimitMilliCPU:   1000,
					MemoryRequestBytes: 256 << 20,
					MemoryLimitBytes:   1 << 30,
				},
				Enabled: true,
			},
			{
				ID:   2,
				Slug: "large",
				Name: "Large",
				Resources: workerspec.ResourceRequestsLimits{
					CPURequestMilliCPU: 1000,
					CPULimitMilliCPU:   2000,
					MemoryRequestBytes: 1 << 30,
					MemoryLimitBytes:   4 << 30,
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
