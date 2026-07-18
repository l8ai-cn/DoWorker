package runner

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const officePreviewConversionTimeout = 2 * time.Minute

type officePreviewConverter func(
	ctx context.Context,
	workDir, relativePath string,
) ([]byte, error)

func convertOfficePreview(
	parent context.Context,
	workDir, relativePath string,
) ([]byte, error) {
	source, err := resolveOfficePreviewSourcePath(workDir, relativePath)
	if err != nil {
		return nil, err
	}
	soffice, err := exec.LookPath("soffice")
	if err != nil {
		return nil, fmt.Errorf("LibreOffice soffice executable is required: %w", err)
	}
	tempDir, err := os.MkdirTemp("", "agentsmesh-office-preview-*")
	if err != nil {
		return nil, fmt.Errorf("create office preview directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	profileDir := filepath.Join(tempDir, "profile")
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return nil, fmt.Errorf("create office preview output: %w", err)
	}
	ctx, cancel := context.WithTimeout(parent, officePreviewConversionTimeout)
	defer cancel()
	profileURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(profileDir),
	}).String()
	command := exec.CommandContext(
		ctx,
		soffice,
		"--headless",
		"--nologo",
		"--nodefault",
		"--nolockcheck",
		"--nofirststartwizard",
		"-env:UserInstallation="+profileURL,
		"--convert-to",
		"pdf",
		"--outdir",
		outputDir,
		source,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("office preview conversion timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf(
			"office preview conversion failed: %w: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}
	pdfPath := filepath.Join(
		outputDir,
		strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))+".pdf",
	)
	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("read office preview PDF: %w", err)
	}
	if len(pdf) == 0 {
		return nil, fmt.Errorf("office preview PDF is empty")
	}
	return pdf, nil
}

func resolveOfficePreviewSourcePath(
	workDir, relativePath string,
) (string, error) {
	if relativePath == "" || filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("office preview source path is invalid")
	}
	root, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve work directory: %w", err)
	}
	source, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		return "", fmt.Errorf("resolve office preview source: %w", err)
	}
	relative, err := filepath.Rel(root, source)
	if err != nil || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("office preview source escapes work directory")
	}
	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("stat office preview source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("office preview source is not a regular file")
	}
	return source, nil
}
