package workbench

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

type PublishedArtifactDeclaration struct {
	ArtifactID      string `json:"artifact_id"`
	Revision        uint64 `json:"revision"`
	DeclarationPath string `json:"declaration_path"`
}

func PublishArtifactDeclaration(
	workspace string,
	raw json.RawMessage,
) (*PublishedArtifactDeclaration, error) {
	if len(raw) > maxArtifactDeclarationSize {
		return nil, fmt.Errorf("artifact declaration exceeds %d bytes", maxArtifactDeclarationSize)
	}
	var declaration artifactDeclaration
	if err := decodeStrictJSON(raw, &declaration); err != nil {
		return nil, err
	}
	_, _, err := resolveArtifactDeclaration(workspace, declaration)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(workspace)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	declarationDirectory := filepath.FromSlash(artifactDeclarationDirectory)
	if err := root.MkdirAll(declarationDirectory, 0o755); err != nil {
		return nil, err
	}
	if err := rejectDeclaredSymlinks(root, artifactDeclarationDirectory); err != nil {
		return nil, err
	}
	declarationPath := path.Join(
		artifactDeclarationDirectory,
		declaration.ArtifactID+".json",
	)
	changed, err := validatePublishedTransition(root, declarationPath, declaration)
	if err != nil {
		return nil, err
	}
	published := &PublishedArtifactDeclaration{
		ArtifactID:      declaration.ArtifactID,
		Revision:        declaration.Revision,
		DeclarationPath: declarationPath,
	}
	if !changed {
		return published, nil
	}
	encoded, err := json.MarshalIndent(declaration, "", "  ")
	if err != nil {
		return nil, err
	}
	encoded = append(encoded, '\n')
	if err := writePublishedDeclaration(root, declarationPath, encoded); err != nil {
		return nil, err
	}
	return published, nil
}

func validatePublishedTransition(
	root *os.Root,
	declarationPath string,
	current artifactDeclaration,
) (bool, error) {
	raw, err := root.ReadFile(filepath.FromSlash(declarationPath))
	if os.IsNotExist(err) {
		if current.Revision != 1 {
			return false, fmt.Errorf("first revision must be 1")
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if len(raw) > maxArtifactDeclarationSize {
		return false, fmt.Errorf("existing artifact declaration exceeds %d bytes", maxArtifactDeclarationSize)
	}
	var declaration artifactDeclaration
	if err := decodeStrictJSON(raw, &declaration); err != nil {
		return false, err
	}
	if err := validateArtifactDeclarationHeader(declaration); err != nil {
		return false, err
	}
	if declaration.Producer != current.Producer {
		return false, fmt.Errorf("producer must remain stable across revisions")
	}
	if current.Revision == declaration.Revision {
		previousCanonical, marshalErr := json.Marshal(declaration)
		if marshalErr != nil {
			return false, marshalErr
		}
		currentCanonical, marshalErr := json.Marshal(current)
		if marshalErr != nil {
			return false, marshalErr
		}
		if string(previousCanonical) == string(currentCanonical) {
			return false, nil
		}
	}
	expectedRevision := declaration.Revision + 1
	if current.Revision != expectedRevision {
		return false, fmt.Errorf(
			"changed artifact revision must be %d",
			expectedRevision,
		)
	}
	return true, nil
}

func writePublishedDeclaration(
	root *os.Root,
	declarationPath string,
	content []byte,
) error {
	workbenchDirectory := path.Dir(artifactDeclarationDirectory)
	tempPath := path.Join(
		workbenchDirectory,
		".artifact-"+uuid.NewString()+".tmp",
	)
	temp, err := root.OpenFile(
		filepath.FromSlash(tempPath),
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o644,
	)
	if err != nil {
		return err
	}
	defer root.Remove(filepath.FromSlash(tempPath))
	if _, err := temp.Write(content); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := root.Rename(
		filepath.FromSlash(tempPath),
		filepath.FromSlash(declarationPath),
	); err != nil {
		return err
	}
	directory, err := root.Open(filepath.FromSlash(artifactDeclarationDirectory))
	if err != nil {
		return err
	}
	defer directory.Close()
	if runtime.GOOS == "windows" {
		return nil
	}
	return directory.Sync()
}
