package workbench

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func resolveDeclaredRepresentations(
	workspacePath string,
	declarations []artifactDeclarationRepresentation,
) ([]declaredArtifactRepresentation, []string, error) {
	root, err := os.OpenRoot(workspacePath)
	if err != nil {
		return nil, nil, err
	}
	defer root.Close()
	representations := make(
		[]declaredArtifactRepresentation,
		0,
		len(declarations),
	)
	paths := make([]string, 0, len(declarations))
	seenIDs := make(map[string]struct{}, len(declarations))
	seenPaths := make(map[string]struct{}, len(declarations))
	for index, declaration := range declarations {
		if len(declaration.RepresentationID) < 2 ||
			len(declaration.RepresentationID) > 100 ||
			!artifactDeclarationIdentifier.MatchString(declaration.RepresentationID) {
			return nil, nil, fmt.Errorf(
				"representations[%d].representation_id is invalid",
				index,
			)
		}
		if _, exists := seenIDs[declaration.RepresentationID]; exists {
			return nil, nil, fmt.Errorf(
				"representation_id %q is duplicated",
				declaration.RepresentationID,
			)
		}
		seenIDs[declaration.RepresentationID] = struct{}{}
		display, err := validateDeclaredWorkspacePath(declaration.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("representations[%d].path: %w", index, err)
		}
		if _, exists := seenPaths[display]; exists {
			return nil, nil, fmt.Errorf("workspace path %q is duplicated", display)
		}
		seenPaths[display] = struct{}{}
		file, err := declaredArtifactFile(root, display, declaration.MediaType)
		if err != nil {
			return nil, nil, fmt.Errorf("representations[%d]: %w", index, err)
		}
		role := declaration.Role
		if role != "" && !validDeclarationLabel(role, 64) {
			return nil, nil, fmt.Errorf("representations[%d].role is invalid", index)
		}
		var dimensions *agentworkbenchv2.ArtifactDimensions
		if declaration.Dimensions != nil {
			if declaration.Dimensions.Width == 0 || declaration.Dimensions.Height == 0 {
				return nil, nil, fmt.Errorf(
					"representations[%d].dimensions must be positive",
					index,
				)
			}
			dimensions = &agentworkbenchv2.ArtifactDimensions{
				Width: declaration.Dimensions.Width, Height: declaration.Dimensions.Height,
			}
		}
		representations = append(representations, declaredArtifactRepresentation{
			representationID: declaration.RepresentationID,
			role:             role,
			file:             file,
			dimensions:       dimensions,
			durationMillis:   declaration.DurationMillis,
		})
		paths = append(paths, display)
	}
	return representations, paths, nil
}

func validateDeclaredWorkspacePath(value string) (string, error) {
	if value == "" || strings.Contains(value, "\\") ||
		strings.ContainsRune(value, 0) || path.IsAbs(value) {
		return "", fmt.Errorf("must be a POSIX workspace-relative path")
	}
	clean := path.Clean(value)
	if clean != value || clean == "." || clean == ".." ||
		strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("must be normalized and stay inside the workspace")
	}
	for _, segment := range strings.Split(clean, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("contains an invalid path segment")
		}
	}
	return clean, nil
}

func declaredArtifactFile(
	root *os.Root,
	display string,
	declaredMediaType string,
) (artifactFile, error) {
	if err := rejectDeclaredSymlinks(root, display); err != nil {
		return artifactFile{}, err
	}
	info, err := root.Stat(filepath.FromSlash(display))
	if err != nil {
		return artifactFile{}, err
	}
	if !info.Mode().IsRegular() {
		return artifactFile{}, fmt.Errorf("workspace path %q is not a regular file", display)
	}
	actualMediaType := artifactMediaType(display)
	if actualMediaType == "" || actualMediaType != declaredMediaType {
		return artifactFile{}, fmt.Errorf(
			"media_type %q does not match workspace file type %q",
			declaredMediaType,
			actualMediaType,
		)
	}
	digest, err := artifactDigestFromRoot(root, display)
	if err != nil {
		return artifactFile{}, err
	}
	return artifactFile{
		path: display, filename: path.Base(display), mediaType: actualMediaType,
		digest: digest, byteSize: uint64(info.Size()),
	}, nil
}

func rejectDeclaredSymlinks(root *os.Root, display string) error {
	segments := strings.Split(display, "/")
	for index := range segments {
		current := filepath.FromSlash(strings.Join(segments[:index+1], "/"))
		info, err := root.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("workspace path %q contains a symbolic link", display)
		}
	}
	return nil
}

func artifactDigestFromRoot(root *os.Root, display string) (string, error) {
	file, err := root.Open(filepath.FromSlash(display))
	if err != nil {
		return "", err
	}
	defer file.Close()
	return artifactDigestReader(file)
}
