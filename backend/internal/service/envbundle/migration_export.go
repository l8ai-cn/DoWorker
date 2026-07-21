package envbundle

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
)

func (s *Service) DecryptForMigration(kind string, data envbundle.BundleData) (map[string]string, error) {
	return s.decryptData(kind, data)
}
