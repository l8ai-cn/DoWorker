package runner

import (
	"fmt"
	"os"
)

func openSandboxRegularFile(
	root *os.Root,
	path string,
) (*os.File, os.FileInfo, error) {
	file, err := openSandboxFileForRead(root, path)
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, nil, fmt.Errorf("not a regular file")
	}
	return file, info, nil
}
