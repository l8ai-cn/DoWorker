package codeximport

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

const sampleRollout = `{"timestamp":"t","type":"session_meta","payload":{"session_id":"019f40af-abcd","cwd":"/tmp"}}
{"timestamp":"t","type":"event_msg","payload":{"type":"task_started"}}
{"timestamp":"t","type":"response_item","payload":{"type":"message","role":"developer","content":[{"type":"input_text","text":"SYSTEM PROMPT NOISE"}]}}
{"timestamp":"t","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"帮我安装这个 skill"}]}}
{"timestamp":"t","type":"response_item","payload":{"type":"reasoning","id":"rs_1","summary":[],"encrypted_content":"opaque"}}
{"timestamp":"t","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"好的，我先看看包结构。"}]}}
{"timestamp":"t","type":"response_item","payload":{"type":"function_call","id":"fc_1","name":"exec_command","arguments":"{\"cmd\":\"ls\"}","call_id":"call_1"}}
{"timestamp":"t","type":"response_item","payload":{"type":"function_call_output","call_id":"call_1","output":"file1\nfile2"}}
{"timestamp":"t","type":"response_item","payload":{"type":"custom_tool_call","id":"ctc_1","name":"apply_patch","input":"*** Begin Patch","call_id":"call_2"}}
{"timestamp":"t","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_2","output":"patched"}}
{"timestamp":"t","type":"response_item","payload":{"type":"image_generation_call","id":"ig_1","status":"completed","revised_prompt":"a squirrel teacher","result":"BASE64_SHOULD_BE_DROPPED"}}
`

func TestConvertRollout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rollout-2026-07-08T15-44-00-019f40af.jsonl")
	writeFile(t, path, sampleRollout)

	res, err := Convert(path)
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if res.Kind != KindRollout {
		t.Fatalf("kind = %q, want %q", res.Kind, KindRollout)
	}
	if res.SourceID != "019f40af-abcd" {
		t.Fatalf("source id = %q", res.SourceID)
	}
	if res.Title != "帮我安装这个 skill" {
		t.Fatalf("title = %q", res.Title)
	}

	// developer message + reasoning are dropped -> 6 items remain.
	wantTypes := []string{
		"message", // user
		"message", // assistant
		"function_call",
		"function_call_output",
		"function_call", // custom_tool_call mapped
		"function_call_output",
		"image_generation_call",
	}
	if len(res.Items) != len(wantTypes) {
		t.Fatalf("got %d items, want %d: %+v", len(res.Items), len(wantTypes), res.Items)
	}
	for i, want := range wantTypes {
		if res.Items[i].Type != want {
			t.Errorf("item[%d].Type = %q, want %q", i, res.Items[i].Type, want)
		}
	}

	// Only the user message starts a turn.
	if !res.Items[0].StartsTurn {
		t.Errorf("user message should start a turn")
	}
	if res.Items[1].StartsTurn {
		t.Errorf("assistant message should not start a turn")
	}

	// custom_tool_call arguments come from Codex "input".
	if got := res.Items[4].Payload["arguments"]; got != "*** Begin Patch" {
		t.Errorf("custom tool arguments = %v", got)
	}
	// image result base64 must be dropped.
	if _, ok := res.Items[6].Payload["result"]; ok {
		t.Errorf("image_generation_call result should be dropped")
	}
	if got := res.Items[6].Payload["revised_prompt"]; got != "a squirrel teacher" {
		t.Errorf("revised_prompt = %v", got)
	}
}

func TestConvertOutputDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "conversation_input.json"), `{"target":"松鼠老师讲分数","new_ip_ideas":"松鼠老师；儿童教育","requested_scene_video_model":"Seedance 2.5"}`)
	writeFile(t, filepath.Join(dir, "run_manifest.json"), `{"target":"松鼠老师讲分数","age":"5-8岁","video_style":"儿童教育动画","ip_name":"松鼠老师","reasoning_model":"gpt-5.4","scene_video_provider":"lovart","scene_video_model":"lovart-kling-2.6"}`)
	writeFile(t, filepath.Join(dir, "workflow_checkpoint.json"), `{"stage":"review_keyframes","generated_image_paths":["a.png","b.png"],"generated_scene_video_files":["v1.mp4"]}`)
	writeFile(t, filepath.Join(dir, "scene_1.png"), "x")
	writeFile(t, filepath.Join(dir, "scene_2.png"), "x")
	writeFile(t, filepath.Join(dir, "scene_1_raw.png"), "x")

	res, err := Convert(dir)
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if res.Kind != KindOutputDir {
		t.Fatalf("kind = %q", res.Kind)
	}
	if len(res.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(res.Items))
	}
	if role, _ := res.Items[0].Payload["role"].(string); role != "user" {
		t.Errorf("item0 role = %q", role)
	}
	if !res.Items[0].StartsTurn {
		t.Errorf("first item should start a turn")
	}
	if role, _ := res.Items[1].Payload["role"].(string); role != "assistant" {
		t.Errorf("item1 role = %q", role)
	}
	summary := firstText(res.Items[1].Payload)
	for _, want := range []string{"松鼠老师", "gpt-5.4", "关键帧/图片：2", "分镜视频：1", "scene_1.png"} {
		if !contains(summary, want) {
			t.Errorf("summary missing %q; got:\n%s", want, summary)
		}
	}
	// _raw scene images are excluded from the keyframe list.
	if contains(summary, "scene_1_raw.png") {
		t.Errorf("summary should not include raw scene image")
	}
}

func TestDetect(t *testing.T) {
	dir := t.TempDir()

	// A plain .jsonl file -> rollout.
	jsonl := filepath.Join(dir, "x.jsonl")
	writeFile(t, jsonl, "{}\n")
	if k, p, err := Detect(jsonl); err != nil || k != KindRollout || p != jsonl {
		t.Fatalf("Detect(file) = %q %q %v", k, p, err)
	}

	// A directory holding a rollout-*.jsonl -> rollout of that file.
	rdir := t.TempDir()
	roll := filepath.Join(rdir, "rollout-1.jsonl")
	writeFile(t, roll, "{}\n")
	writeFile(t, filepath.Join(rdir, "conversation_input.json"), "{}")
	if k, p, err := Detect(rdir); err != nil || k != KindRollout || p != roll {
		t.Fatalf("Detect(rollout dir) = %q %q %v", k, p, err)
	}

	// A workflow output dir (no rollout) -> output dir.
	odir := t.TempDir()
	writeFile(t, filepath.Join(odir, "run_manifest.json"), "{}")
	if k, p, err := Detect(odir); err != nil || k != KindOutputDir || p != odir {
		t.Fatalf("Detect(output dir) = %q %q %v", k, p, err)
	}

	// Unrecognized.
	edir := t.TempDir()
	if _, _, err := Detect(edir); err == nil {
		t.Fatalf("Detect(empty dir) should error")
	}
	if _, _, err := Detect(filepath.Join(dir, "missing")); err == nil {
		t.Fatalf("Detect(missing) should error")
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
