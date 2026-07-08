package sessionapi

func configBundlesFromMount(m *workerModelMount) map[string]interface{} {
	if m == nil {
		return nil
	}
	return m.ConfigBundles
}

func envBundlesFromMount(m *workerModelMount) map[string]map[string]string {
	if m == nil {
		return nil
	}
	return m.EnvBundles
}
