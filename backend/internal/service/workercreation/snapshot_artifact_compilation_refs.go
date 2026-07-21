package workercreation

import (
	"sort"
	"strconv"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func (resolver *artifactCompilationResolver) compilationReferences(
	secretRefs map[string]specdomain.SecretReference,
) (compilationReferences, error) {
	var refs compilationReferences
	if repository := resolver.document.Repository; repository != nil {
		refs.RepositorySlug = repository.Pin.Reference.Name.String()
	}
	for _, skill := range resolver.document.Skills {
		refs.SkillSlugs = append(refs.SkillSlugs, skill.Slug.String())
	}
	for _, kb := range resolver.document.KnowledgeBases {
		refs.Knowledge = append(refs.Knowledge, knowledgeReference{
			Slug: kb.Slug.String(), Mode: kb.Mode,
		})
	}
	if err := resolver.appendBundleReferences(&refs, secretRefs); err != nil {
		return compilationReferences{}, err
	}
	return refs, nil
}

func (resolver *artifactCompilationResolver) appendBundleReferences(
	refs *compilationReferences,
	secretRefs map[string]specdomain.SecretReference,
) error {
	seen := map[string]struct{}{}
	for _, bundle := range resolver.document.RuntimeBundles {
		if bundle.ConfigDocument != nil {
			refs.ConfigDocumentIDs = append(refs.ConfigDocumentIDs, bundle.ConfigDocument.ID)
			continue
		}
		if err := appendUniqueArtifactBundleName(refs, seen, bundle.Pin.DomainID, bundle.Pin.Reference.Name.String()); err != nil {
			return err
		}
	}
	fields := make([]string, 0, len(secretRefs))
	for field := range secretRefs {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		for _, secret := range resolver.document.SecretReferences {
			if secret.Field != field {
				continue
			}
			if err := appendUniqueArtifactBundleName(refs, seen, secret.Pin.DomainID, secret.Pin.Reference.Name.String()); err != nil {
				return err
			}
		}
	}
	return nil
}

func appendUniqueArtifactBundleName(
	refs *compilationReferences,
	seen map[string]struct{},
	domainID int64,
	name string,
) error {
	key := strconv.FormatInt(domainID, 10)
	if _, exists := seen[key]; exists {
		return nil
	}
	if err := appendEnvBundleName(refs, name); err != nil {
		return err
	}
	seen[key] = struct{}{}
	return nil
}
