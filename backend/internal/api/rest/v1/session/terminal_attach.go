package sessionapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var terminalUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (d *Deps) handleTerminalAttach(c *gin.Context) {
	_, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if c.Param("terminal_id") != terminalMainID {
		c.JSON(http.StatusNotFound, gin.H{"error": "terminal not found", "code": "not_found"})
		return
	}
	if pod == nil || !pod.IsActive() || pod.RunnerID == 0 {
		closeTerminalWS(c, 4404, "session not found")
		return
	}
	connection, status, message := d.sessionRelayConnection(c, pod)
	if status != 0 {
		c.JSON(status, gin.H{"error": message})
		return
	}
	clientWS, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	upstream, _, err := dialRelayWS(connection.RelayURL, connection.Token)
	if err != nil {
		_ = clientWS.Close()
		return
	}
	_ = upstream.WriteMessage(websocket.BinaryMessage, []byte{relayMsgResync})
	readOnly := strings.EqualFold(c.Query("read_only"), "true")
	go func() {
		defer clientWS.Close()
		defer upstream.Close()
		bridgeTerminalWS(clientWS, upstream, readOnly)
	}()
}

func closeTerminalWS(c *gin.Context, code int, reason string) {
	ws, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": reason})
		return
	}
	_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason))
	_ = ws.Close()
}
