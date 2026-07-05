package omnigent

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListItems(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok || d.Items == nil {
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	desc := c.Query("order") == "desc"
	page, err := d.Items.ListPage(c.Request.Context(), row.ID, limit, c.Query("after"), desc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list items failed"})
		return
	}
	data := make([]json.RawMessage, 0, len(page.Items))
	for _, it := range page.Items {
		data = append(data, it.Payload)
	}
	c.JSON(http.StatusOK, gin.H{
		"object":   "list",
		"data":     data,
		"has_more": page.HasMore,
	})
}
