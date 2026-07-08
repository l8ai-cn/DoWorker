package v1

import (
	"github.com/gin-gonic/gin"
)

// registerSkillRoutes mounts the git-backed author-in-platform skill routes.
// No-op when the skill service is nil (git-backing / packager not configured),
// so it is fully additive to the existing external-import skill flow.
func registerSkillRoutes(rg *gin.RouterGroup, svc *Services) {
	if svc.Skill == nil {
		return
	}
	h := NewSkillHandler(svc.Skill)
	skills := rg.Group("/authored-skills")
	{
		skills.GET("", h.ListSkills)
		skills.POST("", h.CreateSkill)
		skills.GET("/:skillSlug", h.GetSkill)
		skills.PATCH("/:skillSlug", h.UpdateSkill)
		skills.DELETE("/:skillSlug", h.DeleteSkill)
		skills.GET("/:skillSlug/tree", h.GetSkillTree)
		skills.GET("/:skillSlug/files/*path", h.GetSkillFile)
	}
}
