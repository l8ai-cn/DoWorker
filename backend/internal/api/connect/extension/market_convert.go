package extensionconnect

import (
	"encoding/json"

	extdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/extension"
	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	"github.com/l8ai-cn/agentcloud/backend/pkg/protoconv"
	extensionv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/extension/v1"
)

// toProtoSkillMarketItem projects a unified-catalog skill row onto the
// marketplace wire shape. The proto id is the skills-table row id (clients
// pass it back to InstallSkillFromMarket). registry_id is retired and stays 0.
func toProtoSkillMarketItem(m *skilldom.Skill) *extensionv1.SkillMarketItem {
	if m == nil {
		return nil
	}
	out := &extensionv1.SkillMarketItem{
		Id:            m.ID,
		Slug:          m.Slug,
		DisplayName:   m.DisplayName,
		Description:   m.Description,
		License:       m.License,
		Compatibility: m.Compatibility,
		AllowedTools:  m.AllowedTools,
		Category:      m.Category,
		ContentSha:    m.ContentSha,
		StorageKey:    m.StorageKey,
		PackageSize:   m.PackageSize,
		Version:       int32(m.Version),
		IsActive:      m.IsActive,
		CreatedAt:     protoconv.RFC3339(m.CreatedAt),
		UpdatedAt:     protoconv.RFC3339(m.UpdatedAt),
	}
	if filter := m.GetAgentFilter(); len(filter) > 0 {
		out.AgentFilter = filter
	}
	return out
}

// toProtoMcpMarketItem mirrors REST's serialization of McpMarketItem. JSON
// raw fields (default_args, default_http_headers, registry_meta) are passed
// through as strings — the renderer parses them locally.
func toProtoMcpMarketItem(m *extdom.McpMarketItem) *extensionv1.McpMarketItem {
	if m == nil {
		return nil
	}
	out := &extensionv1.McpMarketItem{
		Id:                 m.ID,
		Slug:               m.Slug,
		Name:               m.Name,
		Description:        m.Description,
		Icon:               m.Icon,
		TransportType:      m.TransportType,
		Command:            m.Command,
		DefaultArgs:        jsonRawString(m.DefaultArgs, "[]"),
		DefaultHttpUrl:     m.DefaultHttpURL,
		DefaultHttpHeaders: jsonRawString(m.DefaultHttpHeaders, "[]"),
		Category:           m.Category,
		IsActive:           m.IsActive,
		Source:             m.Source,
		RegistryName:       m.RegistryName,
		Version:            m.Version,
		RepositoryUrl:      m.RepositoryURL,
		RegistryMeta:       jsonRawString(m.RegistryMeta, "{}"),
		CreatedAt:          protoconv.RFC3339(m.CreatedAt),
		UpdatedAt:          protoconv.RFC3339(m.UpdatedAt),
	}
	if filter := m.GetAgentFilter(); len(filter) > 0 {
		out.AgentFilter = filter
	}
	if len(m.EnvVarSchema) > 0 {
		var entries []extdom.EnvVarSchemaEntry
		if err := json.Unmarshal(m.EnvVarSchema, &entries); err == nil {
			out.EnvVarSchema = make([]*extensionv1.McpEnvVarSchemaEntry, 0, len(entries))
			for _, e := range entries {
				out.EnvVarSchema = append(out.EnvVarSchema, &extensionv1.McpEnvVarSchemaEntry{
					Name:        e.Name,
					Label:       e.Label,
					Required:    e.Required,
					Sensitive:   e.Sensitive,
					Placeholder: e.Placeholder,
				})
			}
		}
	}
	if m.LastSyncedAt != nil {
		out.LastSyncedAt = protoconv.RFC3339Ptr(m.LastSyncedAt)
	}
	return out
}

// toProtoInstalledSkill converts a domain InstalledSkill. market_item is
// intentionally not surfaced — the renderer fetches catalog rows separately.
func toProtoInstalledSkill(s *extdom.InstalledSkill) *extensionv1.InstalledSkill {
	if s == nil {
		return nil
	}
	out := &extensionv1.InstalledSkill{
		Id:             s.ID,
		OrganizationId: s.OrganizationID,
		RepositoryId:   s.RepositoryID,
		Scope:          s.Scope,
		Slug:           s.Slug,
		InstallSource:  s.InstallSource,
		SourceUrl:      s.SourceURL,
		ContentSha:     s.ContentSha,
		StorageKey:     s.StorageKey,
		PackageSize:    s.PackageSize,
		IsEnabled:      s.IsEnabled,
		CreatedAt:      protoconv.RFC3339(s.CreatedAt),
		UpdatedAt:      protoconv.RFC3339(s.UpdatedAt),
	}
	if s.SkillID != nil {
		// Wire field predates the unified catalog; it now carries the
		// skills-table row id.
		out.MarketItemId = s.SkillID
	}
	if s.InstalledBy != nil {
		out.InstalledBy = s.InstalledBy
	}
	if s.PinnedVersion != nil {
		v := int32(*s.PinnedVersion)
		out.PinnedVersion = &v
	}
	return out
}

// toProtoInstalledMcpServer converts a domain InstalledMcpServer. EnvVars
// are emitted as decrypted-at-read JSON (the service layer decrypts before
// it returns the value); we trust that contract here and pass through.
func toProtoInstalledMcpServer(s *extdom.InstalledMcpServer) *extensionv1.InstalledMcpServer {
	if s == nil {
		return nil
	}
	out := &extensionv1.InstalledMcpServer{
		Id:             s.ID,
		OrganizationId: s.OrganizationID,
		RepositoryId:   s.RepositoryID,
		Scope:          s.Scope,
		Name:           s.Name,
		Slug:           s.Slug,
		TransportType:  s.TransportType,
		Command:        s.Command,
		Args:           jsonRawString(s.Args, "[]"),
		HttpUrl:        s.HttpURL,
		HttpHeaders:    jsonRawString(s.HttpHeaders, "{}"),
		EnvVars:        jsonRawString(s.EnvVars, "{}"),
		IsEnabled:      s.IsEnabled,
		CreatedAt:      protoconv.RFC3339(s.CreatedAt),
		UpdatedAt:      protoconv.RFC3339(s.UpdatedAt),
	}
	if s.MarketItemID != nil {
		out.MarketItemId = s.MarketItemID
	}
	if s.InstalledBy != nil {
		out.InstalledBy = s.InstalledBy
	}
	return out
}

// jsonRawString returns raw JSON as a string with a fallback for empty bytes.
// The proto wire encodes these as `string` (conventions: opaque user JSON
// stays opaque between server and client; only the renderer parses).
func jsonRawString(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 {
		return fallback
	}
	return string(raw)
}
