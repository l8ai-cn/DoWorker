package main

import (
	"net/http"

	"connectrpc.com/connect"

	podconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/pod"
	runnerconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/runner"
	v1 "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	"github.com/anthropics/agentsmesh/backend/internal/config"
)

// Runner and Pod share runtime collaborators sourced from the same REST
// service container so both protocol surfaces observe identical instances.
func mountRunnerService(
	mux *http.ServeMux,
	svc *serviceContainer,
	rest *v1.Services,
	cfg *config.Config,
	opts []connect.HandlerOption,
) {
	serverOpts := []runnerconnect.Option{
		runnerconnect.WithBaseURL(cfg.BaseURL()),
	}
	if rest.VersionChecker != nil {
		serverOpts = append(serverOpts, runnerconnect.WithVersionChecker(rest.VersionChecker))
	}
	if rest.PodCoordinator != nil {
		serverOpts = append(serverOpts, runnerconnect.WithPodCoordinator(rest.PodCoordinator))
	}
	if rest.SandboxQueryService != nil {
		serverOpts = append(serverOpts, runnerconnect.WithSandboxQueryService(rest.SandboxQueryService))
	}
	if rest.UpgradeCommandSender != nil {
		serverOpts = append(serverOpts, runnerconnect.WithUpgradeCommandSender(rest.UpgradeCommandSender))
	}
	if rest.LogUploadSender != nil {
		serverOpts = append(serverOpts, runnerconnect.WithLogUploadSender(rest.LogUploadSender))
	}
	if rest.LogUploadService != nil {
		serverOpts = append(serverOpts, runnerconnect.WithLogUploadService(rest.LogUploadService))
	}
	if rest.GRPCRunnerHandler != nil && rest.GRPCRunnerHandler.PKIService() != nil {
		serverOpts = append(serverOpts, runnerconnect.WithPKIService(rest.GRPCRunnerHandler.PKIService()))
		serverOpts = append(serverOpts, runnerconnect.WithGRPCEndpoint(cfg.GRPC.Endpoint))
	}
	srv := runnerconnect.NewServer(svc.runner, svc.org, serverOpts...)
	runnerconnect.Mount(mux, srv, opts...)
	runnerconnect.MountPublic(mux, srv)
}

func mountPodService(
	mux *http.ServeMux,
	svc *serviceContainer,
	rest *v1.Services,
	cfg *config.Config,
	opts []connect.HandlerOption,
) {
	serverOpts := []podconnect.Option{podconnect.WithBaseURL(cfg.PublicWebBaseURL())}
	if rest.PodOrchestrator != nil {
		serverOpts = append(serverOpts, podconnect.WithOrchestrator(rest.PodOrchestrator))
	}
	if rest.PodCoordinator != nil {
		serverOpts = append(serverOpts, podconnect.WithPodCoordinator(rest.PodCoordinator))
		if sender := rest.PodCoordinator.GetCommandSender(); sender != nil {
			serverOpts = append(serverOpts, podconnect.WithCommandSender(sender))
		}
	}
	if rest.RelayManager != nil {
		serverOpts = append(serverOpts, podconnect.WithRelayManager(rest.RelayManager))
	}
	if rest.RelayTokenGenerator != nil {
		serverOpts = append(serverOpts, podconnect.WithTokenGenerator(rest.RelayTokenGenerator))
	}
	if rest.GeoResolver != nil {
		serverOpts = append(serverOpts, podconnect.WithGeoResolver(rest.GeoResolver))
	}
	if rest.Grant != nil {
		serverOpts = append(serverOpts, podconnect.WithGrantService(rest.Grant))
	}
	if rest.EventBus != nil {
		serverOpts = append(serverOpts, podconnect.WithEventBus(rest.EventBus))
	}
	if svc.workerCreation != nil {
		serverOpts = append(serverOpts, podconnect.WithWorkerCreation(svc.workerCreation))
	}
	if svc.workerDraftFiller != nil {
		serverOpts = append(serverOpts, podconnect.WithWorkerDraftFiller(svc.workerDraftFiller))
	}
	srv := podconnect.NewServer(svc.pod, svc.org, serverOpts...)
	podconnect.Mount(mux, srv, opts...)
}
