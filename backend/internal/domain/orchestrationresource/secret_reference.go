package orchestrationresource

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type SecretReference struct {
	Name     slugkit.Slug `json:"name" yaml:"name"`
	Key      slugkit.Slug `json:"key" yaml:"key"`
	Revision int64        `json:"revision,omitempty" yaml:"revision,omitempty"`
}

func (ref SecretReference) Validate() error {
	if err := slugkit.Validate(ref.Name.String()); err != nil {
		return fmt.Errorf("secretReference.name: %w", err)
	}
	if err := slugkit.Validate(ref.Key.String()); err != nil {
		return fmt.Errorf("secretReference.key: %w", err)
	}
	if ref.Revision < 0 {
		return fmt.Errorf("secretReference.revision must not be negative")
	}
	return nil
}
