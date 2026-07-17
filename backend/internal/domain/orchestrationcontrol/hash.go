package orchestrationcontrol

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	planHashPayloadVersion  = 2
	maxOptionsRevisionRunes = 128
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type PlanHashInput struct {
	Operation           PlanOperation
	Scope               Scope
	Target              ResourceTarget
	BaseUID             string
	BaseResourceVersion int64
	DraftHash           string
	ResolvedReferences  []ResolvedReference
	ArtifactDigest      string
	OptionsRevision     string
}

type planHashPayload struct {
	Version             int                 `json:"version"`
	Operation           PlanOperation       `json:"operation"`
	Scope               Scope               `json:"scope"`
	Target              ResourceTarget      `json:"target"`
	BaseUID             string              `json:"baseUid"`
	BaseResourceVersion int64               `json:"baseResourceVersion"`
	DraftHash           string              `json:"draftHash"`
	ResolvedReferences  []ResolvedReference `json:"resolvedReferences"`
	ArtifactDigest      string              `json:"artifactDigest"`
	OptionsRevision     string              `json:"optionsRevision"`
}

func ComputePlanHash(input PlanHashInput) (string, error) {
	if err := input.validate(); err != nil {
		return "", err
	}
	references, err := sortedResolvedReferences(input.Scope, input.ResolvedReferences)
	if err != nil {
		return "", err
	}
	return DigestCanonicalJSON(planHashPayload{
		Version:             planHashPayloadVersion,
		Operation:           input.Operation,
		Scope:               input.Scope,
		Target:              input.Target,
		BaseUID:             input.BaseUID,
		BaseResourceVersion: input.BaseResourceVersion,
		DraftHash:           input.DraftHash,
		ResolvedReferences:  references,
		ArtifactDigest:      input.ArtifactDigest,
		OptionsRevision:     input.OptionsRevision,
	})
}

func (input PlanHashInput) validate() error {
	if err := input.Target.Validate(input.Scope); err != nil {
		return err
	}
	if err := input.Operation.validate(); err != nil {
		return err
	}
	if err := validateBaseState(
		input.Operation,
		input.BaseUID,
		input.BaseResourceVersion,
	); err != nil {
		return err
	}
	if !digestPattern.MatchString(input.DraftHash) {
		return invalid("draftHash", "must be a lowercase SHA-256 digest")
	}
	if !digestPattern.MatchString(input.ArtifactDigest) {
		return invalid("artifactDigest", "must be a lowercase SHA-256 digest")
	}
	if err := validateOptionsRevision(input.OptionsRevision); err != nil {
		return err
	}
	return nil
}

func validateOptionsRevision(value string) error {
	if value == "" || strings.TrimSpace(value) != value ||
		!utf8.ValidString(value) ||
		utf8.RuneCountInString(value) > maxOptionsRevisionRunes {
		return invalid("optionsRevision", "must be a bounded revision token")
	}
	for _, character := range value {
		if unicode.IsControl(character) || unicode.Is(unicode.Bidi_Control, character) {
			return invalid("optionsRevision", "must not contain control characters")
		}
	}
	return nil
}

func sortedResolvedReferences(
	scope Scope,
	references []ResolvedReference,
) ([]ResolvedReference, error) {
	sorted := append([]ResolvedReference(nil), references...)
	for index := range sorted {
		if err := sorted[index].Validate(scope); err != nil {
			return nil, err
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].sortKey() < sorted[j].sortKey()
	})
	for index := 1; index < len(sorted); index++ {
		if sorted[index-1].duplicateKey() == sorted[index].duplicateKey() {
			return nil, invalid("resolvedReferences", "must not contain duplicate identity revisions")
		}
	}
	return sorted, nil
}
