package workerdependencyartifact

import (
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type testHelper interface {
	Helper()
	Fatalf(string, ...any)
}

func validInput(t testHelper) Input {
	scope := control.Scope{
		OrganizationID:   7,
		OrganizationSlug: slugkit.MustNewForTest("team-alpha"),
		ActorID:          9,
	}
	tool := resolvedReference(resource.KindToolBinding, "video-tool")
	model := resolvedReference(resource.KindModelBinding, "video-model")
	compute := resolvedReference(resource.KindComputeTarget, "runner-pool")
	profile := resolvedReference(resource.KindResourceProfile, "standard")
	imageDigest := workerdependency.TextDigest("worker-image")
	definition := definitionSnapshot(t, false)
	agentfile := definition.AgentFile
	definitionHash := definition.DefinitionHash
	profilePin := dependencyPin(profile, 41)
	document := workerdependency.Document{
		Version:        workerdependency.VersionV1,
		OrganizationID: scope.OrganizationID,
		Namespace:      scope.OrganizationSlug,
		Worker: workerdependency.Worker{
			WorkerType:             slugkit.MustNewForTest("codex-cli"),
			AdapterID:              slugkit.MustNewForTest("codex-app-server"),
			SpecVersion:            workerspec.VersionV1,
			DefinitionHash:         definitionHash,
			ModelManagedFields:     []string{},
			CredentialBundleFields: []string{},
			AgentfileSource:        agentfile,
			AgentfileSourceDigest:  workerdependency.TextDigest(agentfile),
		},
		Models: workerdependency.Models{
			Tools: []workerdependency.ToolModel{{
				Binding: referenceFromResolved(tool),
				Role:    slugkit.MustNewForTest("video-generation"),
				Model: workerdependency.Model{
					Pin:                dependencyPin(model, 31),
					ResourceRevision:   4,
					ConnectionID:       51,
					ConnectionRevision: 6,
					ProviderKey:        slugkit.MustNewForTest("doubao"),
					ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
					ModelID:            "seedance-1-0-pro",
					BaseURL:            "https://ark.example.com/api/v3",
					Modalities:         []airesource.Modality{airesource.ModalityVideo},
					Capabilities: []airesource.Capability{
						airesource.CapabilityVideoGeneration,
					},
				},
				Modality:   airesource.ModalityVideo,
				Capability: airesource.CapabilityVideoGeneration,
				Environment: workerdependency.ToolModelEnvironment{
					APIKeyTarget:  "VIDEO_API_KEY",
					BaseURLTarget: "VIDEO_BASE_URL",
					ModelIDTarget: "VIDEO_MODEL_ID",
				},
			}},
		},
		Skills:           []workerdependency.Skill{},
		KnowledgeBases:   []workerdependency.KnowledgeBase{},
		RuntimeBundles:   []workerdependency.RuntimeBundle{},
		SecretReferences: []workerdependency.SecretReference{},
		Placement: workerdependency.Placement{
			CatalogRevision: "catalog-v1",
			RuntimeImage: workerdependency.RuntimeImage{
				ID:        11,
				Reference: "registry.example.com/worker@" + imageDigest,
				Digest:    imageDigest,
			},
			ComputeTarget:   dependencyPin(compute, 21),
			ResourceProfile: &profilePin,
			Spec: workerspec.Placement{
				Policy: workerspec.PlacementPolicyExplicit,
				ComputeTarget: workerspec.ComputeTarget{
					ID: 21, Kind: workerspec.ComputeTargetKindRunnerPool,
				},
				DeploymentMode: workerspec.DeploymentModePooled,
				ResourceProfile: workerspec.ResourceProfile{
					ID: 41,
					Resources: workerspec.ResourceRequestsLimits{
						CPURequestMilliCPU: 200,
						CPULimitMilliCPU:   1000,
						MemoryRequestBytes: 256 << 20,
						MemoryLimitBytes:   1 << 30,
					},
				},
			},
		},
	}
	spec := workerspec.NewV1(
		workerspec.Runtime{
			WorkerType: workerspec.WorkerType{
				Slug:           document.Worker.WorkerType,
				DefinitionHash: document.Worker.DefinitionHash,
			},
			Image: workerspec.RuntimeImage{
				ID:     document.Placement.RuntimeImage.ID,
				Digest: document.Placement.RuntimeImage.Digest,
			},
			ToolModelBindings: []workerspec.ToolModelBinding{{
				Role:         document.Models.Tools[0].Role,
				ModelBinding: workerSpecModelBinding(document.Models.Tools[0].Model),
				Modality:     document.Models.Tools[0].Modality,
				Capability:   document.Models.Tools[0].Capability,
				Environment: workerspec.ToolModelEnvironment{
					APIKey:  document.Models.Tools[0].Environment.APIKeyTarget,
					BaseURL: document.Models.Tools[0].Environment.BaseURLTarget,
					ModelID: document.Models.Tools[0].Environment.ModelIDTarget,
				},
			}},
		},
		document.Placement.Spec,
		workerspec.TypeConfig{
			SchemaVersion:   1,
			Values:          map[string]any{},
			SecretRefs:      map[string]workerspec.SecretReference{},
			InteractionMode: workerspec.InteractionModePTY,
			AutomationLevel: workerspec.AutomationLevelInteractive,
		},
		workerspec.Workspace{
			SkillIDs:               []int64{},
			KnowledgeMounts:        []workerspec.KnowledgeMount{},
			EnvBundleIDs:           []workerspec.RuntimeEnvBundleID{},
			ConfigDocumentBindings: []workerspec.ConfigDocumentBinding{},
		},
		workerspec.Lifecycle{
			TerminationPolicy: workerspec.TerminationPolicyManual,
		},
		workerspec.Metadata{},
	)
	return Input{
		Scope:          scope,
		Definition:     definition,
		PlanReferences: []control.ResolvedReference{tool, compute, profile},
		WorkerSpec:     spec,
		Dependencies: ResolvedDependencies{
			ToolModels: []ToolModelResolution{{
				Binding: tool, Role: document.Models.Tools[0].Role,
				Model:      modelResolution(model, document.Models.Tools[0].Model),
				Modality:   document.Models.Tools[0].Modality,
				Capability: document.Models.Tools[0].Capability,
				Environment: ToolModelEnvironmentResolution{
					APIKeyTarget:  document.Models.Tools[0].Environment.APIKeyTarget,
					BaseURLTarget: document.Models.Tools[0].Environment.BaseURLTarget,
					ModelIDTarget: document.Models.Tools[0].Environment.ModelIDTarget,
				},
			}},
			Placement: PlacementResolution{
				CatalogRevision: document.Placement.CatalogRevision,
				RuntimeImageID:  document.Placement.RuntimeImage.ID,
				ImageReference:  document.Placement.RuntimeImage.Reference,
				ImageDigest:     document.Placement.RuntimeImage.Digest,
				ComputeTarget: resourceResolution(
					compute,
					document.Placement.ComputeTarget.DomainID,
				),
				ResourceProfile: &ResourceResolution{
					reference: profile,
					domainID:  document.Placement.ResourceProfile.DomainID,
				},
				Spec: document.Placement.Spec,
			},
		},
	}
}

func addWorkspaceDependencies(t testHelper, input *Input) {
	t.Helper()
	primaryRef := resolvedReference(resource.KindModelBinding, "primary-model")
	primary := workerdependency.Model{
		Pin:                dependencyPin(primaryRef, 32),
		ResourceRevision:   5,
		ConnectionID:       52,
		ConnectionRevision: 7,
		ProviderKey:        slugkit.MustNewForTest("openai"),
		ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
		ModelID:            "gpt-5",
		BaseURL:            "https://api.example.com/v1",
		Modalities:         []airesource.Modality{airesource.ModalityChat},
		Capabilities:       []airesource.Capability{airesource.CapabilityTextGeneration},
	}
	primaryResolution := modelResolution(primaryRef, primary)
	input.Dependencies.PrimaryModel = &primaryResolution
	input.WorkerSpec.Runtime.ModelBinding = workerSpecModelBinding(primary)
	appendPlanReference(input, primaryRef)

	repositoryRef := resolvedReference(resource.KindRepository, "agents-mesh")
	repositoryID := int64(61)
	input.Dependencies.Repository = &RepositoryResolution{
		ResourceResolution: resourceResolution(repositoryRef, repositoryID),
		HTTPCloneURL:       "https://git.example.com/agents-mesh.git",
		Branch:             "main",
		CommitSHA:          strings.Repeat("b", 40),
		CredentialType:     workerdependency.RepositoryCredentialTypeNone,
	}
	input.WorkerSpec.Workspace.RepositoryID = &repositoryID
	input.WorkerSpec.Workspace.Branch = "main"
	appendPlanReference(input, repositoryRef)

	skillRef := resolvedReference(resource.KindSkill, "review")
	input.Dependencies.Skills = []SkillResolution{{
		ResourceResolution: resourceResolution(skillRef, 71),
		Slug:               slugkit.MustNewForTest("review"),
		Version:            2,
		ContentDigest:      workerdependency.TextDigest("review-package"),
		StorageKey:         "skills/review.tgz",
		PackageSize:        1024,
	}}
	input.WorkerSpec.Workspace.SkillIDs = []int64{71}
	appendPlanReference(input, skillRef)

	knowledgeRef := resolvedReference(resource.KindKnowledgeBase, "engineering-docs")
	input.Dependencies.KnowledgeBases = []KnowledgeBaseResolution{{
		ResourceResolution: resourceResolution(knowledgeRef, 81),
		Slug:               slugkit.MustNewForTest("engineering-docs"),
		HTTPCloneURL:       "https://git.example.com/knowledge/engineering-docs.git",
		Branch:             "main",
		CommitSHA:          strings.Repeat("c", 40),
		Mode:               workerspec.KnowledgeMountReadOnly,
	}}
	input.WorkerSpec.Workspace.KnowledgeMounts = []workerspec.KnowledgeMount{{
		KnowledgeBaseID: 81,
		Mode:            workerspec.KnowledgeMountReadOnly,
	}}
	appendPlanReference(input, knowledgeRef)

	runtimeRef := resolvedReference(resource.KindEnvironmentBundle, "runtime")
	runtimeValues := []workerdependency.RuntimeValue{{Name: "LOG_LEVEL", Value: "debug"}}
	runtimeDigest, err := workerdependency.DigestRuntimeValues(runtimeValues)
	if err != nil {
		t.Fatalf("digest runtime values: %v", err)
	}
	configRef := resolvedReference(resource.KindEnvironmentBundle, "config")
	configValues := []workerdependency.RuntimeValue{{
		Name: envbundle.ConfigJSONDataKey, Value: `{}`,
	}}
	configDigest, err := workerdependency.DigestRuntimeValues(configValues)
	if err != nil {
		t.Fatalf("digest config values: %v", err)
	}
	input.Dependencies.RuntimeBundles = []RuntimeBundleResolution{
		{
			ResourceResolution: resourceResolution(runtimeRef, 91),
			Kind:               envbundle.KindRuntime, ContentDigest: runtimeDigest,
			Values: runtimeValueResolutions(runtimeValues),
		},
		{
			ResourceResolution: resourceResolution(configRef, 92),
			Kind:               envbundle.KindConfig, ContentDigest: configDigest,
			Values: runtimeValueResolutions(configValues),
			ConfigDocument: &ConfigDocumentResolution{
				ID: "settings", Format: "json", TargetPath: ".codex/settings.json",
			},
		},
	}
	input.WorkerSpec.Workspace.EnvBundleIDs = []workerspec.RuntimeEnvBundleID{91}
	input.WorkerSpec.Workspace.ConfigDocumentBindings =
		[]workerspec.ConfigDocumentBinding{{
			DocumentID: "settings", ConfigBundleID: 92,
		}}
	appendPlanReference(input, runtimeRef)
	appendPlanReference(input, configRef)

	secretRef := resolvedReference(resource.KindEnvironmentBundle, "credentials")
	input.Dependencies.SecretReferences = []SecretReferenceResolution{{
		ResourceResolution: resourceResolution(secretRef, 93),
		Field:              "CURSOR_API_KEY",
		BundleKey:          "CURSOR_API_KEY", OwnerScope: envbundle.OwnerScopeOrg,
		OwnerID: input.Scope.OrganizationID,
	}}
	input.WorkerSpec.TypeConfig.SecretRefs["CURSOR_API_KEY"] =
		workerspec.SecretReference{
			Kind: slugkit.MustNewForTest("env-bundle"),
			ID:   93,
		}
	appendPlanReference(input, secretRef)
	input.Definition = definitionSnapshot(t, true)
	input.WorkerSpec.Runtime.WorkerType.DefinitionHash =
		input.Definition.DefinitionHash
}

func appendPlanReference(input *Input, reference control.ResolvedReference) {
	input.PlanReferences = append(input.PlanReferences, reference)
}
