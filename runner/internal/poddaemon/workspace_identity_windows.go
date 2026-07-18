//go:build windows

package poddaemon

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

const workspaceIdentityKind = "windows-volume-file-v1"

func openWorkspaceIdentityFile(path string) (*os.File, error) {
	return openWindowsWorkspaceFile(
		path,
		windows.FILE_SHARE_READ|
			windows.FILE_SHARE_WRITE|
			windows.FILE_SHARE_DELETE,
	)
}

func openWorkspaceLaunchFile(path string) (*os.File, error) {
	return openWindowsWorkspaceFile(
		path,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
	)
}

func openWindowsWorkspaceFile(path string, shareMode uint32) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(
		name,
		windows.FILE_READ_ATTRIBUTES,
		shareMode,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		_ = windows.CloseHandle(handle)
		return nil, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		_ = windows.CloseHandle(handle)
		return nil, fmt.Errorf("workspace path is a reparse point")
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		_ = windows.CloseHandle(handle)
		return nil, fmt.Errorf("workspace path is not a directory")
	}
	return os.NewFile(uintptr(handle), path), nil
}

func workspaceFileIdentity(file *os.File) (string, uint64, uint64, error) {
	var info windows.ByHandleFileInformation
	err := windows.GetFileInformationByHandle(
		windows.Handle(file.Fd()),
		&info,
	)
	if err != nil {
		return "", 0, 0, fmt.Errorf("stat workspace identity: %w", err)
	}
	fileID := uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow)
	return workspaceIdentityKind, uint64(info.VolumeSerialNumber), fileID, nil
}
