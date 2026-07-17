package adminconnect

const ServiceName = "proto.admin.v1.AdminService"

const (
	GetDashboardStatsProcedure = "/" + ServiceName + "/GetDashboardStats"

	ListUsersProcedure         = "/" + ServiceName + "/ListUsers"
	GetUserProcedure           = "/" + ServiceName + "/GetUser"
	UpdateUserProcedure        = "/" + ServiceName + "/UpdateUser"
	DisableUserProcedure       = "/" + ServiceName + "/DisableUser"
	EnableUserProcedure        = "/" + ServiceName + "/EnableUser"
	GrantAdminProcedure        = "/" + ServiceName + "/GrantAdmin"
	RevokeAdminProcedure       = "/" + ServiceName + "/RevokeAdmin"
	VerifyUserEmailProcedure   = "/" + ServiceName + "/VerifyUserEmail"
	UnverifyUserEmailProcedure = "/" + ServiceName + "/UnverifyUserEmail"

	ListOrganizationsProcedure      = "/" + ServiceName + "/ListOrganizations"
	GetOrganizationProcedure        = "/" + ServiceName + "/GetOrganization"
	GetOrganizationMembersProcedure = "/" + ServiceName + "/GetOrganizationMembers"
	DeleteOrganizationProcedure     = "/" + ServiceName + "/DeleteOrganization"

	ListAuditLogsProcedure = "/" + ServiceName + "/ListAuditLogs"

	ListRunnersProcedure   = "/" + ServiceName + "/ListRunners"
	GetRunnerProcedure     = "/" + ServiceName + "/GetRunner"
	DisableRunnerProcedure = "/" + ServiceName + "/DisableRunner"
	EnableRunnerProcedure  = "/" + ServiceName + "/EnableRunner"
	DeleteRunnerProcedure  = "/" + ServiceName + "/DeleteRunner"

	ListRelaysProcedure           = "/" + ServiceName + "/ListRelays"
	GetRelayProcedure             = "/" + ServiceName + "/GetRelay"
	GetRelayStatsProcedure        = "/" + ServiceName + "/GetRelayStats"
	ForceUnregisterRelayProcedure = "/" + ServiceName + "/ForceUnregisterRelay"

	ListDeadLettersProcedure  = "/" + ServiceName + "/ListDeadLetters"
	ReplayDeadLetterProcedure = "/" + ServiceName + "/ReplayDeadLetter"

	ListExpertMarketReleasesProcedure   = "/" + ServiceName + "/ListExpertMarketReleases"
	GetExpertMarketReleaseProcedure     = "/" + ServiceName + "/GetExpertMarketRelease"
	ApproveExpertMarketReleaseProcedure = "/" + ServiceName + "/ApproveExpertMarketRelease"
	RejectExpertMarketReleaseProcedure  = "/" + ServiceName + "/RejectExpertMarketRelease"
)
