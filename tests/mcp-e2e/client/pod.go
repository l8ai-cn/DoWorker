package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type CreatePodRequest struct {
	WorkerSpec WorkerSpecDraft
	Cols       int32
	Rows       int32
}

type Pod struct {
	ID       int64
	PodKey   string
	Status   string
	Agent    string
	OrgSlug  string
	RunnerID int64
}

func decodePodWire(raw json.RawMessage) (*Pod, error) {
	var wire struct {
		ID       string `json:"id"`
		PodKey   string `json:"podKey"`
		Status   string `json:"status"`
		Agent    string `json:"agentSlug"`
		OrgSlug  string `json:"organizationSlug,omitempty"`
		RunnerID string `json:"runnerId,omitempty"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, err
	}
	id, err := parsePositivePodWireID("id", wire.ID)
	if err != nil {
		return nil, err
	}
	runnerID, err := parsePositivePodWireID("runnerId", wire.RunnerID)
	if err != nil {
		return nil, err
	}
	return &Pod{
		ID: id, PodKey: wire.PodKey, Status: wire.Status,
		Agent: wire.Agent, OrgSlug: wire.OrgSlug, RunnerID: runnerID,
	}, nil
}

func parsePositivePodWireID(field, value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("decode pod %s %q: %w", field, value, err)
	}
	if id <= 0 {
		return 0, fmt.Errorf("decode pod %s %q: must be positive", field, value)
	}
	return id, nil
}

func (r *REST) CreatePod(
	ctx context.Context,
	orgSlug string,
	req CreatePodRequest,
) (*Pod, error) {
	wireReq := struct {
		OrgSlug    string         `json:"orgSlug"`
		Cols       int32          `json:"cols"`
		Rows       int32          `json:"rows"`
		WorkerSpec workerSpecWire `json:"workerSpec"`
	}{
		OrgSlug: orgSlug,
		Cols:    req.Cols, Rows: req.Rows,
		WorkerSpec: req.WorkerSpec.wire(),
	}
	var resp struct {
		Pod json.RawMessage `json:"pod"`
	}
	if err := r.connectCall(
		ctx,
		"/proto.pod.v1.PodService/CreatePod",
		wireReq,
		&resp,
	); err != nil {
		return nil, err
	}
	return decodePodWire(resp.Pod)
}

func (r *REST) TerminatePod(ctx context.Context, orgSlug, podKey string) error {
	req := map[string]string{"orgSlug": orgSlug, "podKey": podKey}
	return r.connectCall(
		ctx,
		"/proto.pod.v1.PodService/TerminatePod",
		req,
		nil,
	)
}

func (r *REST) GetPod(
	ctx context.Context,
	orgSlug, podKey string,
) (*Pod, error) {
	req := map[string]string{"orgSlug": orgSlug, "podKey": podKey}
	var podRaw json.RawMessage
	if err := r.connectCall(
		ctx,
		"/proto.pod.v1.PodService/GetPod",
		req,
		&podRaw,
	); err != nil {
		return nil, err
	}
	return decodePodWire(podRaw)
}
