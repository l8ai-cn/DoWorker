package client

import (
	"context"
	"fmt"
	"strconv"
)

type Runner struct {
	ID                int64  `json:"-"`
	NodeID            string `json:"nodeId"`
	Status            string `json:"status"`
	IsEnabled         bool   `json:"isEnabled"`
	MaxConcurrentPods int32  `json:"maxConcurrentPods"`
}

func (r *REST) ListRunners(ctx context.Context, orgSlug string) ([]Runner, error) {
	request := map[string]string{"orgSlug": orgSlug}
	var response struct {
		Items []struct {
			ID                string `json:"id"`
			NodeID            string `json:"nodeId"`
			Status            string `json:"status"`
			IsEnabled         bool   `json:"isEnabled"`
			MaxConcurrentPods int32  `json:"maxConcurrentPods"`
		} `json:"items"`
	}
	if err := r.connectCall(
		ctx,
		"/proto.runner_api.v1.RunnerService/ListRunners",
		request,
		&response,
	); err != nil {
		return nil, err
	}
	runners := make([]Runner, 0, len(response.Items))
	for _, item := range response.Items {
		id, err := strconv.ParseInt(item.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("decode runner id %q: %w", item.ID, err)
		}
		runners = append(runners, Runner{
			ID:                id,
			NodeID:            item.NodeID,
			Status:            item.Status,
			IsEnabled:         item.IsEnabled,
			MaxConcurrentPods: item.MaxConcurrentPods,
		})
	}
	return runners, nil
}

func (r *REST) UpdateRunnerScheduling(
	ctx context.Context,
	orgSlug string,
	runnerID int64,
	enabled bool,
	maxConcurrentPods int32,
) error {
	request := map[string]any{
		"orgSlug":           orgSlug,
		"id":                strconv.FormatInt(runnerID, 10),
		"isEnabled":         enabled,
		"maxConcurrentPods": maxConcurrentPods,
	}
	return r.connectCall(
		ctx,
		"/proto.runner_api.v1.RunnerService/UpdateRunner",
		request,
		nil,
	)
}
