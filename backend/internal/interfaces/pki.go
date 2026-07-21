package interfaces

import (
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/pki"
)

type PKICertificateIssuer interface {
	IssueRunnerCertificate(nodeID, orgSlug string) (*pki.CertificateInfo, error)

	CACertPEM() []byte
}
