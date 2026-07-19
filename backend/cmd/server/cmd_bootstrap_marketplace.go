package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/domain/organization"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/infra/database"
	"github.com/anthropics/agentsmesh/backend/internal/service/operatorcatalog"
	skillsvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"gorm.io/gorm"
)

type marketplaceBootstrapOptions struct {
	organizationSlug string
	publisherEmail   string
	reviewerEmail    string
	modelResourceID  int64
	runtimeImageID   int64
}

func runBootstrapMarketplace(arguments []string) error {
	options, err := parseMarketplaceBootstrapOptions(arguments)
	if err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := database.New(cfg.Database)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	services, err := initializeServices(cfg, db, nil)
	if err != nil {
		return err
	}
	defer services.Close()
	expertService, _ := newExpertAndSkillServices(
		cfg,
		db,
		services,
		nil,
		slog.Default(),
	)
	var packager skillsvc.SkillPackagerBridge
	if services.extension != nil {
		packager = services.extension.SkillPackager()
	}
	platformSkills := skillsvc.NewPlatformCatalogService(
		infra.NewSkillCatalogRepository(db),
		packager,
	)
	identity, err := resolveMarketplaceBootstrapIdentity(
		context.Background(),
		db,
		options,
	)
	if err != nil {
		return err
	}
	organizationSlug, err := slugkit.NewFromTrusted(options.organizationSlug)
	if err != nil {
		return err
	}
	result, err := operatorcatalog.NewBootstrapper(
		platformSkills,
		expertService,
		services.workerCreation,
		services.workerSpecs,
		infra.NewWorkerSpecDependencyArtifactRepository(db),
	).Run(context.Background(), operatorcatalog.BootstrapRequest{
		OrganizationID:   identity.organizationID,
		OrganizationSlug: organizationSlug,
		PublisherUserID:  identity.publisherID,
		ReviewerUserID:   identity.reviewerID,
		ModelResourceID:  options.modelResourceID,
		RuntimeImageID:   options.runtimeImageID,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}

func parseMarketplaceBootstrapOptions(
	arguments []string,
) (marketplaceBootstrapOptions, error) {
	var options marketplaceBootstrapOptions
	flags := flag.NewFlagSet("bootstrap-marketplace", flag.ContinueOnError)
	flags.StringVar(&options.organizationSlug, "organization", "", "publisher organization slug")
	flags.StringVar(&options.publisherEmail, "publisher", "", "publisher user email")
	flags.StringVar(&options.reviewerEmail, "reviewer", "", "system administrator email")
	flags.Int64Var(&options.modelResourceID, "model-resource-id", 0, "publisher model resource ID")
	flags.Int64Var(&options.runtimeImageID, "runtime-image-id", 4, "video-studio runtime image ID")
	if err := flags.Parse(arguments); err != nil {
		return options, err
	}
	if options.organizationSlug == "" || options.publisherEmail == "" ||
		options.reviewerEmail == "" || options.modelResourceID <= 0 ||
		options.runtimeImageID <= 0 {
		return options, errors.New(
			"organization, publisher, reviewer, model-resource-id and runtime-image-id are required",
		)
	}
	return options, nil
}

type marketplaceBootstrapIdentity struct {
	organizationID int64
	publisherID    int64
	reviewerID     int64
}

func resolveMarketplaceBootstrapIdentity(
	ctx context.Context,
	db *gorm.DB,
	options marketplaceBootstrapOptions,
) (marketplaceBootstrapIdentity, error) {
	var organizationRow organization.Organization
	if err := db.WithContext(ctx).
		Where("slug = ?", options.organizationSlug).
		First(&organizationRow).Error; err != nil {
		return marketplaceBootstrapIdentity{}, fmt.Errorf("publisher organization: %w", err)
	}
	publisher, err := bootstrapUserByEmail(ctx, db, options.publisherEmail)
	if err != nil {
		return marketplaceBootstrapIdentity{}, fmt.Errorf("publisher: %w", err)
	}
	var membershipCount int64
	if err := db.WithContext(ctx).
		Table("organization_members").
		Where(
			"organization_id = ? AND user_id = ?",
			organizationRow.ID,
			publisher.ID,
		).
		Count(&membershipCount).Error; err != nil {
		return marketplaceBootstrapIdentity{}, err
	}
	if membershipCount != 1 {
		return marketplaceBootstrapIdentity{}, errors.New(
			"publisher is not a member of the publisher organization",
		)
	}
	reviewer, err := bootstrapUserByEmail(ctx, db, options.reviewerEmail)
	if err != nil {
		return marketplaceBootstrapIdentity{}, fmt.Errorf("reviewer: %w", err)
	}
	if !reviewer.IsActive || !reviewer.IsSystemAdmin {
		return marketplaceBootstrapIdentity{}, errors.New(
			"reviewer must be an active system administrator",
		)
	}
	return marketplaceBootstrapIdentity{
		organizationID: organizationRow.ID,
		publisherID:    publisher.ID,
		reviewerID:     reviewer.ID,
	}, nil
}

func bootstrapUserByEmail(
	ctx context.Context,
	db *gorm.DB,
	email string,
) (*user.User, error) {
	var row user.User
	if err := db.WithContext(ctx).
		Where("email = ?", email).
		First(&row).Error; err != nil {
		return nil, err
	}
	if !row.IsActive {
		return nil, errors.New("user is inactive")
	}
	return &row, nil
}
