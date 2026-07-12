package testkit

func executionClusterTableDDLs() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS execution_clusters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			slug TEXT NOT NULL,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(organization_id, slug)
		)`,
	}
}
