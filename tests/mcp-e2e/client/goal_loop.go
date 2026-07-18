package client

import (
	"context"
	"fmt"
	"strconv"
)

type CreateGoalLoopRequest struct {
	Name                string
	WorkerTemplateName  string
	Objective           string
	AcceptanceCriteria  []string
	VerificationCommand string
}

type GoalLoop struct {
	ID     int64
	Slug   string
	Name   string
	Status string
}

type GoalLoopPage struct {
	Items  []GoalLoop
	Total  int64
	Limit  int32
	Offset int32
}

func (r *REST) CreateGoalLoop(
	ctx context.Context,
	orgSlug string,
	req CreateGoalLoopRequest,
) (*GoalLoop, error) {
	manifest := map[string]any{
		"apiVersion": "agentsmesh.io/v1alpha1",
		"kind":       "GoalLoop",
		"metadata": map[string]any{
			"name": req.Name, "namespace": orgSlug,
			"displayName": req.Name,
		},
		"spec": map[string]any{
			"workerTemplateRef": map[string]any{
				"kind": "WorkerTemplate", "name": req.WorkerTemplateName,
			},
			"objective":           req.Objective,
			"acceptanceCriteria":  req.AcceptanceCriteria,
			"verificationCommand": req.VerificationCommand,
			"maxIterations":       3, "timeoutMinutes": 5,
			"noProgressLimit": 2, "sameErrorLimit": 2,
			"escalationPolicy": "fail",
		},
	}
	planID, err := r.validateAndPlanResource(ctx, orgSlug, manifest)
	if err != nil {
		return nil, err
	}
	wireReq := map[string]string{"orgSlug": orgSlug, "planId": planID}
	var wire struct {
		GoalLoopID string `json:"goalLoopId"`
	}
	if err := r.connectCall(
		ctx,
		"/proto.orchestration_resource.v1."+
			"OrchestrationResourceService/CreateGoalLoopFromPlan",
		wireReq,
		&wire,
	); err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(wire.GoalLoopID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"decode applied goal loop id %q: %w",
			wire.GoalLoopID,
			err,
		)
	}
	loop, err := r.GetGoalLoop(ctx, orgSlug, req.Name)
	if err != nil {
		return nil, fmt.Errorf("load applied goal loop %q: %w", req.Name, err)
	}
	if loop.ID != id {
		return nil, fmt.Errorf(
			"applied goal loop id = %d, persisted id = %d",
			id,
			loop.ID,
		)
	}
	return loop, nil
}

func (r *REST) GetGoalLoop(
	ctx context.Context,
	orgSlug, loopSlug string,
) (*GoalLoop, error) {
	wireReq := map[string]string{
		"orgSlug": orgSlug, "loopSlug": loopSlug,
	}
	var wire goalLoopWire
	if err := r.connectCall(
		ctx,
		"/proto.goalloop.v1.GoalLoopService/GetGoalLoop",
		wireReq,
		&wire,
	); err != nil {
		return nil, err
	}
	return decodeGoalLoop(wire)
}

func (r *REST) ListGoalLoops(
	ctx context.Context,
	orgSlug, query string,
	limit, offset int32,
) (*GoalLoopPage, error) {
	wireReq := map[string]any{
		"orgSlug": orgSlug,
		"query":   query,
		"limit":   limit,
		"offset":  offset,
	}
	var wire struct {
		Items  []goalLoopWire `json:"items"`
		Total  string         `json:"total"`
		Limit  int32          `json:"limit"`
		Offset int32          `json:"offset"`
	}
	if err := r.connectCall(ctx, "/proto.goalloop.v1.GoalLoopService/ListGoalLoops", wireReq, &wire); err != nil {
		return nil, err
	}
	var total int64
	if wire.Total != "" {
		var err error
		total, err = strconv.ParseInt(wire.Total, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"decode goal loop total %q: %w",
				wire.Total,
				err,
			)
		}
	}
	page := &GoalLoopPage{Total: total, Limit: wire.Limit, Offset: wire.Offset}
	for _, item := range wire.Items {
		loop, err := decodeGoalLoop(item)
		if err != nil {
			return nil, err
		}
		page.Items = append(page.Items, *loop)
	}
	return page, nil
}

func (r *REST) StartGoalLoop(ctx context.Context, orgSlug, loopSlug string) (*GoalLoop, error) {
	wireReq := map[string]string{"orgSlug": orgSlug, "loopSlug": loopSlug}
	var wire goalLoopWire
	if err := r.connectCall(ctx, "/proto.goalloop.v1.GoalLoopService/StartGoalLoop", wireReq, &wire); err != nil {
		return nil, err
	}
	return decodeGoalLoop(wire)
}

type goalLoopWire struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func decodeGoalLoop(wire goalLoopWire) (*GoalLoop, error) {
	id, err := strconv.ParseInt(wire.ID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("decode goal loop id %q: %w", wire.ID, err)
	}
	return &GoalLoop{ID: id, Slug: wire.Slug, Name: wire.Name, Status: wire.Status}, nil
}
