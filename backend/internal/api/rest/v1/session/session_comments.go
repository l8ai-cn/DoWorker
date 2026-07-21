package sessionapi

import (
	"net/http"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/sessioncomment"
	commentsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessioncomment"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListComments(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if d.SessionComments == nil {
		c.JSON(http.StatusOK, []any{})
		return
	}
	rows, err := d.SessionComments.List(c.Request.Context(), row.ID, c.Query("path"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, commentWire(&rows[i], row.ID))
	}
	c.JSON(http.StatusOK, out)
}

func (d *Deps) handleCreateComment(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	if d.SessionComments == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	var body struct {
		Path          string  `json:"path"`
		StartIndex    int     `json:"start_index"`
		EndIndex      int     `json:"end_index"`
		Body          string  `json:"body"`
		AnchorContent *string `json:"anchor_content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Path == "" || body.EndIndex <= body.StartIndex {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	author := d.viewerEmail(c)
	var createdBy *string
	if author != "" {
		createdBy = &author
	}
	rowComment, err := d.SessionComments.Create(c.Request.Context(), commentsvc.CreateInput{
		SessionID: row.ID, Path: body.Path, Body: body.Body,
		StartIndex: body.StartIndex, EndIndex: body.EndIndex,
		AnchorContent: body.AnchorContent, CreatedBy: createdBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusOK, commentWire(rowComment, row.ID))
}

func (d *Deps) handlePatchComment(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	if d.SessionComments == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	var body struct {
		Status *string `json:"status"`
		Body   *string `json:"body"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	rowComment, err := d.SessionComments.Update(c.Request.Context(), row.ID, c.Param("comment_id"), body.Status, body.Body)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, commentWire(rowComment, row.ID))
}

func (d *Deps) handleDeleteComment(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	if d.SessionComments == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	if err := d.SessionComments.Delete(c.Request.Context(), row.ID, c.Param("comment_id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleSendComments(c *gin.Context) {
	row, _, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	if !d.requireSessionLevel(c, row, levelEdit) {
		return
	}
	if d.SessionComments == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	var body struct {
		CommentIDs  []string `json:"comment_ids"`
		Instruction string   `json:"instruction"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.CommentIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	comments := make([]domain.Comment, 0, len(body.CommentIDs))
	for _, id := range body.CommentIDs {
		cmt, err := d.SessionComments.Get(c.Request.Context(), row.ID, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
		comments = append(comments, *cmt)
	}
	msg := commentsvc.FormatSendMessage(comments, body.Instruction)
	_ = d.SessionComments.MarkAddressed(c.Request.Context(), row.ID, body.CommentIDs)
	c.JSON(http.StatusOK, gin.H{
		"formatted_message": msg, "sent_comment_ids": body.CommentIDs,
	})
}

func commentWire(row *domain.Comment, sessionID string) map[string]any {
	wire := map[string]any{
		"id": row.ID, "conversation_id": sessionID, "path": row.Path,
		"start_index": row.StartIndex, "end_index": row.EndIndex,
		"body": row.Body, "status": row.Status,
		"created_at": row.CreatedAt.Unix(), "updated_at": row.UpdatedAt.UnixMicro(),
		"anchor_content": row.AnchorContent, "created_by": row.CreatedBy,
	}
	if row.AnchorContent == nil {
		wire["anchor_content"] = nil
	}
	return wire
}
