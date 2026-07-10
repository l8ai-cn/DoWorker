package airesourceconnect

import (
	"net/http"

	"connectrpc.com/connect"
)

const serviceName = "proto.ai_resource.v1.AIResourceService"

const (
	GetCatalogProcedure                         = "/" + serviceName + "/GetCatalog"
	ListPersonalConnectionsProcedure            = "/" + serviceName + "/ListPersonalConnections"
	ListOrganizationConnectionsProcedure        = "/" + serviceName + "/ListOrganizationConnections"
	ListPersonalEffectiveResourcesProcedure     = "/" + serviceName + "/ListPersonalEffectiveResources"
	ListOrganizationEffectiveResourcesProcedure = "/" + serviceName + "/ListOrganizationEffectiveResources"
	CreatePersonalConnectionProcedure           = "/" + serviceName + "/CreatePersonalConnection"
	CreateOrganizationConnectionProcedure       = "/" + serviceName + "/CreateOrganizationConnection"
	UpdateConnectionProcedure                   = "/" + serviceName + "/UpdateConnection"
	RotateConnectionCredentialsProcedure        = "/" + serviceName + "/RotateConnectionCredentials"
	SetConnectionEnabledProcedure               = "/" + serviceName + "/SetConnectionEnabled"
	ValidateConnectionProcedure                 = "/" + serviceName + "/ValidateConnection"
	DeleteConnectionProcedure                   = "/" + serviceName + "/DeleteConnection"
	CreateResourceProcedure                     = "/" + serviceName + "/CreateResource"
	UpdateResourceProcedure                     = "/" + serviceName + "/UpdateResource"
	SetResourceEnabledProcedure                 = "/" + serviceName + "/SetResourceEnabled"
	DeleteResourceProcedure                     = "/" + serviceName + "/DeleteResource"
	SetDefaultProcedure                         = "/" + serviceName + "/SetDefault"
)

func Mount(mux *http.ServeMux, server *Server, options ...connect.HandlerOption) {
	mux.Handle(GetCatalogProcedure, connect.NewUnaryHandler(GetCatalogProcedure, server.GetCatalog, options...))
	mux.Handle(ListPersonalConnectionsProcedure, connect.NewUnaryHandler(ListPersonalConnectionsProcedure, server.ListPersonalConnections, options...))
	mux.Handle(ListOrganizationConnectionsProcedure, connect.NewUnaryHandler(ListOrganizationConnectionsProcedure, server.ListOrganizationConnections, options...))
	mux.Handle(ListPersonalEffectiveResourcesProcedure, connect.NewUnaryHandler(ListPersonalEffectiveResourcesProcedure, server.ListPersonalEffectiveResources, options...))
	mux.Handle(ListOrganizationEffectiveResourcesProcedure, connect.NewUnaryHandler(ListOrganizationEffectiveResourcesProcedure, server.ListOrganizationEffectiveResources, options...))
	mux.Handle(CreatePersonalConnectionProcedure, connect.NewUnaryHandler(CreatePersonalConnectionProcedure, server.CreatePersonalConnection, options...))
	mux.Handle(CreateOrganizationConnectionProcedure, connect.NewUnaryHandler(CreateOrganizationConnectionProcedure, server.CreateOrganizationConnection, options...))
	mux.Handle(UpdateConnectionProcedure, connect.NewUnaryHandler(UpdateConnectionProcedure, server.UpdateConnection, options...))
	mux.Handle(RotateConnectionCredentialsProcedure, connect.NewUnaryHandler(RotateConnectionCredentialsProcedure, server.RotateConnectionCredentials, options...))
	mux.Handle(SetConnectionEnabledProcedure, connect.NewUnaryHandler(SetConnectionEnabledProcedure, server.SetConnectionEnabled, options...))
	mux.Handle(ValidateConnectionProcedure, connect.NewUnaryHandler(ValidateConnectionProcedure, server.ValidateConnection, options...))
	mux.Handle(DeleteConnectionProcedure, connect.NewUnaryHandler(DeleteConnectionProcedure, server.DeleteConnection, options...))
	mux.Handle(CreateResourceProcedure, connect.NewUnaryHandler(CreateResourceProcedure, server.CreateResource, options...))
	mux.Handle(UpdateResourceProcedure, connect.NewUnaryHandler(UpdateResourceProcedure, server.UpdateResource, options...))
	mux.Handle(SetResourceEnabledProcedure, connect.NewUnaryHandler(SetResourceEnabledProcedure, server.SetResourceEnabled, options...))
	mux.Handle(DeleteResourceProcedure, connect.NewUnaryHandler(DeleteResourceProcedure, server.DeleteResource, options...))
	mux.Handle(SetDefaultProcedure, connect.NewUnaryHandler(SetDefaultProcedure, server.SetDefault, options...))
}
