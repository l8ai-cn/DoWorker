package orchestrationresource

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

var ErrCrossNamespaceReference = errors.New("cross-namespace reference")

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type Reference struct {
	APIVersion string       `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string       `json:"kind" yaml:"kind"`
	Namespace  slugkit.Slug `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name       slugkit.Slug `json:"name" yaml:"name"`
	UID        string       `json:"uid,omitempty" yaml:"uid,omitempty"`
	Revision   int64        `json:"revision,omitempty" yaml:"revision,omitempty"`
	Digest     string       `json:"digest,omitempty" yaml:"digest,omitempty"`
}

func (ref Reference) ValidateDraft(defaultNamespace string) error {
	if err := slugkit.Validate(defaultNamespace); err != nil {
		return fmt.Errorf("defaultNamespace: %w", err)
	}

	apiVersion := ref.APIVersion
	if apiVersion == "" {
		apiVersion = APIVersionV1Alpha1
	}
	if err := (TypeMeta{APIVersion: apiVersion, Kind: ref.Kind}).Validate(); err != nil {
		return err
	}

	namespace := ref.Namespace.String()
	if namespace == "" {
		namespace = defaultNamespace
	} else {
		if err := slugkit.Validate(namespace); err != nil {
			return fmt.Errorf("reference.namespace: %w", err)
		}
		if namespace != defaultNamespace {
			return fmt.Errorf("reference.namespace: %w", ErrCrossNamespaceReference)
		}
	}

	if err := slugkit.Validate(ref.Name.String()); err != nil {
		return fmt.Errorf("reference.name: %w", err)
	}
	if ref.UID != "" {
		return fmt.Errorf("reference.uid must be empty")
	}
	if ref.Digest != "" {
		return fmt.Errorf("reference.digest must be empty")
	}
	if ref.Revision < 0 {
		return fmt.Errorf("reference.revision must not be negative")
	}
	return nil
}

func (ref Reference) ValidateResolved(defaultNamespace string) error {
	if err := slugkit.Validate(defaultNamespace); err != nil {
		return fmt.Errorf("defaultNamespace: %w", err)
	}

	if err := (TypeMeta{APIVersion: ref.APIVersion, Kind: ref.Kind}).Validate(); err != nil {
		return err
	}

	namespace := ref.Namespace.String()
	if namespace == "" {
		return fmt.Errorf("reference.namespace: is required")
	}
	if err := slugkit.Validate(namespace); err != nil {
		return fmt.Errorf("reference.namespace: %w", err)
	}
	if namespace != defaultNamespace {
		return fmt.Errorf("reference.namespace: %w", ErrCrossNamespaceReference)
	}

	if err := slugkit.Validate(ref.Name.String()); err != nil {
		return fmt.Errorf("reference.name: %w", err)
	}
	if ref.UID == "" {
		return fmt.Errorf("reference.uid: is required")
	}
	if err := validateMetadataText("reference.uid", ref.UID, maxServerFieldRunes); err != nil {
		return err
	}
	if ref.Revision <= 0 {
		return fmt.Errorf("reference.revision must be greater than 0")
	}
	if !digestPattern.MatchString(ref.Digest) {
		return fmt.Errorf("reference.digest must match %s", digestPattern.String())
	}
	return nil
}
