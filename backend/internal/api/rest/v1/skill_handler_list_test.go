package v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

type skillHandlerListService struct {
	skillHandlerService
	listAllOrgID int64
	listLimit    int
	listOffset   int
}

func (s *skillHandlerListService) List(
	_ context.Context,
	_ int64,
	limit, offset int,
) ([]skilldom.Skill, int64, error) {
	s.listLimit = limit
	s.listOffset = offset
	return []skilldom.Skill{{ID: 3, Slug: "paged"}}, 21, nil
}

func (s *skillHandlerListService) ListAll(
	_ context.Context,
	orgID int64,
) ([]skilldom.Skill, error) {
	s.listAllOrgID = orgID
	return []skilldom.Skill{{ID: 2, Slug: "audio"}, {ID: 5, Slug: "video"}}, nil
}

func TestListSkillsAllReturnsCompleteCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &skillHandlerListService{}
	handler := NewSkillHandler(service)
	context, recorder := newSkillTagContext(
		http.MethodGet,
		"/authored-skills?all=true",
		"",
	)

	handler.ListSkills(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, int64(7), service.listAllOrgID)
	assert.JSONEq(t, `{"skills":[{"id":2,"organization_id":null,"slug":"audio","display_name":"","description":"","license":"","tags":null,"is_active":false,"git_repo_path":"","default_branch":"","install_source":"","content_sha":"","storage_key":"","package_size":0,"version":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"},{"id":5,"organization_id":null,"slug":"video","display_name":"","description":"","license":"","tags":null,"is_active":false,"git_repo_path":"","default_branch":"","install_source":"","content_sha":"","storage_key":"","package_size":0,"version":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}],"total":2}`, recorder.Body.String())
}

func TestListSkillsKeepsPaginatedContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &skillHandlerListService{}
	handler := NewSkillHandler(service)
	context, recorder := newSkillTagContext(
		http.MethodGet,
		"/authored-skills?limit=10&offset=20",
		"",
	)

	handler.ListSkills(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 10, service.listLimit)
	assert.Equal(t, 20, service.listOffset)
	assert.JSONEq(t, `{"skills":[{"id":3,"organization_id":null,"slug":"paged","display_name":"","description":"","license":"","tags":null,"is_active":false,"git_repo_path":"","default_branch":"","install_source":"","content_sha":"","storage_key":"","package_size":0,"version":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}],"total":21,"limit":10,"offset":20}`, recorder.Body.String())
}
