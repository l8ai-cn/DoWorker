package v1

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	skillSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/skill"
)

type skillHandlerTagService struct {
	skillHandlerService
	createRequest *skillSvc.CreateSkillRequest
	updateRequest *skillSvc.UpdateSkillRequest
	serviceErr    error
}

func (s *skillHandlerTagService) Create(
	_ context.Context,
	req *skillSvc.CreateSkillRequest,
) (*skilldom.Skill, error) {
	s.createRequest = req
	if s.serviceErr != nil {
		return nil, s.serviceErr
	}
	return &skilldom.Skill{ID: 1, Slug: req.Slug, Tags: skilldom.NormalizeTags(req.Tags)}, nil
}

func (s *skillHandlerTagService) Get(
	_ context.Context,
	_ int64,
	slug string,
) (*skilldom.Skill, error) {
	return &skilldom.Skill{ID: 1, Slug: slug}, nil
}

func (s *skillHandlerTagService) Update(
	_ context.Context,
	req *skillSvc.UpdateSkillRequest,
) (*skilldom.Skill, error) {
	s.updateRequest = req
	if s.serviceErr != nil {
		return nil, s.serviceErr
	}
	return &skilldom.Skill{
		ID: req.SkillID, Slug: "video-editing", Tags: skilldom.NormalizeTags(*req.Tags),
	}, nil
}

func (s *skillHandlerTagService) ImportFromGit(
	_ context.Context,
	_ *skillSvc.ImportFromGitRequest,
) ([]*skilldom.Skill, error) {
	return []*skilldom.Skill{{ID: 1, Slug: "valid"}}, s.serviceErr
}

func (s *skillHandlerTagService) SyncFromUpstream(
	_ context.Context,
	_ int64,
	_ string,
) (*skilldom.Skill, error) {
	return nil, s.serviceErr
}

func newSkillTagContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("tenant", &middleware.TenantContext{OrganizationID: 7, UserID: 11})
	context.Params = gin.Params{{Key: "skillSlug", Value: "video-editing"}}
	return context, recorder
}

func TestCreateSkillAcceptsTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &skillHandlerTagService{}
	handler := NewSkillHandler(service)
	context, recorder := newSkillTagContext(
		http.MethodPost,
		"/authored-skills",
		`{"slug":"video-editing","name":"Video Editing","instructions":"Use ffmpeg.","tags":[" Video ","editing"]}`,
	)

	handler.CreateSkill(context)

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.NotNil(t, service.createRequest)
	assert.Equal(t, []string{" Video ", "editing"}, service.createRequest.Tags)
	assert.JSONEq(t, `{"skill":{"id":1,"organization_id":null,"slug":"video-editing","display_name":"","description":"","license":"","tags":["editing","video"],"is_active":false,"git_repo_path":"","default_branch":"","install_source":"","content_sha":"","storage_key":"","package_size":0,"version":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`, recorder.Body.String())
}

func TestUpdateSkillAcceptsTags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &skillHandlerTagService{}
	handler := NewSkillHandler(service)
	context, recorder := newSkillTagContext(
		http.MethodPatch,
		"/authored-skills/video-editing",
		`{"tags":[" Motion ","editing","MOTION"]}`,
	)

	handler.UpdateSkill(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, service.updateRequest)
	require.NotNil(t, service.updateRequest.Tags)
	assert.Equal(t, []string{" Motion ", "editing", "MOTION"}, *service.updateRequest.Tags)
	assert.JSONEq(t, `{"skill":{"id":1,"organization_id":null,"slug":"video-editing","display_name":"","description":"","license":"","tags":["editing","motion"],"is_active":false,"git_repo_path":"","default_branch":"","install_source":"","content_sha":"","storage_key":"","package_size":0,"version":0,"created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`, recorder.Body.String())
}

func TestInvalidTagsMapToBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewSkillHandler(&skillHandlerTagService{serviceErr: skillSvc.ErrInvalidTags})
	context, recorder := newSkillTagContext(
		http.MethodPatch,
		"/authored-skills/video-editing",
		`{"tags":["invalid"]}`,
	)

	handler.UpdateSkill(context)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestImportAndSyncInvalidTagsMapToBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewSkillHandler(&skillHandlerTagService{serviceErr: skillSvc.ErrInvalidTags})

	importContext, importRecorder := newSkillTagContext(
		http.MethodPost,
		"/authored-skills/import",
		`{"url":"https://example.test/skills.git"}`,
	)
	handler.ImportSkills(importContext)
	assert.Equal(t, http.StatusBadRequest, importRecorder.Code)

	syncContext, syncRecorder := newSkillTagContext(
		http.MethodPost,
		"/authored-skills/video-editing/sync-upstream",
		"",
	)
	handler.SyncSkillUpstream(syncContext)
	assert.Equal(t, http.StatusBadRequest, syncRecorder.Code)
}
