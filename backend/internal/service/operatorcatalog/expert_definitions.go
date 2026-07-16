package operatorcatalog

func Experts() []ExpertDefinition {
	return []ExpertDefinition{
		{
			Slug: "video-production-expert", Name: "视频制作专家",
			Summary:     "从创意到可交付成片的一体化视频制作。",
			Description: "负责短视频方案、素材、Remotion 合成、动效和交付质检。",
			Category:    "video", Icon: "clapperboard",
			Tags:     []string{"video", "production", "remotion", "short-video"},
			Outcomes: []string{"成片方案", "可复现工程", "平台规格成片", "质检报告"},
			SkillSlugs: []string{
				"short-video-directing", "media-rights-research",
				"remotion-video-production", "video-motion-graphics",
				"video-delivery-qa",
			},
			Prompt: "Own the video from brief to verified master. Confirm creative direction before rendering, preserve asset rights evidence, and do not deliver an unverified file.",
		},
		{
			Slug: "video-editing-expert", Name: "视频剪辑专家",
			Summary:     "以叙事、节奏、声音和字幕为核心完成专业剪辑。",
			Description: "分析素材，确认剪辑策略，生成 EDL，完成动效、字幕、声音和交付检查。",
			Category:    "video", Icon: "scissors",
			Tags:     []string{"video", "editing", "ffmpeg", "subtitles"},
			Outcomes: []string{"剪辑策略", "EDL", "预览片", "交付母版", "质检报告"},
			SkillSlugs: []string{
				"video-editing-workflow", "video-motion-graphics",
				"video-delivery-qa",
			},
			Prompt: "Edit from an explicit strategy and reversible EDL. Preserve source media, keep subtitles on the output timeline, and verify every cut before delivery.",
		},
		{
			Slug: "short-video-director", Name: "短视频编导专家",
			Summary:     "把传播目标转成可拍、可剪、可验证的短视频执行方案。",
			Description: "完成选题提炼、结构、脚本、分镜、镜头表、素材计划和剪辑简报。",
			Category:    "video", Icon: "film",
			Tags:     []string{"video", "directing", "script", "storyboard"},
			Outcomes: []string{"创意简报", "成稿脚本", "镜头表", "连续性说明", "剪辑简报"},
			SkillSlugs: []string{
				"short-video-directing", "media-rights-research",
				"video-delivery-qa",
			},
			Prompt: "Turn the objective into a feasible short-video plan. Ground claims in supplied facts, identify every visual source, and hand production an executable script and shot list.",
		},
	}
}
