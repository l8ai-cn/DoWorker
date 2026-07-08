package codeximport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// convertOutputDir synthesizes a short request/result conversation from a
// workflow output_* directory. Unlike a rollout it is not a turn-by-turn
// transcript, so we reconstruct one user prompt (the workflow request) and one
// assistant summary (what the run produced) from the JSON manifests.
func convertOutputDir(dir string) (*Result, error) {
	res := &Result{
		Kind:       KindOutputDir,
		SourcePath: dir,
		SourceID:   filepath.Base(dir),
	}

	input := readJSONMap(filepath.Join(dir, "conversation_input.json"))
	manifest := readJSONMap(filepath.Join(dir, "run_manifest.json"))
	checkpoint := readJSONMap(filepath.Join(dir, "workflow_checkpoint.json"))

	target := firstString(manifest["target"], input["target"])
	if target == "" {
		target = res.SourceID
	}
	res.Title = truncateTitle(fmt.Sprintf("Codex 工作流：%s", target))

	// User turn: the original workflow request.
	userText := buildRequestText(input, manifest, target)
	res.Items = append(res.Items, Item{
		Type:       "message",
		Status:     "completed",
		StartsTurn: true,
		Payload: map[string]any{
			"type":    "message",
			"role":    "user",
			"content": []map[string]any{{"type": "input_text", "text": userText}},
		},
	})

	// Assistant turn: a summary of what the run produced.
	assistantText := buildSummaryText(manifest, checkpoint, dir)
	res.Items = append(res.Items, Item{
		Type:   "message",
		Status: "completed",
		Payload: map[string]any{
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]any{{"type": "output_text", "text": assistantText}},
		},
	})

	return res, nil
}

func buildRequestText(input, manifest map[string]any, target string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "生成视频：%s\n", target)
	if v := firstString(manifest["age"]); v != "" {
		fmt.Fprintf(&b, "- 目标年龄：%s\n", v)
	}
	if v := firstString(manifest["video_style"]); v != "" {
		fmt.Fprintf(&b, "- 视频风格：%s\n", v)
	}
	if v := firstString(input["new_ip_ideas"]); v != "" {
		fmt.Fprintf(&b, "- IP 设定：%s\n", v)
	}
	if v := firstString(input["scene_video_provider_preference"], manifest["scene_video_provider"]); v != "" {
		fmt.Fprintf(&b, "- 分镜视频提供方：%s\n", v)
	}
	if v := firstString(input["requested_scene_video_model"], manifest["scene_video_model_requested_by_user"]); v != "" {
		fmt.Fprintf(&b, "- 期望视频模型：%s\n", v)
	}
	return strings.TrimRight(b.String(), "\n")
}

func buildSummaryText(manifest, checkpoint map[string]any, dir string) string {
	var b strings.Builder
	b.WriteString("已完成本次 Codex 视频工作流。\n\n")

	if v := firstString(manifest["ip_name"]); v != "" {
		fmt.Fprintf(&b, "- IP：%s\n", v)
	}
	if v := firstString(manifest["reasoning_model"]); v != "" {
		fmt.Fprintf(&b, "- 推理模型：%s\n", v)
	}
	if v := firstString(manifest["image_generation_provider"]); v != "" {
		fmt.Fprintf(&b, "- 图像生成：%s\n", v)
	}
	if v := firstString(manifest["scene_video_provider"]); v != "" {
		model := firstString(manifest["scene_video_model"])
		if model != "" {
			fmt.Fprintf(&b, "- 分镜视频：%s（%s）\n", v, model)
		} else {
			fmt.Fprintf(&b, "- 分镜视频：%s\n", v)
		}
	}
	if v := firstString(checkpoint["stage"]); v != "" {
		fmt.Fprintf(&b, "- 当前阶段：%s\n", v)
	}

	images := countStringSlice(checkpoint["generated_image_paths"])
	videos := countStringSlice(checkpoint["generated_scene_video_files"])
	tts := countStringSlice(checkpoint["generated_tts_files"])
	if images+videos+tts > 0 {
		b.WriteString("\n产物统计：\n")
		if images > 0 {
			fmt.Fprintf(&b, "- 关键帧/图片：%d\n", images)
		}
		if videos > 0 {
			fmt.Fprintf(&b, "- 分镜视频：%d\n", videos)
		}
		if tts > 0 {
			fmt.Fprintf(&b, "- 配音片段：%d\n", tts)
		}
	}

	if scenes := listSceneImages(dir); len(scenes) > 0 {
		b.WriteString("\n主要关键帧：\n")
		for _, s := range scenes {
			fmt.Fprintf(&b, "- %s\n", s)
		}
	}

	fmt.Fprintf(&b, "\n完整产物目录：%s", dir)
	return strings.TrimRight(b.String(), "\n")
}

// listSceneImages returns the scene_N.png basenames present in dir, ordered by
// scene number, capped so the summary stays readable.
func listSceneImages(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "scene_") && strings.HasSuffix(n, ".png") && !strings.Contains(n, "_raw") && !strings.Contains(n, "_candidate") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	const cap = 15
	if len(names) > cap {
		names = names[:cap]
	}
	return names
}

func readJSONMap(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]any{}
	}
	return m
}

func firstString(vals ...any) string {
	for _, v := range vals {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func countStringSlice(v any) int {
	if arr, ok := v.([]any); ok {
		return len(arr)
	}
	return 0
}
