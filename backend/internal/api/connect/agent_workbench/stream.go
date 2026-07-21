package agentworkbenchconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	workbenchsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentworkbench"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

func (s *Server) StreamSessionDeltas(
	ctx context.Context,
	request *connect.Request[agentworkbenchv2.StreamSessionDeltasRequest],
	stream *connect.ServerStream[agentworkbenchv2.SessionDeltaBatch],
) error {
	cursor := request.Msg.GetCursor()
	if cursor == nil {
		return invalidArgument("cursor is required")
	}
	ctx, snapshot, err := s.loadAuthorizedSnapshot(
		ctx,
		request.Msg,
		cursor.GetSessionId(),
	)
	if err != nil {
		return err
	}
	position, err := validateCursor(cursor, snapshot)
	if err != nil {
		return err
	}
	position, err = s.replayTo(
		ctx,
		stream,
		position,
		snapshot,
		replayPageSize(request.Msg.GetReplayLimit()),
	)
	if err != nil {
		return err
	}
	if s.hub == nil {
		return unavailable("agent workbench delta stream is unavailable")
	}
	subscription := s.hub.Subscribe(cursor.GetSessionId())
	defer subscription.Close()

	bridge, err := s.loadSnapshot(ctx, cursor.GetSessionId())
	if err != nil {
		return err
	}
	position, err = s.replayTo(
		ctx,
		stream,
		position,
		bridge,
		replayPageSize(request.Msg.GetReplayLimit()),
	)
	if err != nil {
		return err
	}
	return streamLiveDeltas(ctx, stream, subscription, position)
}

func streamLiveDeltas(
	ctx context.Context,
	stream *connect.ServerStream[agentworkbenchv2.SessionDeltaBatch],
	subscription *workbenchsvc.DeltaSubscription,
	position streamPosition,
) error {
	deltas := subscription.Deltas
	failures := subscription.Errors
	for deltas != nil || failures != nil {
		select {
		case <-ctx.Done():
			return canceled(ctx.Err())
		case err, open := <-failures:
			if !open {
				failures = nil
				continue
			}
			if errors.Is(err, workbenchsvc.ErrSubscriberLagged) {
				return failedPrecondition("agent workbench stream lagged; resync required")
			}
			return internalError(err)
		case delta, open := <-deltas:
			if !open {
				deltas = nil
				continue
			}
			duplicate, next, err := validateLiveDelta(delta, position)
			if err != nil {
				return err
			}
			if duplicate {
				continue
			}
			decorated, err := authorizedDelta(ctx, delta)
			if err != nil {
				return internalError(err)
			}
			if err := stream.Send(decorated); err != nil {
				if ctx.Err() != nil {
					return canceled(ctx.Err())
				}
				return err
			}
			position = next
		}
	}
	return unavailable("agent workbench delta stream closed")
}
