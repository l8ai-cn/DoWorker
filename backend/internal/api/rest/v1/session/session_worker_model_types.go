package sessionapi

// workerModelMount is ephemeral per-session model injection (not persisted).
type workerModelMount struct {
	ConfigBundles map[string]interface{}
	EnvBundles    map[string]map[string]string
}

func (m *workerModelMount) mounted() bool {
	return m != nil && (len(m.ConfigBundles) > 0 || len(m.EnvBundles) > 0)
}
