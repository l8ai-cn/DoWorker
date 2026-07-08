package v1

import "github.com/gin-gonic/gin"

func registerIMBridgeRoutes(rg *gin.RouterGroup, svc *Services) {
	if svc.IMBridge == nil {
		return
	}
	h := NewIMBridgeHandler(svc.IMBridge)
	group := rg.Group("/im-channels")
	{
		group.GET("/providers", h.ListProviders)
		group.POST("/weixin/qr/start", h.StartWeixinQRLogin)
		group.GET("/weixin/qr/:sessionId/status", h.GetWeixinQRLoginStatus)
		group.GET("/weixin/qr/:sessionId/image", h.GetWeixinQRImage)
		group.GET("", h.ListConnections)
		group.POST("", h.CreateConnection)
		group.GET("/:connectionId", h.GetConnection)
		group.PATCH("/:connectionId", h.UpdateConnection)
		group.DELETE("/:connectionId", h.DeleteConnection)
	}
}
