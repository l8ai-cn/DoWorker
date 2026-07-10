package audit

const (
	ActionProviderConnectionCreated            = "provider_connection.created"
	ActionProviderConnectionUpdated            = "provider_connection.updated"
	ActionProviderConnectionCredentialsRotated = "provider_connection.credentials_rotated"
	ActionProviderConnectionValidationStarted  = "provider_connection.validation_started"
	ActionProviderConnectionValidated          = "provider_connection.validated"
	ActionProviderConnectionEnabled            = "provider_connection.enabled"
	ActionProviderConnectionDisabled           = "provider_connection.disabled"
	ActionProviderConnectionDeleted            = "provider_connection.deleted"
	ActionModelResourceCreated                 = "model_resource.created"
	ActionModelResourceUpdated                 = "model_resource.updated"
	ActionModelResourceEnabled                 = "model_resource.enabled"
	ActionModelResourceDisabled                = "model_resource.disabled"
	ActionModelResourceDefaulted               = "model_resource.defaulted"
	ActionModelResourceDeleted                 = "model_resource.deleted"
)

const (
	ResourceProviderConnection = "provider_connection"
	ResourceModelResource      = "model_resource"
)
