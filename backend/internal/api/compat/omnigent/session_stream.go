package omnigent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (d *Deps) handleSessionStream(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Hub == nil {
		return
	}
	sessionID := c.Param("id")
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)
	flusher, canFlush := c.Writer.(http.Flusher)

	status := mapSessionStatus(pod)
	_, _ = fmt.Fprint(c.Writer, formatSSE("session.status", map[string]any{
		"conversation_id": sessionID, "status": status,
	}))
	if canFlush {
		flusher.Flush()
	}

	ch := d.Hub.Subscribe(sessionID)
	defer d.Hub.Unsubscribe(sessionID, ch)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case frame, open := <-ch:
			if !open {
				return
			}
			_, _ = fmt.Fprint(c.Writer, frame)
			if canFlush {
				flusher.Flush()
			}
		case <-ticker.C:
			_, _ = fmt.Fprint(c.Writer, ": keepalive\n\n")
			if canFlush {
				flusher.Flush()
			}
		}
	}
}
