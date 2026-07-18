package orchestrationresource

import (
	"bytes"
	"encoding/json"
)

func rejectWorkerTemplateRuntimeNulls(source []byte) error {
	var runtime map[string]json.RawMessage
	if err := json.Unmarshal(source, &runtime); err != nil {
		return err
	}
	if runtime == nil {
		return typedJSONNullError("spec.runtime")
	}
	if err := rejectNullJSONFields(runtime, "spec.runtime",
		"runtimeImageId",
		"placementPolicy",
		"computeTargetRef",
		"deploymentMode",
		"resourceProfileRef",
		"customResources",
	); err != nil {
		return err
	}
	if rawComputeTarget, exists := runtime["computeTargetRef"]; exists {
		if err := rejectReferenceJSONNulls(
			rawComputeTarget,
			"spec.runtime.computeTargetRef",
		); err != nil {
			return err
		}
	}
	if rawProfile, exists := runtime["resourceProfileRef"]; exists {
		if err := rejectReferenceJSONNulls(
			rawProfile,
			"spec.runtime.resourceProfileRef",
		); err != nil {
			return err
		}
	}

	rawResources, exists := runtime["customResources"]
	if !exists {
		return nil
	}
	var resources map[string]json.RawMessage
	if err := json.Unmarshal(rawResources, &resources); err != nil {
		return err
	}
	return rejectNullJSONFields(resources, "spec.runtime.customResources",
		"cpuRequestMilliCPU",
		"cpuLimitMilliCPU",
		"memoryRequestBytes",
		"memoryLimitBytes",
		"storageRequestBytes",
		"storageLimitBytes",
		"gpuRequest",
		"gpuLimit",
	)
}

func rejectReferenceJSONNulls(source json.RawMessage, path string) error {
	var reference map[string]json.RawMessage
	if err := json.Unmarshal(source, &reference); err != nil {
		return err
	}
	if reference == nil {
		return typedJSONNullError(path)
	}
	return rejectNullJSONFields(reference, path,
		"apiVersion",
		"kind",
		"namespace",
		"name",
		"uid",
		"revision",
		"digest",
	)
}

func rejectNullJSONFields(
	object map[string]json.RawMessage,
	path string,
	fields ...string,
) error {
	for _, field := range fields {
		raw, exists := object[field]
		if exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return typedJSONNullError(path + "." + field)
		}
	}
	return nil
}

func typedJSONNullError(path string) error {
	return boundedTypedJSONError(
		ErrTypedJSONType,
		"typed JSON type error: null is not allowed at path "+path,
	)
}
