package envbundle

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
)

func (s *Service) DecryptForMigration(kind string, data envbundle.BundleData) (map[string]string, error) {
	return s.decryptData(kind, data)
}
