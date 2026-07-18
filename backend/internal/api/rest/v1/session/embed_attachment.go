package sessionapi

import "github.com/gin-gonic/gin"

func (d *Deps) handleUploadEmbedAttachment(c *gin.Context) {
	d.handleUploadSessionFile(c)
}
