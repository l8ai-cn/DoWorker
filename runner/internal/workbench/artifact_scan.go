package workbench

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func scanArtifactFiles(
	root string,
	excluded map[string]struct{},
) (map[string]artifactFile, error) {
	files := make(map[string]artifactFile)
	err := filepath.WalkDir(root, func(
		path string,
		entry os.DirEntry,
		walkErr error,
	) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		display := filepath.ToSlash(relative)
		if entry.IsDir() {
			if ignoredArtifactPath(display) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || ignoredArtifactPath(display) {
			return nil
		}
		if _, isExcluded := excluded[display]; isExcluded {
			return nil
		}
		mediaType := artifactMediaType(display)
		if mediaType == "" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		digest, err := artifactDigest(path)
		if err != nil {
			return err
		}
		files[display] = artifactFile{
			path:      display,
			filename:  entry.Name(),
			mediaType: mediaType,
			digest:    digest,
			byteSize:  uint64(info.Size()),
		}
		return nil
	})
	return files, err
}

func artifactDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return artifactDigestReader(file)
}

func artifactDigestReader(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

func ignoredArtifactPath(path string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(path), "/") {
		if strings.HasPrefix(segment, ".") {
			return true
		}
		switch segment {
		case "node_modules", "target", "vendor":
			return true
		}
	}
	return false
}

func artifactMediaType(path string) string {
	extension := strings.TrimPrefix(
		strings.ToLower(filepath.Ext(path)),
		".",
	)
	if mediaType := artifactMediaTypes[extension]; mediaType != "" {
		return mediaType
	}
	if !artifactDeliverableRoot(path) {
		return ""
	}
	return artifactTextMediaTypes[extension]
}

func ArtifactMediaType(path string) string {
	return artifactMediaType(path)
}

func artifactDeliverableRoot(path string) bool {
	root, _, _ := strings.Cut(filepath.ToSlash(path), "/")
	switch root {
	case "artifacts", "deliverables", "output", "outputs":
		return true
	default:
		return false
	}
}

var artifactMediaTypes = map[string]string{
	"3mf": "model/3mf", "aac": "audio/aac", "avif": "image/avif",
	"blend": "application/x-blender", "gif": "image/gif",
	"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"flac": "audio/flac",
	"glb":  "model/gltf-binary", "gltf": "model/gltf+json",
	"htm": "text/html", "html": "text/html",
	"jpeg": "image/jpeg", "jpg": "image/jpeg",
	"m4a": "audio/mp4", "m4v": "video/x-m4v", "mov": "video/quicktime",
	"mp3": "audio/mpeg", "mp4": "video/mp4",
	"pdf": "application/pdf", "png": "image/png",
	"ppt":  "application/vnd.ms-powerpoint",
	"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"step": "model/step", "stl": "model/stl", "stp": "model/step",
	"svg": "image/svg+xml", "wav": "audio/wav", "webm": "video/webm",
	"webp": "image/webp", "xls": "application/vnd.ms-excel",
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

var artifactTextMediaTypes = map[string]string{
	"cjs": "text/javascript", "css": "text/css", "csv": "text/csv",
	"go": "text/x-go",
	"js": "text/javascript", "jsx": "text/javascript", "json": "application/json",
	"md": "text/markdown", "mjs": "text/javascript", "py": "text/x-python",
	"rs": "text/x-rust", "scad": "text/plain", "scss": "text/x-scss",
	"sh": "text/x-shellscript", "toml": "text/plain", "ts": "text/typescript",
	"tsx": "text/typescript", "txt": "text/plain",
	"yaml": "text/yaml", "yml": "text/yaml",
}
