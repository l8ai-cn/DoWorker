package automation

// gemini-cli has no fine-grained permission CONFIG in its builtin AgentFile
// yet, so it uses the default translation: autonomous ⇒ MODE acp (guaranteeing
// non-interactive execution), interactive/auto_edit leave MODE untouched. When
// gemini gains a native approval knob, give it a dedicated adapter here.
func init() { register("gemini-cli", defaultAdapter{}) }
