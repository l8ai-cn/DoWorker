package operatorcatalog

func skillDefinitions() []SkillDefinition {
	videoUse := ResearchSource{
		URL:     "https://github.com/browser-use/video-use",
		Commit:  "92c2b34e44c205cbc2acae7f6ca7c1c219d5dd66",
		License: "MIT",
	}
	return []SkillDefinition{
		{
			Slug: "seedance-expert", Name: "Seedance 视频生成",
			Description: "使用已绑定的 Seedance 视频模型生成 MP4，并输出为平台可预览的 Worker 成果。",
			License:     "Apache-2.0", Tags: []string{"video", "seedance", "generation"},
		},
		{
			Slug: "short-video-directing", Name: "短视频编导",
			Description: "从传播目标产出脚本、镜头表、连续性说明和剪辑简报。",
			License:     "Apache-2.0", Tags: []string{"video", "short-video", "directing"},
		},
		{
			Slug: "video-editing-workflow", Name: "视频剪辑工作流",
			Description: "基于 EDL、FFmpeg、字幕与音频规则完成可回溯剪辑。",
			License:     "Apache-2.0", Tags: []string{"video", "editing", "ffmpeg"},
			ResearchSources: []ResearchSource{videoUse},
		},
		{
			Slug: "remotion-video-production", Name: "Remotion 视频制作",
			Description: "用 React 与 Remotion 构建确定性视频合成并完成渲染。",
			License:     "Apache-2.0", Tags: []string{"video", "production", "remotion"},
		},
		{
			Slug: "video-motion-graphics", Name: "视频动效设计",
			Description: "为字幕、信息卡、产品演示和叙事节点设计可读动效。",
			License:     "Apache-2.0", Tags: []string{"video", "motion", "graphics"},
		},
		{
			Slug: "video-delivery-qa", Name: "视频交付质检",
			Description: "用媒体探测、抽帧、听检和字幕检查验证最终成片。",
			License:     "Apache-2.0", Tags: []string{"video", "qa", "delivery"},
		},
		{
			Slug: "media-rights-research", Name: "媒体素材版权调研",
			Description: "检索可用素材并保存来源、许可、署名与使用边界。",
			License:     "Apache-2.0", Tags: []string{"video", "research", "rights"},
		},
	}
}
