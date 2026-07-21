package airesource

import (
	"errors"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
)

var (
	ErrNotFound                    = errors.New("AI resource not found")
	ErrForbidden                   = errors.New("AI resource access forbidden")
	ErrInvalidOwner                = errors.New("invalid AI resource owner")
	ErrInvalidProvider             = errors.New("invalid AI resource provider")
	ErrInvalidCredentials          = errors.New("invalid AI resource credentials")
	ErrInvalidEndpoint             = errors.New("invalid AI resource endpoint")
	ErrDisabled                    = errors.New("AI resource disabled")
	ErrUnhealthy                   = errors.New("AI resource unhealthy")
	ErrUnchecked                   = errors.New("AI resource unchecked")
	ErrIncompatibleModality        = errors.New("incompatible AI resource modality")
	ErrIncompatibleCapability      = errors.New("incompatible AI resource capability")
	ErrIncompatibleProtocolAdapter = errors.New("incompatible AI resource protocol adapter")
	ErrInvalidRequirements         = errors.New("invalid AI resource resolution requirements")
	ErrEncrypt                     = errors.New("AI resource credential encryption failed")
	ErrDecrypt                     = errors.New("AI resource credential decryption failed")
	ErrValidation                  = errors.New("AI resource connection validation failed")
	ErrProviderEndpointUnavailable = errors.New("AI resource provider endpoint unavailable")
	ErrProbeUnsupported            = errors.New("AI resource provider validation unsupported")
	ErrAudit                       = errors.New("AI resource audit failed")
	ErrConflict                    = domain.ErrConflict
)
