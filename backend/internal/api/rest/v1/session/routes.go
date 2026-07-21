package sessionapi

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, d Deps) {
	registerAuthRoutes(r, d)

	v1 := r.Group("/v1")
	v1.GET("/info", d.handleInfo)
	registerEmbedContextRoutes(v1, d)
	registerEmbedRoutes(v1, d)

	auth := v1.Group("")
	auth.Use(accessTokenMiddleware(d))
	auth.GET("/me", d.handleMe)

	orgScoped := auth.Group("")
	orgScoped.Use(headerTenant(d.Org))
	{
		orgScoped.GET("/org/usage/summary", d.handleOrgUsageSummary)
		orgScoped.GET("/agents", d.handleListAgents)
		orgScoped.GET("/harnesses", d.handleListHarnesses)
		orgScoped.GET("/runners", d.handleListRunners)
		orgScoped.GET("/policy-registry", d.handlePolicyRegistry)
		orgScoped.GET("/policies", d.handleListPolicies)
		orgScoped.POST("/policies", middleware.RequireAdmin(), d.handleCreatePolicy)
		orgScoped.PATCH("/policies/:id", middleware.RequireAdmin(), d.handlePatchPolicy)
		orgScoped.DELETE("/policies/:id", middleware.RequireAdmin(), d.handleDeletePolicy)
		orgScoped.GET("/sessions/projects", d.handleListProjects)
		orgScoped.GET("/sessions", d.handleListSessions)
		orgScoped.GET("/sessions/updates", d.handleSessionUpdates)
		orgScoped.GET("/sessions/by-pod/:pod_key", d.handleGetSessionByPodKey)
		orgScoped.POST("/sessions", d.handleCreateSession)
		orgScoped.POST("/sessions/:id/embed-context", d.handleCreateEmbedContext)
		orgScoped.POST("/sessions/import", d.handleImportSession)
		orgScoped.GET("/sessions/:id", d.handleGetSession)
		orgScoped.GET("/sessions/:id/relay-connection", d.handleGetSessionRelayConnection)
		orgScoped.PATCH("/sessions/:id", d.handlePatchSession)
		orgScoped.DELETE("/sessions/:id", d.handleDeleteSession)
		orgScoped.GET("/sessions/:id/read-state", d.handleGetReadState)
		orgScoped.PUT("/sessions/:id/read-state", d.handlePutReadState)
		orgScoped.GET("/sessions/:id/agent", d.handleGetSessionAgent)
		orgScoped.GET("/sessions/:id/owner", d.handleGetSessionOwner)
		orgScoped.GET("/sessions/:id/permissions", d.handleListPermissions)
		orgScoped.PUT("/sessions/:id/permissions", d.handlePutPermission)
		orgScoped.DELETE("/sessions/:id/permissions/:user_id", d.handleDeletePermission)
		orgScoped.GET("/sessions/:id/policies", d.handleListSessionPolicies)
		orgScoped.POST("/sessions/:id/policies", d.handleCreateSessionPolicy)
		orgScoped.DELETE("/sessions/:id/policies/:policy_id", d.handleDeleteSessionPolicy)
		orgScoped.POST("/sessions/:id/switch-agent", d.handleSwitchAgent)
		orgScoped.GET("/sessions/:id/codex_goal", d.handleGetCodexGoal)
		orgScoped.PUT("/sessions/:id/codex_goal", d.handlePutCodexGoal)
		orgScoped.PATCH("/sessions/:id/codex_goal/status", d.handlePatchCodexGoalStatus)
		orgScoped.DELETE("/sessions/:id/codex_goal", d.handleDeleteCodexGoal)
		orgScoped.POST("/sessions/:id/agent/mcp-servers", d.handleCreateMcpServer)
		orgScoped.PUT("/sessions/:id/agent/mcp-servers/:server_name", d.handleUpdateMcpServer)
		orgScoped.DELETE("/sessions/:id/agent/mcp-servers/:server_name", d.handleDeleteMcpServer)
		orgScoped.GET("/sessions/:id/child_sessions", d.handleListChildSessions)
		orgScoped.GET("/sessions/:id/comments", d.handleListComments)
		orgScoped.POST("/sessions/:id/comments", d.handleCreateComment)
		orgScoped.PATCH("/sessions/:id/comments/:comment_id", d.handlePatchComment)
		orgScoped.DELETE("/sessions/:id/comments/:comment_id", d.handleDeleteComment)
		orgScoped.POST("/sessions/:id/comments/send", d.handleSendComments)
		orgScoped.GET("/sessions/:id/items", d.handleListItems)
		orgScoped.GET("/sessions/:id/elicitations/:elicitation_id", d.handleGetElicitation)
		orgScoped.POST("/sessions/:id/elicitations/:elicitation_id/resolve", d.handleResolveElicitation)
		orgScoped.POST("/sessions/:id/events", d.handlePostEvent)
		orgScoped.GET("/sessions/:id/stream", d.handleSessionStream)
		orgScoped.POST("/sessions/:id/fork", d.handleForkSession)
		orgScoped.GET("/sessions/:id/resources/terminals", d.handleListTerminals)
		orgScoped.POST("/sessions/:id/resources/terminals", d.handleCreateTerminal)
		orgScoped.GET("/sessions/:id/resources/terminals/:terminal_id/attach", d.handleTerminalAttach)
		orgScoped.POST("/sessions/:id/resources/files", d.handleUploadSessionFile)
		orgScoped.GET("/sessions/:id/resources/files/:file_id/content", d.handleGetSessionFileContent)
		orgScoped.GET("/sessions/:id/artifacts/content", d.handleGetSessionArtifactRepresentation)
		orgScoped.GET("/sessions/:id/resources/environments/:env", d.handleSessionEnvironment)
		orgScoped.GET("/sessions/:id/resources/environments/:env/changes", d.handleSessionFilesystemChanges)
		orgScoped.GET("/sessions/:id/resources/environments/:env/diff/*filepath", d.handleSessionFilesystemDiff)
		orgScoped.GET("/sessions/:id/resources/environments/:env/search", d.handleSessionFilesystemSearch)
		orgScoped.GET("/sessions/:id/resources/environments/:env/filesystem", d.handleSessionFilesystemList)
		orgScoped.PUT("/sessions/:id/resources/environments/:env/filesystem/*filepath", d.handleSessionFilesystemWrite)
		orgScoped.GET("/sessions/:id/resources/environments/:env/filesystem/*filepath", d.handleSessionFilesystemList)
		orgScoped.GET("/model-resources", d.handleListModelResources)
		orgScoped.GET("/virtual-keys", d.handleListVirtualKeys)
		orgScoped.POST("/virtual-keys", d.handleCreateVirtualKey)
		orgScoped.DELETE("/virtual-keys/:id", d.handleRevokeVirtualKey)
		orgScoped.GET("/token-quotas", d.handleListTokenQuotas)
		orgScoped.PUT("/token-quotas", d.handleUpsertTokenQuota)
		orgScoped.DELETE("/token-quotas/:id", d.handleDeleteTokenQuota)
		orgScoped.GET("/usage/quota-report", d.handleQuotaReport)
		orgScoped.GET("/hosts", d.handleListHosts)
		orgScoped.GET("/hosts/:id/filesystem", d.handleHostFilesystem)
		orgScoped.GET("/hosts/:id/filesystem/*filepath", d.handleHostFilesystem)
		orgScoped.POST("/hosts/:id/directories", d.handleCreateHostDirectory)
		orgScoped.POST("/hosts/:id/runners", d.handleBindHostRunner)
	}
}

func registerEmbedContextRoutes(v1 *gin.RouterGroup, d Deps) {
	v1.POST("/embed-contexts/inspect", d.handleInspectEmbedContext)
	v1.POST("/embed-contexts/redeem", d.handleRedeemEmbedContext)
}

func registerEmbedRoutes(v1 *gin.RouterGroup, d Deps) {
	embedded := v1.Group("/embed")
	embedded.Use(d.embedSessionAuth())
	embedded.GET("/sessions/:id", d.handleGetSession)
	embedded.GET("/sessions/:id/items", d.handleListItems)
	embedded.GET("/sessions/:id/stream", d.handleSessionStream)
	embedded.GET(
		"/sessions/:id/resources/files/:file_id/content",
		d.handleGetSessionFileContent,
	)
	embedded.GET(
		"/sessions/:id/artifacts/content",
		d.handleGetSessionArtifactRepresentation,
	)
	embedded.GET(
		"/sessions/:id/resources/environments/:env/filesystem/*filepath",
		d.handleSessionFilesystemList,
	)
	embedded.GET(
		"/sessions/:id/resources/environments/:env/changes",
		d.handleSessionFilesystemChanges,
	)

	embedded.POST("/sessions/:id/events", d.handlePostEvent)
	registerEmbedAttachmentRoutes(embedded, d)

	approvals := embedded.Group("")
	approvals.Use(requireEmbedCapability("approve"))
	approvals.POST(
		"/sessions/:id/elicitations/:elicitation_id/resolve",
		d.handleResolveElicitation,
	)

	terminals := embedded.Group("")
	terminals.Use(requireEmbedCapability("terminal"))
	terminals.GET("/sessions/:id/resources/terminals", d.handleListTerminals)

	control := terminals.Group("")
	control.Use(requireEmbedCapability("control"))
	control.GET("/sessions/:id/relay-connection", d.handleGetSessionRelayConnection)

	acpControl := embedded.Group("")
	acpControl.Use(requireEmbedCapability("control"))
	acpControl.GET(
		"/sessions/:id/acp-relay-connection",
		d.handleGetSessionACPRelayConnection,
	)
}
