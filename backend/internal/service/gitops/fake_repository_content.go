package gitops

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
)

func (r *fakeRepo) put(path string, content []byte) {
	buf := make([]byte, len(content))
	copy(buf, content)
	r.Files[path] = buf
	sum := sha1.Sum(content)
	r.SHAs[path] = hex.EncodeToString(sum[:])
}

func underDir(path, dir string) (string, bool) {
	if dir == "" {
		return path, true
	}
	prefix := dir + "/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	return path[len(prefix):], true
}

func sortedEntries(entries map[string]Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
