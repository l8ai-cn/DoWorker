package runner

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func officePreviewSourceDigest(workDir, relativePath string) (string, error) {
	source, err := resolveOfficePreviewSourcePath(workDir, relativePath)
	if err != nil {
		return "", err
	}
	file, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

func writeOfficePreviewArtifact(
	workDir string,
	artifactID string,
	revision uint64,
	pdf []byte,
) (string, string, uint64, error) {
	if len(pdf) == 0 {
		return "", "", 0, fmt.Errorf("office preview PDF is empty")
	}
	artifactHash := sha256.Sum256([]byte(artifactID))
	relative := filepath.ToSlash(filepath.Join(
		".agent-cloud",
		"workbench",
		"previews",
		fmt.Sprintf("%x-r%d.pdf", artifactHash[:12], revision),
	))
	target, err := resolveOfficePreviewTargetPath(workDir, relative)
	if err != nil {
		return "", "", 0, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return "", "", 0, fmt.Errorf("create office preview storage: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".preview-*.pdf")
	if err != nil {
		return "", "", 0, fmt.Errorf("create office preview file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return "", "", 0, err
	}
	if _, err := temp.Write(pdf); err != nil {
		temp.Close()
		return "", "", 0, fmt.Errorf("write office preview file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return "", "", 0, fmt.Errorf("close office preview file: %w", err)
	}
	if err := os.Rename(tempPath, target); err != nil {
		return "", "", 0, fmt.Errorf("publish office preview file: %w", err)
	}
	digest := sha256.Sum256(pdf)
	return relative, fmt.Sprintf("sha256:%x", digest[:]), uint64(len(pdf)), nil
}

func resolveOfficePreviewTargetPath(
	workDir string,
	relativePath string,
) (string, error) {
	root, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve work directory: %w", err)
	}
	target, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		return "", fmt.Errorf("resolve office preview target: %w", err)
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("office preview target escapes work directory")
	}
	return target, nil
}
