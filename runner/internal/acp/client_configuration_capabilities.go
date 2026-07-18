package acp

func (c *ACPClient) captureTransportCapabilities() {
	permissionModes := c.transport.SupportedPermissionModes()
	artifactActions := c.transport.SupportedArtifactActions()
	if len(permissionModes) == 0 && len(artifactActions) == 0 {
		return
	}
	c.configMu.Lock()
	defer c.configMu.Unlock()
	if len(permissionModes) > 0 {
		c.configuration.SupportedPermissionModes = append(
			[]string(nil),
			permissionModes...,
		)
	}
	if len(artifactActions) > 0 {
		c.configuration.SupportedArtifactActions = append(
			[]string(nil),
			artifactActions...,
		)
	}
}

func (c *ACPClient) SupportedPermissionModes() []string {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	return append([]string(nil), c.configuration.SupportedPermissionModes...)
}

func (c *ACPClient) SeedConfiguration(configuration Configuration) {
	c.configMu.Lock()
	defer c.configMu.Unlock()
	if configuration.PermissionMode != "" {
		c.configuration.PermissionMode = configuration.PermissionMode
	}
	if configuration.Model != "" {
		c.configuration.Model = configuration.Model
	}
	if len(configuration.SupportedPermissionModes) > 0 {
		c.configuration.SupportedPermissionModes = append(
			[]string(nil),
			configuration.SupportedPermissionModes...,
		)
	}
	if len(configuration.SupportedArtifactActions) > 0 {
		c.configuration.SupportedArtifactActions = append(
			[]string(nil),
			configuration.SupportedArtifactActions...,
		)
	}
}
