package gitops

import (
	"fmt"
	"strings"
)

// repoName maps (orgID, slug) -> "org<ID>-<slug>". Slugs are unique per Do
// Worker org, so the org prefix disambiguates repos that share one namespace.
func repoName(orgID int64, slug string) string {
	return fmt.Sprintf("org%d-%s", orgID, slug)
}

// repoPath maps (namespace, orgID, slug) -> "<namespace>/org<ID>-<slug>",
// the value persisted as git_repo_path on domain rows.
func repoPath(namespace string, orgID int64, slug string) string {
	return namespace + "/" + repoName(orgID, slug)
}

// repoNameFromPath strips the namespace prefix off a stored git_repo_path,
// yielding the bare repo name the gitea contents API expects.
func repoNameFromPath(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}
