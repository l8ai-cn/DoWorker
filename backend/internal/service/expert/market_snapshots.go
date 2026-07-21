package expert

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/lib/pq"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func encodeMarketSnapshots(
	source *expertdom.Expert,
	specSnapshot specdomain.Snapshot,
	skills []skilldom.Skill,
) (json.RawMessage, json.RawMessage, json.RawMessage, error) {
	metadata := source.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}
	marketSpec, err := marketWorkerSpec(specSnapshot.Spec, skills)
	if err != nil {
		return nil, nil, nil, errors.Join(ErrMarketSnapshotInvalid, err)
	}
	if err := validateExpertMatchesWorkerSpec(source, marketSpec, skills); err != nil {
		return nil, nil, nil, errors.Join(ErrMarketSnapshotInvalid, err)
	}
	prompt := optionalMarketPrompt(marketSpec.Workspace.Instructions)
	snapshot := marketExpertSnapshot{
		Version:         1,
		Slug:            source.Slug,
		Name:            source.Name,
		Description:     source.Description,
		AgentSlug:       marketSpec.Runtime.WorkerType.Slug.String(),
		Prompt:          prompt,
		InteractionMode: string(marketSpec.TypeConfig.InteractionMode),
		AutomationLevel: string(marketSpec.TypeConfig.AutomationLevel),
		Perpetual:       false,
		UsedEnvBundles:  []string{},
		SkillSlugs:      marketSkillSlugs(skills),
		KnowledgeMounts: []expertdom.KnowledgeMount{},
		ConfigOverrides: normalizedMarketConfig(marketSpec.TypeConfig.Values),
		Metadata:        append(json.RawMessage(nil), metadata...),
	}
	if err := validateMarketExpertSnapshot(snapshot); err != nil {
		return nil, nil, nil, errors.Join(ErrMarketSnapshotInvalid, err)
	}
	expertSnapshot, err := json.Marshal(snapshot)
	if err != nil {
		return nil, nil, nil, err
	}
	summary, err := specdomain.Summarize(marketSpec)
	if err != nil {
		return nil, nil, nil, errors.Join(ErrMarketSnapshotInvalid, err)
	}
	workerSnapshot, err := json.Marshal(marketWorkerSpecSnapshot{
		Version: 1,
		Spec:    marketSpec,
		Summary: summary,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	dependencies := make(
		[]MarketSkillDependency,
		0,
		len(marketSpec.Workspace.SkillPackages),
	)
	for _, pkg := range marketSpec.Workspace.SkillPackages {
		dependencies = append(dependencies, MarketSkillDependency{
			SkillID:     pkg.SkillID,
			Slug:        pkg.Slug,
			Version:     pkg.Version,
			ContentSHA:  pkg.ContentSHA,
			StorageKey:  pkg.StorageKey,
			PackageSize: pkg.PackageSize,
		})
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Slug < dependencies[j].Slug
	})
	dependencySnapshot, err := json.Marshal(dependencies)
	return expertSnapshot, workerSnapshot, dependencySnapshot, err
}

func decodeMarketReleaseSnapshots(
	release *expertmarket.Release,
) (marketExpertSnapshot, marketWorkerSpecSnapshot, error) {
	if !validMarketIcon(release.Icon) {
		return marketExpertSnapshot{}, marketWorkerSpecSnapshot{},
			ErrMarketSnapshotInvalid
	}
	var expertSnapshot marketExpertSnapshot
	if err := decodeStrictJSON(release.ExpertSnapshot, &expertSnapshot); err != nil {
		return marketExpertSnapshot{}, marketWorkerSpecSnapshot{},
			errors.Join(ErrMarketSnapshotInvalid, err)
	}
	var workerSnapshot marketWorkerSpecSnapshot
	if err := decodeStrictJSON(release.WorkerSpecSnapshot, &workerSnapshot); err != nil {
		return marketExpertSnapshot{}, marketWorkerSpecSnapshot{},
			errors.Join(ErrMarketSnapshotInvalid, err)
	}
	if expertSnapshot.Version != 1 || workerSnapshot.Version != 1 {
		return marketExpertSnapshot{}, marketWorkerSpecSnapshot{},
			ErrMarketSnapshotInvalid
	}
	if err := validateMarketExpertSnapshot(expertSnapshot); err != nil {
		return marketExpertSnapshot{}, marketWorkerSpecSnapshot{},
			errors.Join(ErrMarketSnapshotInvalid, err)
	}
	spec, err := specdomain.NormalizeAndValidate(workerSnapshot.Spec)
	if err != nil {
		return marketExpertSnapshot{}, marketWorkerSpecSnapshot{},
			errors.Join(ErrMarketSnapshotInvalid, err)
	}
	summary, err := specdomain.Summarize(spec)
	if err != nil || !reflect.DeepEqual(summary, workerSnapshot.Summary) {
		return marketExpertSnapshot{}, marketWorkerSpecSnapshot{},
			ErrMarketSnapshotInvalid
	}
	workerSnapshot.Spec = spec
	return expertSnapshot, workerSnapshot, nil
}

func marketReleaseUpdate(
	snapshot marketExpertSnapshot,
	workerSpecSnapshotID, releaseID int64,
	expertType string,
) expertdom.MarketReleaseUpdate {
	config, _ := json.Marshal(snapshot.ConfigOverrides)
	mounts, _ := json.Marshal(snapshot.KnowledgeMounts)
	return expertdom.MarketReleaseUpdate{
		Name:                  strings.TrimSpace(snapshot.Name),
		Description:           snapshot.Description,
		AgentSlug:             strings.TrimSpace(snapshot.AgentSlug),
		Prompt:                snapshot.Prompt,
		InteractionMode:       snapshot.InteractionMode,
		AutomationLevel:       snapshot.AutomationLevel,
		Perpetual:             snapshot.Perpetual,
		UsedEnvBundles:        pq.StringArray(cloneMarketStrings(snapshot.UsedEnvBundles)),
		SkillSlugs:            pq.StringArray(cloneMarketStrings(snapshot.SkillSlugs)),
		KnowledgeMounts:       mounts,
		ConfigOverrides:       config,
		Metadata:              mergeMetadata(snapshot.Metadata, nil, &expertType),
		WorkerSpecSnapshotID:  workerSpecSnapshotID,
		SourceMarketReleaseID: releaseID,
	}
}

func cloneMarketStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func marketSkillSlugs(skills []skilldom.Skill) []string {
	slugs := make([]string, 0, len(skills))
	for _, skill := range skills {
		slugs = append(slugs, skill.Slug)
	}
	sort.Strings(slugs)
	return slugs
}

func decodeStrictJSON(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ErrMarketSnapshotInvalid
		}
		return err
	}
	return nil
}
