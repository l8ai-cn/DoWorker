package executionclusterconnect

import (
	"net/http"

	"connectrpc.com/connect"
)

func Mount(mux *http.ServeMux, server *Server, options ...connect.HandlerOption) {
	mux.Handle(ListExecutionClustersProcedure, connect.NewUnaryHandler(
		ListExecutionClustersProcedure,
		server.ListExecutionClusters,
		options...,
	))
	mux.Handle(CreateRegistrationCommandProcedure, connect.NewUnaryHandler(
		CreateRegistrationCommandProcedure,
		server.CreateRegistrationCommand,
		options...,
	))
}
