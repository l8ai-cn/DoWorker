// Startup prompt for a KB ingest pod — the llm-wiki maintenance run:
// compile new raw/ material into wiki/ pages per the KB's AGENTS.md schema.
export function buildKbIngestPrompt(kbSlug: string): string {
  return [
    `知识库 ${kbSlug} 已以读写模式挂载到 kb/${kbSlug}/。`,
    "请执行一次 ingest 维护：",
    `1. 阅读 kb/${kbSlug}/AGENTS.md 了解本知识库的维护规范；`,
    `2. 阅读 kb/${kbSlug}/llms.txt 与 wiki/index.md 了解现有结构；`,
    "3. 检查 raw/ 中尚未编入 wiki 的资料，将其提炼、更新到对应 wiki 页面，维护交叉引用；",
    "4. 同步更新 llms.txt 索引与 wiki/log.md 变更日志；",
    "5. 完成后 git commit 并 push。",
  ].join("\n");
}
