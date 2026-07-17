package adminconnect

import (
	"net/http"

	"connectrpc.com/connect"
)

func MountExpertMarketplace(
	mux *http.ServeMux,
	srv *Server,
	opts ...connect.HandlerOption,
) {
	mux.Handle(ListExpertMarketReleasesProcedure, connect.NewUnaryHandler(
		ListExpertMarketReleasesProcedure,
		srv.ListExpertMarketReleases,
		opts...,
	))
	mux.Handle(GetExpertMarketReleaseProcedure, connect.NewUnaryHandler(
		GetExpertMarketReleaseProcedure,
		srv.GetExpertMarketRelease,
		opts...,
	))
	mux.Handle(ApproveExpertMarketReleaseProcedure, connect.NewUnaryHandler(
		ApproveExpertMarketReleaseProcedure,
		srv.ApproveExpertMarketRelease,
		opts...,
	))
	mux.Handle(RejectExpertMarketReleaseProcedure, connect.NewUnaryHandler(
		RejectExpertMarketReleaseProcedure,
		srv.RejectExpertMarketRelease,
		opts...,
	))
}
