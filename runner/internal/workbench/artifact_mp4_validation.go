package workbench

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maximumMP4BoxCount = 1024

func validateDeclaredFileContent(
	root *os.Root,
	display string,
	mediaType string,
) error {
	if mediaType != "video/mp4" {
		return nil
	}
	file, err := root.Open(filepath.FromSlash(display))
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() < 24 || !isMP4Container(file, info.Size()) {
		return invalidMP4Error(display)
	}
	return nil
}

func isMP4Container(file *os.File, fileSize int64) bool {
	var foundFileType, foundMediaData, foundVideoTrack bool
	for offset, boxes := int64(0), 0; offset+8 <= fileSize &&
		boxes < maximumMP4BoxCount; boxes++ {
		boxType, payloadStart, end, ok := readMP4Box(file, offset, fileSize)
		if !ok {
			return false
		}
		switch boxType {
		case "ftyp":
			foundFileType = end-payloadStart >= 8
		case "mdat":
			foundMediaData = end > payloadStart
		case "moov":
			foundVideoTrack = hasMP4VideoTrack(file, payloadStart, end)
		}
		offset = end
		if foundFileType && foundMediaData && foundVideoTrack {
			return true
		}
	}
	return false
}

func readMP4Box(
	file *os.File,
	offset int64,
	limit int64,
) (string, int64, int64, bool) {
	if offset < 0 || offset+8 > limit {
		return "", 0, 0, false
	}
	header := make([]byte, 16)
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return "", 0, 0, false
	}
	if _, err := io.ReadFull(file, header[:8]); err != nil {
		return "", 0, 0, false
	}
	size := uint64(binary.BigEndian.Uint32(header[:4]))
	headerSize := uint64(8)
	if size == 1 {
		if offset+16 > limit {
			return "", 0, 0, false
		}
		if _, err := io.ReadFull(file, header[8:16]); err != nil {
			return "", 0, 0, false
		}
		size = binary.BigEndian.Uint64(header[8:16])
		headerSize = 16
	}
	if size == 0 {
		size = uint64(limit - offset)
	}
	if size < headerSize || size > uint64(limit-offset) {
		return "", 0, 0, false
	}
	return string(header[4:8]), offset + int64(headerSize),
		offset + int64(size), true
}

func hasMP4VideoTrack(file *os.File, start int64, end int64) bool {
	for offset, boxes := start, 0; offset+8 <= end &&
		boxes < maximumMP4BoxCount; boxes++ {
		boxType, payloadStart, boxEnd, ok := readMP4Box(file, offset, end)
		if !ok {
			return false
		}
		if boxType == "trak" && mp4TrackIsVideo(file, payloadStart, boxEnd) {
			return true
		}
		offset = boxEnd
	}
	return false
}

func mp4TrackIsVideo(file *os.File, start int64, end int64) bool {
	for offset, boxes := start, 0; offset+8 <= end &&
		boxes < maximumMP4BoxCount; boxes++ {
		boxType, payloadStart, boxEnd, ok := readMP4Box(file, offset, end)
		if !ok {
			return false
		}
		if boxType == "mdia" {
			return mp4MediaHandlerIsVideo(file, payloadStart, boxEnd)
		}
		offset = boxEnd
	}
	return false
}

func mp4MediaHandlerIsVideo(file *os.File, start int64, end int64) bool {
	for offset, boxes := start, 0; offset+8 <= end &&
		boxes < maximumMP4BoxCount; boxes++ {
		boxType, payloadStart, boxEnd, ok := readMP4Box(file, offset, end)
		if !ok {
			return false
		}
		if boxType == "hdlr" {
			if boxEnd-payloadStart < 12 {
				return false
			}
			handler := make([]byte, 4)
			if _, err := file.ReadAt(handler, payloadStart+8); err != nil {
				return false
			}
			return string(handler) == "vide"
		}
		offset = boxEnd
	}
	return false
}

func invalidMP4Error(display string) error {
	return fmt.Errorf("workspace path %q is not a valid MP4 file", display)
}
