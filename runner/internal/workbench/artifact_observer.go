package workbench

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type artifactFile struct {
	path      string
	filename  string
	mediaType string
	digest    string
	byteSize  uint64
}

type emittedArtifact struct {
	file     artifactFile
	revision uint64
	deleted  bool
}

type ArtifactObserver struct {
	root             string
	baseline         map[string]artifactFile
	emitted          map[string]emittedArtifact
	declaredBaseline map[string]declaredArtifact
	declaredEmitted  map[string]emittedDeclaredArtifact
	reservedPaths    map[string]struct{}
}

func NewArtifactObserver(root string) (*ArtifactObserver, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve artifact workspace: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, fmt.Errorf("stat artifact workspace: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("artifact workspace is not a directory")
	}
	declaredBaseline, reservedPaths, err := scanArtifactDeclarations(absolute)
	if err != nil {
		return nil, fmt.Errorf("scan artifact declaration baseline: %w", err)
	}
	baseline, err := scanArtifactFiles(absolute, reservedPaths)
	if err != nil {
		return nil, fmt.Errorf("scan artifact baseline: %w", err)
	}
	return &ArtifactObserver{
		root:             absolute,
		baseline:         baseline,
		emitted:          make(map[string]emittedArtifact),
		declaredBaseline: declaredBaseline,
		declaredEmitted:  make(map[string]emittedDeclaredArtifact),
		reservedPaths:    reservedPaths,
	}, nil
}

func (o *ArtifactObserver) Scan() ([]*ArtifactDescriptor, error) {
	declared, reservedPaths, err := scanArtifactDeclarations(o.root)
	if err != nil {
		return nil, fmt.Errorf("scan workspace artifact declarations: %w", err)
	}
	for path := range reservedPaths {
		o.reservedPaths[path] = struct{}{}
	}
	current, err := scanArtifactFiles(o.root, o.reservedPaths)
	if err != nil {
		return nil, fmt.Errorf("scan workspace artifacts: %w", err)
	}
	descriptors, err := o.changedDeclaredArtifacts(declared)
	if err != nil {
		return nil, fmt.Errorf("validate workspace artifact declarations: %w", err)
	}
	paths := sortedArtifactPaths(current)
	for _, path := range paths {
		file := current[path]
		previous, wasEmitted := o.emitted[path]
		baseline, wasBaseline := o.baseline[path]
		if !artifactChanged(file, previous, wasEmitted, baseline, wasBaseline) {
			continue
		}
		revision := uint64(1)
		if wasEmitted {
			revision = previous.revision + 1
		}
		descriptors = append(descriptors, readyArtifactDescriptor(file, revision))
		o.emitted[path] = emittedArtifact{file: file, revision: revision}
	}
	deletedPaths := sortedEmittedPaths(o.emitted)
	for _, path := range deletedPaths {
		previous := o.emitted[path]
		if previous.deleted {
			continue
		}
		if _, exists := current[path]; exists {
			continue
		}
		revision := previous.revision + 1
		descriptors = append(
			descriptors,
			deletedArtifactDescriptor(previous.file, revision),
		)
		previous.revision = revision
		previous.deleted = true
		o.emitted[path] = previous
	}
	return descriptors, nil
}

func artifactChanged(
	current artifactFile,
	emitted emittedArtifact,
	wasEmitted bool,
	baseline artifactFile,
	wasBaseline bool,
) bool {
	if wasEmitted {
		return emitted.deleted || emitted.file.digest != current.digest
	}
	return !wasBaseline || baseline.digest != current.digest
}

func sortedArtifactPaths(files map[string]artifactFile) []string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func sortedEmittedPaths(files map[string]emittedArtifact) []string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
