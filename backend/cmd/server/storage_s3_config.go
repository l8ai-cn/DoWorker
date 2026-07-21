package main

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/storage"
)

func storageS3Config(cfg *config.Config) storage.S3Config {
	return storage.S3Config{
		Endpoint:       cfg.Storage.Endpoint,
		PublicEndpoint: cfg.Storage.PublicEndpoint,
		RunnerEndpoint: cfg.Storage.RunnerEndpoint,
		Region:         cfg.Storage.Region,
		Bucket:         cfg.Storage.Bucket,
		AccessKey:      cfg.Storage.AccessKey,
		SecretKey:      cfg.Storage.SecretKey,
		UseSSL:         cfg.Storage.UseSSL,
		PublicUseSSL:   cfg.Storage.PublicUseSSL,
		UsePathStyle:   cfg.Storage.UsePathStyle,
	}
}
