package runner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
)

var verifiedArtifactSHA256 = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

type verifiedArtifactRead struct {
	ArtifactID       string `json:"artifact_id"`
	Digest           string `json:"digest"`
	FileBytes        uint64 `json:"file_bytes"`
	RepresentationID string `json:"representation_id"`
	Revision         uint64 `json:"revision"`
}

func parseVerifiedArtifactRead(payload string) (verifiedArtifactRead, error) {
	var request verifiedArtifactRead
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		return request, fmt.Errorf("invalid artifact verification request")
	}
	if request.ArtifactID == "" || request.RepresentationID == "" ||
		request.Revision == 0 || request.FileBytes > math.MaxInt64 ||
		!verifiedArtifactSHA256.MatchString(request.Digest) {
		return request, fmt.Errorf("invalid artifact verification request")
	}
	return request, nil
}

func readVerifiedArtifactRange(
	file *os.File,
	offset int64,
	length int64,
	request verifiedArtifactRead,
) ([]byte, int64, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	fileBytes := info.Size()
	if fileBytes != int64(request.FileBytes) || offset > fileBytes {
		return nil, 0, fmt.Errorf("artifact integrity mismatch")
	}
	result := make([]byte, min(length, fileBytes-offset))
	hash := sha256.New()
	buffer := make([]byte, 64<<10)
	var position int64
	for {
		count, readErr := file.Read(buffer)
		if count > 0 {
			chunk := buffer[:count]
			_, _ = hash.Write(chunk)
			copyArtifactRange(result, chunk, position, offset)
			position += int64(count)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, 0, readErr
		}
	}
	actualDigest := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	if position != fileBytes || actualDigest != request.Digest {
		return nil, 0, fmt.Errorf("artifact integrity mismatch")
	}
	return result, fileBytes, nil
}

func copyArtifactRange(
	target []byte,
	chunk []byte,
	chunkStart int64,
	rangeStart int64,
) {
	chunkEnd := chunkStart + int64(len(chunk))
	rangeEnd := rangeStart + int64(len(target))
	start := max(chunkStart, rangeStart)
	end := min(chunkEnd, rangeEnd)
	if start >= end {
		return
	}
	copy(
		target[start-rangeStart:end-rangeStart],
		chunk[start-chunkStart:end-chunkStart],
	)
}
