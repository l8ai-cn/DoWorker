package sessionapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

var errSessionCompensationFailed = errors.New("session compensation failed")

func (d *Deps) terminateCreatedSessionPod(ctx context.Context, podKey string) error {
	cleanupCtx, cancel := sessionCompensationContext(ctx)
	defer cancel()
	return d.terminateSessionPod(cleanupCtx, podKey)
}

func (d *Deps) terminateSessionPod(ctx context.Context, podKey string) error {
	err := d.PodCoordinator.TerminatePod(ctx, podKey)
	if errors.Is(err, runnerservice.ErrPodAlreadyTerminated) {
		return nil
	}
	return err
}

func sessionCompensationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
}

func joinSessionCompensationFailure(primary, cleanup error) error {
	if cleanup == nil {
		return primary
	}
	return errors.Join(primary, errSessionCompensationFailed, cleanup)
}

func writeSessionCreationFailure(
	c *gin.Context,
	message string,
	cleanupErr error,
) {
	if cleanupErr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
		return
	}
	slog.ErrorContext(
		c.Request.Context(),
		"session creation compensation failed",
		"error",
		cleanupErr,
	)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "session creation cleanup failed",
		"code":  "session_compensation_failed",
	})
}

func writeSessionCreationCommitFailure(
	c *gin.Context,
	message string,
	cause error,
	cleanupErr error,
) {
	if cleanupErr != nil {
		writeSessionCreationFailure(c, message, cleanupErr)
		return
	}
	switch {
	case errors.Is(cause, podDomain.ErrQueueFull):
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "runner queue full",
			"code":  "runner_queue_full",
		})
	case errors.Is(cause, runnerservice.ErrRunnerNotConnected),
		errors.Is(cause, runnerservice.ErrRunnerOffline):
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "runner unavailable",
			"code":  "runner_unavailable",
		})
	default:
		writeSessionCreationFailure(c, message, nil)
	}
}
