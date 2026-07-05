// Package knowledgebaseconnect hosts Connect-RPC handlers for the org-scoped
// KnowledgeBaseService — llm-wiki knowledge bases backed by internal Gitea.
//
// Split (200-line rule):
//   - server.go  — scaffolding + Mount
//   - crud.go    — List/Get/Create/Update/Delete
//   - mounts.go  — agent mount RPCs
//   - files.go   — repo content browsing (file/dir)
//   - convert.go — domain → proto translation + error mapping
package knowledgebaseconnect

import (
	"net/http"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
)

const ServiceName = "proto.knowledgebase.v1.KnowledgeBaseService"

const (
	ListKnowledgeBasesProcedure   = "/" + ServiceName + "/ListKnowledgeBases"
	GetKnowledgeBaseProcedure     = "/" + ServiceName + "/GetKnowledgeBase"
	CreateKnowledgeBaseProcedure  = "/" + ServiceName + "/CreateKnowledgeBase"
	UpdateKnowledgeBaseProcedure  = "/" + ServiceName + "/UpdateKnowledgeBase"
	DeleteKnowledgeBaseProcedure  = "/" + ServiceName + "/DeleteKnowledgeBase"
	SetAgentMountsProcedure       = "/" + ServiceName + "/SetAgentMounts"
	ListAgentMountsProcedure      = "/" + ServiceName + "/ListAgentMounts"
	GetKnowledgeBaseFileProcedure = "/" + ServiceName + "/GetKnowledgeBaseFile"
	ListKnowledgeBaseDirProcedure = "/" + ServiceName + "/ListKnowledgeBaseDir"
)

type Server struct {
	svc    *kbservice.Service
	orgSvc middleware.OrganizationService
}

func NewServer(svc *kbservice.Service, orgSvc middleware.OrganizationService) *Server {
	return &Server{svc: svc, orgSvc: orgSvc}
}

func Mount(mux *http.ServeMux, srv *Server, opts ...connect.HandlerOption) {
	mux.Handle(ListKnowledgeBasesProcedure, connect.NewUnaryHandler(
		ListKnowledgeBasesProcedure, srv.ListKnowledgeBases, opts...,
	))
	mux.Handle(GetKnowledgeBaseProcedure, connect.NewUnaryHandler(
		GetKnowledgeBaseProcedure, srv.GetKnowledgeBase, opts...,
	))
	mux.Handle(CreateKnowledgeBaseProcedure, connect.NewUnaryHandler(
		CreateKnowledgeBaseProcedure, srv.CreateKnowledgeBase, opts...,
	))
	mux.Handle(UpdateKnowledgeBaseProcedure, connect.NewUnaryHandler(
		UpdateKnowledgeBaseProcedure, srv.UpdateKnowledgeBase, opts...,
	))
	mux.Handle(DeleteKnowledgeBaseProcedure, connect.NewUnaryHandler(
		DeleteKnowledgeBaseProcedure, srv.DeleteKnowledgeBase, opts...,
	))
	mux.Handle(SetAgentMountsProcedure, connect.NewUnaryHandler(
		SetAgentMountsProcedure, srv.SetAgentMounts, opts...,
	))
	mux.Handle(ListAgentMountsProcedure, connect.NewUnaryHandler(
		ListAgentMountsProcedure, srv.ListAgentMounts, opts...,
	))
	mux.Handle(GetKnowledgeBaseFileProcedure, connect.NewUnaryHandler(
		GetKnowledgeBaseFileProcedure, srv.GetKnowledgeBaseFile, opts...,
	))
	mux.Handle(ListKnowledgeBaseDirProcedure, connect.NewUnaryHandler(
		ListKnowledgeBaseDirProcedure, srv.ListKnowledgeBaseDir, opts...,
	))
}
