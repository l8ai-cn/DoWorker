package workbench

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

const (
	maxArtifactDeclarations    = 256
	maxArtifactDeclarationSize = 1 << 20
)

var artifactDeclarationIdentifier = regexp.MustCompile(
	`^[a-z0-9]+(?:-[a-z0-9]+)*$`,
)

func scanArtifactDeclarations(
	root string,
) (map[string]declaredArtifact, map[string]struct{}, error) {
	directory := filepath.Join(
		root,
		filepath.FromSlash(artifactDeclarationDirectory),
	)
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return map[string]declaredArtifact{}, map[string]struct{}{}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read artifact declarations: %w", err)
	}
	if len(entries) > maxArtifactDeclarations {
		return nil, nil, fmt.Errorf("artifact declarations exceed %d", maxArtifactDeclarations)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	artifacts := make(map[string]declaredArtifact, len(entries))
	reservedPaths := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 ||
			filepath.Ext(entry.Name()) != ".json" {
			return nil, nil, fmt.Errorf(
				"artifact declaration %q must be a regular .json file",
				entry.Name(),
			)
		}
		path := filepath.Join(directory, entry.Name())
		declaration, err := readArtifactDeclaration(path)
		if err != nil {
			return nil, nil, fmt.Errorf("artifact declaration %q: %w", entry.Name(), err)
		}
		artifact, paths, err := resolveArtifactDeclaration(root, declaration)
		if err != nil {
			return nil, nil, fmt.Errorf("artifact declaration %q: %w", entry.Name(), err)
		}
		if _, exists := artifacts[artifact.artifactID]; exists {
			return nil, nil, fmt.Errorf(
				"artifact declaration id %q is duplicated",
				artifact.artifactID,
			)
		}
		artifacts[artifact.artifactID] = artifact
		for _, path := range paths {
			reservedPaths[path] = struct{}{}
		}
	}
	return artifacts, reservedPaths, nil
}

func readArtifactDeclaration(
	path string,
) (artifactDeclaration, error) {
	file, err := os.Open(path)
	if err != nil {
		return artifactDeclaration{}, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, maxArtifactDeclarationSize+1))
	if err != nil {
		return artifactDeclaration{}, err
	}
	if len(raw) > maxArtifactDeclarationSize {
		return artifactDeclaration{}, fmt.Errorf("file exceeds %d bytes", maxArtifactDeclarationSize)
	}
	var declaration artifactDeclaration
	if err := decodeStrictJSON(raw, &declaration); err != nil {
		return artifactDeclaration{}, err
	}
	return declaration, nil
}

func resolveArtifactDeclaration(
	root string,
	declaration artifactDeclaration,
) (declaredArtifact, []string, error) {
	if err := validateArtifactDeclarationHeader(declaration); err != nil {
		return declaredArtifact{}, nil, err
	}
	representations, paths, err := resolveDeclaredRepresentations(
		root,
		declaration.Representations,
	)
	if err != nil {
		return declaredArtifact{}, nil, err
	}
	byID := make(map[string]declaredArtifactRepresentation, len(representations))
	for _, representation := range representations {
		byID[representation.representationID] = representation
	}
	if _, exists := byID[declaration.PrimaryRepresentationID]; !exists {
		return declaredArtifact{}, nil, fmt.Errorf(
			"primary_representation_id %q does not exist",
			declaration.PrimaryRepresentationID,
		)
	}
	manifest, err := declaredArtifactManifest(declaration.Manifest, byID)
	if err != nil {
		return declaredArtifact{}, nil, err
	}
	originalRevision := declaration.Revision
	declaration.Revision = 0
	canonical, err := json.Marshal(declaration)
	if err != nil {
		return declaredArtifact{}, nil, err
	}
	hash := sha256.New()
	hash.Write(canonical)
	for _, representation := range representations {
		hash.Write([]byte{0})
		hash.Write([]byte(representation.file.digest))
	}
	return declaredArtifact{
		artifactID:              declaration.ArtifactID,
		revision:                originalRevision,
		role:                    declaration.Role,
		primaryRepresentationID: declaration.PrimaryRepresentationID,
		producer:                declaration.Producer,
		representations:         representations,
		manifest:                manifest,
		fingerprint:             fmt.Sprintf("sha256:%x", hash.Sum(nil)),
	}, paths, nil
}
