package agentpod_test

import (
	"strings"
	"testing"

	podv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/pod/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestCreatePodRequiresExactModelResourceField(t *testing.T) {
	create := podMessage(t, "CreatePodRequest")
	field := create.Fields().ByName("model_resource_id")
	if field == nil {
		t.Fatal("CreatePodRequest.model_resource_id missing")
	}
	if field.Number() != 19 || field.Kind() != protoreflect.Int64Kind ||
		!field.HasOptionalKeyword() {
		t.Fatalf("CreatePodRequest.model_resource_id has incompatible contract: %s", field)
	}
	for _, legacy := range []string{
		"credential" + "_profile_id",
		"virtual_api" + "_key_id",
		"model" + "_config_id",
	} {
		if create.Fields().ByName(protoreflect.Name(legacy)) != nil {
			t.Errorf("CreatePodRequest.%s must be removed", legacy)
		}
	}
}

func TestWorkerCreationAndPublishingContract(t *testing.T) {
	service := podv1.File_pod_v1_pod_proto.Services().ByName("PodService")
	if service == nil {
		t.Fatal("PodService descriptor missing")
	}
	for _, name := range []string{
		"ListWorkerCreateOptions",
		"PreflightWorker",
		"FillWorkerDraft",
		"DiscoverWorkerSkills",
		"PublishWorkerSkills",
	} {
		if service.Methods().ByName(protoreflect.Name(name)) == nil {
			t.Errorf("PodService method %s missing", name)
		}
	}

	pod := podMessage(t, "Pod")
	snapshot := pod.Fields().ByName("worker_spec_snapshot_id")
	if snapshot == nil || snapshot.Number() != 35 ||
		snapshot.Kind() != protoreflect.Int64Kind || !snapshot.HasOptionalKeyword() {
		t.Errorf("Pod.worker_spec_snapshot_id has incompatible contract: %v", snapshot)
	}

	create := podMessage(t, "CreatePodRequest")
	workerSpec := create.Fields().ByName("worker_spec")
	if workerSpec == nil || workerSpec.Number() != 20 ||
		workerSpec.Kind() != protoreflect.MessageKind ||
		workerSpec.Message().Name() != "WorkerSpecDraft" {
		t.Errorf("CreatePodRequest.worker_spec has incompatible contract: %v", workerSpec)
	}

	draft := podMessage(t, "WorkerSpecDraft")
	for _, field := range []podFieldContract{
		{"model_resource_id", 1, protoreflect.Int64Kind},
		{"worker_type_slug", 2, protoreflect.StringKind},
		{"runtime_image_id", 3, protoreflect.Int64Kind},
		{"placement_policy", 4, protoreflect.StringKind},
		{"compute_target_id", 5, protoreflect.Int64Kind},
		{"deployment_mode", 6, protoreflect.StringKind},
		{"resource_profile_id", 7, protoreflect.Int64Kind},
		{"type_schema_version", 8, protoreflect.Uint32Kind},
		{"type_config_values_json", 9, protoreflect.StringKind},
		{"secret_refs", 10, protoreflect.MessageKind},
		{"interaction_mode", 11, protoreflect.StringKind},
		{"automation_level", 12, protoreflect.StringKind},
		{"repository_id", 13, protoreflect.Int64Kind},
		{"branch", 14, protoreflect.StringKind},
		{"skill_ids", 15, protoreflect.Int64Kind},
		{"knowledge_mounts", 16, protoreflect.MessageKind},
		{"env_bundle_ids", 17, protoreflect.Int64Kind},
		{"instructions", 18, protoreflect.StringKind},
		{"initial_task", 19, protoreflect.StringKind},
		{"termination_policy", 20, protoreflect.StringKind},
		{"idle_timeout_minutes", 21, protoreflect.Uint32Kind},
		{"alias", 22, protoreflect.StringKind},
		{"source_expert_id", 23, protoreflect.Int64Kind},
		{"options_revision", 24, protoreflect.StringKind},
	} {
		assertPodField(t, draft, field)
	}

	for _, forbidden := range []string{
		"default_auth",
		"image_auth",
		"credential_value",
		"api_key",
		"token",
	} {
		for index := 0; index < draft.Fields().Len(); index++ {
			field := draft.Fields().Get(index)
			if strings.Contains(string(field.Name()), forbidden) {
				t.Errorf("WorkerSpecDraft.%s must not expose credential material", field.Name())
			}
		}
	}

	for _, messageName := range []string{
		"ListWorkerCreateOptionsRequest",
		"ListWorkerCreateOptionsResponse",
		"WorkerTypeOption",
		"WorkerRuntimeImageOption",
		"WorkerComputeTargetOption",
		"WorkerDeploymentModeOption",
		"WorkerResourceProfileOption",
		"PreflightWorkerRequest",
		"PreflightWorkerResponse",
		"WorkerPreflightIssue",
		"FillWorkerDraftRequest",
		"FillWorkerDraftResponse",
		"DiscoverWorkerSkillsRequest",
		"DiscoverWorkerSkillsResponse",
		"WorkerSkillCandidate",
		"PublishWorkerSkillsRequest",
		"PublishWorkerSkillsResponse",
		"PublishedWorkerSkill",
	} {
		podMessage(t, messageName)
	}
}

type podFieldContract struct {
	name   string
	number protoreflect.FieldNumber
	kind   protoreflect.Kind
}

func podMessage(t *testing.T, name string) protoreflect.MessageDescriptor {
	t.Helper()
	descriptor, err := protoregistry.GlobalFiles.FindDescriptorByName(
		protoreflect.FullName("proto.pod.v1." + name),
	)
	if err != nil {
		t.Fatalf("find %s: %v", name, err)
	}
	message, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		t.Fatalf("%s is not a message", name)
	}
	return message
}

func assertPodField(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	field podFieldContract,
) {
	t.Helper()
	descriptor := message.Fields().ByName(protoreflect.Name(field.name))
	if descriptor == nil {
		t.Errorf("%s.%s missing", message.Name(), field.name)
		return
	}
	if descriptor.Number() != field.number || descriptor.Kind() != field.kind {
		t.Errorf(
			"%s.%s = field %d %s, want %d %s",
			message.Name(),
			field.name,
			descriptor.Number(),
			descriptor.Kind(),
			field.number,
			field.kind,
		)
	}
}
