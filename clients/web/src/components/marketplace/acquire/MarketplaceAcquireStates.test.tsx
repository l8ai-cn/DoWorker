import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";

import { OrganizationStep, SuccessState } from "./MarketplaceAcquireStates";

describe("MarketplaceAcquireStates", () => {
  it("sends a successful acquisition to its application instance", () => {
    const props = {
      organization: { id: 9, slug: "dev-org", name: "研发组织" },
      installationID: "installation-1",
    };
    render(
      <SuccessState {...props} />,
    );

    expect(screen.getByRole("link", { name: "去应用中心开始第一个任务" }))
      .toHaveAttribute("href", "/dev-org/applications/installation-1");
  });

  it("requires a compatible model before checking installation conditions", async () => {
    const user = userEvent.setup();
    const onModelChange = vi.fn();
    const { rerender } = render(
      <OrganizationStep
        organizations={[{ id: 9, slug: "dev-org", name: "研发组织" }]}
        loadingOrganizations={false}
        value="9"
        onChange={vi.fn()}
        onContinue={vi.fn()}
        modelResources={[{ id: 301, label: "OpenAI · GPT-5" }]}
        modelResourceID=""
        onModelChange={onModelChange}
        loadingModels={false}
        modelError={false}
        incompatibleListing={false}
        onReloadModels={vi.fn()}
        settingsHref="/dev-org/settings?tab=ai-resources"
      />,
    );

    expect(screen.getByRole("button", { name: "检查启用条件" })).toBeDisabled();
    await user.selectOptions(screen.getByLabelText("选择运行模型"), "301");
    expect(onModelChange).toHaveBeenCalledWith("301");

    rerender(
      <OrganizationStep
        organizations={[{ id: 9, slug: "dev-org", name: "研发组织" }]}
        loadingOrganizations={false}
        value="9"
        onChange={vi.fn()}
        onContinue={vi.fn()}
        modelResources={[{ id: 301, label: "OpenAI · GPT-5" }]}
        modelResourceID="301"
        onModelChange={onModelChange}
        loadingModels={false}
        modelError={false}
        incompatibleListing={false}
        onReloadModels={vi.fn()}
        settingsHref="/dev-org/settings?tab=ai-resources"
      />,
    );
    expect(screen.getByRole("button", { name: "检查启用条件" })).toBeEnabled();
  });

  it("keeps the organization loading state distinct from an empty account", () => {
    render(
      <OrganizationStep
        organizations={[]}
        loadingOrganizations
        value=""
        onChange={vi.fn()}
        onContinue={vi.fn()}
        modelResources={[]}
        modelResourceID=""
        onModelChange={vi.fn()}
        loadingModels={false}
        modelError={false}
        incompatibleListing={false}
        onReloadModels={vi.fn()}
        settingsHref=""
      />,
    );

    expect(screen.getByText("正在加载组织")).toBeInTheDocument();
    expect(screen.queryByText("当前账户还没有可用组织，请先创建组织。"))
      .not.toBeInTheDocument();
  });

  it("reports an invalid expert version instead of model configuration", () => {
    render(
      <OrganizationStep
        organizations={[{ id: 9, slug: "dev-org", name: "研发组织" }]}
        loadingOrganizations={false}
        value="9"
        onChange={vi.fn()}
        onContinue={vi.fn()}
        modelResources={[]}
        modelResourceID=""
        onModelChange={vi.fn()}
        loadingModels={false}
        modelError={false}
        incompatibleListing
        onReloadModels={vi.fn()}
        settingsHref="/dev-org/settings?tab=ai-resources"
      />,
    );

    expect(screen.getByText(
      "当前专家版本缺少兼容 Agent，请联系发布者修正后重新上架。",
    )).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "配置兼容模型" }))
      .not.toBeInTheDocument();
  });
});
