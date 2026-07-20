import { fireEvent, render, screen, waitFor } from "@/test/test-utils";
import { describe, expect, it, vi } from "vitest";
import zhMessages from "@/messages/zh/app.json";
import { LoopCustomBlockDialog } from "../loop-custom-block-dialog";
import {
  createLoopWorkbenchMessages,
  type LoopMessageTranslator,
} from "../loop-workbench-messages";

function translator(messages: Record<string, unknown>): LoopMessageTranslator {
  return (key) => String(key.split(".").reduce<unknown>(
    (current, segment) => (current as Record<string, unknown>)[segment],
    messages,
  ));
}

const messages = createLoopWorkbenchMessages(
  translator(zhMessages.loopWorkbench),
).customBlock;

describe("LoopCustomBlockDialog", () => {
  it("creates a versioned custom block definition from editable templates", async () => {
    const onCreate = vi.fn();
    render(
      <LoopCustomBlockDialog
        definitions={[]}
        messages={messages}
        open
        onCreate={onCreate}
        onOpenChange={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText("积木名称"), {
      target: { value: "专业 PPT" },
    });
    fireEvent.change(screen.getByLabelText("积木标识"), {
      target: { value: "ppt-step" },
    });
    fireEvent.change(screen.getByLabelText("任务模板"), {
      target: { value: "制作 {{topic}} 的专业 PPT" },
    });
    fireEvent.change(screen.getByLabelText("验证命令模板"), {
      target: { value: "test -f {{file}}" },
    });
    fireEvent.change(screen.getByLabelText("验收说明模板"), {
      target: { value: "{{file}} 存在且可打开" },
    });
    fireEvent.click(screen.getByRole("button", { name: "创建积木" }));

    await waitFor(() => expect(onCreate).toHaveBeenCalledWith(expect.objectContaining({
      label: "专业 PPT",
      parameters: ["topic", "file"],
      slug: "ppt-step",
      version: 1,
    })));
  });

  it("rejects invalid identifiers before creating the block", () => {
    const onCreate = vi.fn();
    render(
      <LoopCustomBlockDialog
        definitions={[]}
        messages={messages}
        open
        onCreate={onCreate}
        onOpenChange={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText("积木名称"), {
      target: { value: "专业 PPT" },
    });
    fireEvent.change(screen.getByLabelText("积木标识"), {
      target: { value: "PPT_step" },
    });
    fireEvent.click(screen.getByRole("button", { name: "创建积木" }));

    expect(screen.getByText("使用 2-100 位小写字母、数字或连字符")).toBeInTheDocument();
    expect(onCreate).not.toHaveBeenCalled();
  });

  it("creates the next immutable version for an existing block slug", async () => {
    const onCreate = vi.fn();
    render(
      <LoopCustomBlockDialog
        definitions={[{
          slug: "ppt-step",
          version: 1,
          label: "旧版 PPT",
          parameters: [],
          expansion: {
            agentLocalId: "ppt-step-task",
            verifierLocalId: "ppt-step-check",
            promptTemplate: "旧模板",
            commandTemplate: "旧命令",
            acceptTemplate: "旧验收",
          },
        }]}
        messages={messages}
        open
        onCreate={onCreate}
        onOpenChange={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText("积木名称"), {
      target: { value: "新版 PPT" },
    });
    fireEvent.click(screen.getByRole("button", { name: "创建积木" }));

    await waitFor(() => expect(onCreate).toHaveBeenCalledWith(expect.objectContaining({
      label: "新版 PPT",
      slug: "ppt-step",
      version: 2,
    })));
  });
});
