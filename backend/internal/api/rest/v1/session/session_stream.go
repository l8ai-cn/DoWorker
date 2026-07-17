package sessionapi

import (
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
	if err := prepareSessionStreamWriter(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session stream unavailable"})
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	status := mapSessionStatus(pod)
	ch := d.subscribeSessionStream(sessionID, status)
	defer d.Hub.Unsubscribe(sessionID, ch)
	if err := http.NewResponseController(c.Writer).Flush(); err != nil {
		return
	}

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
			if err := writeSessionStreamFrame(c.Writer, frame); err != nil {
				return
			}
		case <-ticker.C:
			if err := writeSessionStreamFrame(c.Writer, ": keepalive\n\n"); err != nil {
				return
			}
		}
	}
}

func (d *Deps) subscribeSessionStream(sessionID, status string) chan string {
	ch := d.Hub.Subscribe(sessionID)
	if d.Stream != nil {
		d.Stream.PublishSessionStatus(sessionID, status)
	}
	return ch
}
