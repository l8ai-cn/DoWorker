package workerspec

type KnowledgeMountMode string

const (
	KnowledgeMountReadOnly  KnowledgeMountMode = "ro"
	KnowledgeMountReadWrite KnowledgeMountMode = "rw"
)

type KnowledgeMount struct {
	KnowledgeBaseID int64              `json:"knowledge_base_id"`
	Mode            KnowledgeMountMode `json:"mode"`
}

type RuntimeEnvBundleID int64

type SkillPackageBinding struct {
	SkillID     int64  `json:"skill_id"`
	Slug        string `json:"slug"`
	Version     int    `json:"version"`
	ContentSHA  string `json:"content_sha"`
	StorageKey  string `json:"storage_key"`
	PackageSize int64  `json:"package_size"`
}

type Workspace struct {
	RepositoryID    *int64                `json:"repository_id,omitempty"`
	Branch          string                `json:"branch"`
	SkillIDs        []int64               `json:"skill_ids"`
	SkillPackages   []SkillPackageBinding `json:"skill_packages"`
	KnowledgeMounts []KnowledgeMount      `json:"knowledge_mounts"`
	EnvBundleIDs    []RuntimeEnvBundleID  `json:"env_bundle_ids"`
	Instructions    string                `json:"instructions"`
	InitialTask     string                `json:"initial_task"`
}
