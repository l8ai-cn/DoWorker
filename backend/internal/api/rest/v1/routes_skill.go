package v1

import (
	"github.com/gin-gonic/gin"
)

// registerSkillRoutes mounts the unified skill catalog routes (authoring,
// import from external git repos, upstream sync). No-op when the skill
// service is nil (git-backing / packager not configured).
func registerSkillRoutes(rg *gin.RouterGroup, svc *Services) {
	if svc.Skill == nil {
		return
	}
	h := NewSkillHandler(svc.Skill)
	skills := rg.Group("/authored-skills")
	{
		skills.GET("", h.ListSkills)
		skills.POST("", h.CreateSkill)
		skills.POST("/import", h.ImportSkills)
		skills.GET("/:skillSlug", h.GetSkill)
		skills.PATCH("/:skillSlug", h.UpdateSkill)
		skills.DELETE("/:skillSlug", h.DeleteSkill)
		skills.POST("/:skillSlug/sync-upstream", h.SyncSkillUpstream)
		skills.GET("/:skillSlug/tree", h.GetSkillTree)
		skills.GET("/:skillSlug/files/*path", h.GetSkillFile)
	}
}
