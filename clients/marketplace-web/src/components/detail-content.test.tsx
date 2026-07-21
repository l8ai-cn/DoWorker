import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DetailContent } from "./detail-content";

describe("DetailContent", () => {
  it("explains the real package and first task before acquisition", () => {
    render(
      <DetailContent
        listing={{
          listing_id: "1",
          listing_version_id: "11",
          slug: "software-delivery-expert",
          resource_type: "application",
          display_name: "软件交付专家",
          tagline: "完成可验证的软件交付",
          publisher: { slug: "agent-cloud", display_name: "Agent Cloud", verified: true },
          spaces: [{ slug: "software-delivery", name: "软件交付" }],
          tags: [],
          published_at: "2026-07-12T08:00:00Z",
          description: "将任务拆成可验证的代码交付。",
          outcomes: ["完成可审查的变更"],
          use_cases: ["修复缺陷"],
          target_audience: ["交付工程师"],
          requirements: ["需要可用 Runner"],
          permissions: ["读取仓库"],
          version: "v1",
          release_notes: "首个公开版本。",
          package_summary: ["代码审查工作流", "端到端验证 Skill"],
          first_task: {
            title: "创建首个交付任务",
            description: "输入代码仓库和验收目标后开始。",
          },
        }}
      />,
    );

    expect(screen.getByRole("heading", { name: "应用包含什么" })).toBeInTheDocument();
    expect(screen.getByText("代码审查工作流")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "启用后从这里开始" })).toBeInTheDocument();
    expect(screen.getByText("创建首个交付任务")).toBeInTheDocument();
  });
});
