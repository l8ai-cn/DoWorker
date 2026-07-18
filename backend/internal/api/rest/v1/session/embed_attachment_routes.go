package sessionapi

import "github.com/gin-gonic/gin"

func registerEmbedAttachmentRoutes(embedded *gin.RouterGroup, d Deps) {
	write := embedded.Group("")
	write.Use(requireEmbedCapability("write"))
	write.POST(
		"/sessions/:id/resources/files",
		d.handleUploadEmbedAttachment,
	)
}
