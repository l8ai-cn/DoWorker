package expertmarket

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestApplicationBeforeSaveValidatesSlug(t *testing.T) {
	for _, slug := range []string{"Bad.Slug", "bad_slug", "A", "中文"} {
		t.Run(slug, func(t *testing.T) {
			app := Application{Slug: slugkit.Slug(slug)}
			require.Error(t, app.BeforeSave(&gorm.DB{}))
		})
	}

	app := Application{Slug: slugkit.Slug("video-production")}
	require.NoError(t, app.BeforeSave(&gorm.DB{}))
}
