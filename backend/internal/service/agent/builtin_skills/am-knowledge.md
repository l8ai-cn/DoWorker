---
name: am-knowledge
description: |
  WHEN to use:
  - 任务需要参考组织知识库（团队约定、产品文档、运行手册等）
  - 学到了新的可复用知识，需要沉淀回知识库
  - 需要跨知识库检索信息
  - 收到 ingest 任务：把 raw/ 中的新资料编译进 wiki/

  WHEN NOT to use:
  - 与组织知识无关的独立编码任务
user-invocable: false
---

# 知识库（llm-wiki）工作流

组织知识库是标准布局的 Git 仓库。挂载到本 Pod 的知识库位于 `kb/{slug}/`
（清单见 `kb/README.md`）；未挂载的知识库通过 MCP 工具访问。

## 仓库布局

```
llms.txt        # 索引：名称 + 摘要 + 分节链接列表 —— 永远先读这个
AGENTS.md       # schema：页面类型、ingest 流程、交叉引用与矛盾标记规范
raw/            # 不可变原始资料（只进不改）
wiki/
  index.md      # 总览
  log.md        # 变更日志（## [YYYY-MM-DD] operation | Title）
  ...           # 实体页/概念页/综述页
```

## 可用工具

- `kb_list` - 列出组织内可见的知识库
- `kb_search` - 跨库检索 wiki 页面（大小写不敏感的文本匹配）
- `kb_read` - 读取单个文件（未挂载的库用它读 llms.txt / wiki 页）
- `kb_write` - 经平台提交单个文件（用于未挂载或只读挂载的库）

## 读取知识

1. 已挂载：直接读 `kb/{slug}/llms.txt` 导航，再按链接读 wiki 页面。
2. 未挂载：`kb_list` → `kb_read(kb_slug, "llms.txt")` → 按需读页面。
3. 不确定在哪个库：`kb_search(query="...")` 跨库检索。

## 写回知识（wiki 维护者职责）

1. 先读目标库的 `AGENTS.md`，遵循其页面结构与命名规范。
2. rw 挂载：直接编辑 `kb/{slug}/wiki/...`，更新 `wiki/log.md`，
   然后 `git add && git commit && git push`（remote 已带推送凭证）。
3. 未挂载 / 只读挂载：用 `kb_write` 提交（commit 会标注本 Pod）。
4. 原则：
   - `raw/` 不可变 —— 新资料只追加，不修改已有文件
   - `wiki/` 是编译产物 —— 保持页面间交叉引用，矛盾处按 AGENTS.md 标记
   - 每次写回都在 `wiki/log.md` 记一行变更日志

## Ingest 任务

收到"把新资料编入知识库"类任务时：

1. 读 `AGENTS.md` 了解 ingest 工作流与页面 schema
2. 读 `raw/` 中的新资料
3. 更新/新建相关 wiki 页面，维护 `llms.txt` 索引和交叉引用
4. 记录 `wiki/log.md`，commit + push
