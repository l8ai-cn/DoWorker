import { create, toBinary } from "@bufbuild/protobuf";
import {
  PlanResourceResponseSchema,
} from "../../../../../proto/gen/ts/orchestration_resource/v1/orchestration_resource_queries_pb";
import {
  IssueSeverity,
  PlanStatus,
  ResourceOperation,
} from "../../../../../proto/gen/ts/orchestration_resource/v1/orchestration_resource_types_pb";
import { expect, test } from "../../fixtures/index";
import { TEST_ORG_SLUG } from "../../helpers/env";

const promptName = "e2e-resource-prompt";
const planEndpoint =
  "**/proto.orchestration_resource.v1.OrchestrationResourceService/PlanResource";
const applyPromptEndpoint =
  "**/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyPromptPlan";

test.describe("Resource-native orchestration editor", () => {
  test("round-trips one Draft and applies the reviewed Prompt plan", async ({
    page,
  }) => {
    const marker = Date.now().toString();
    const formContent = `e2e form prompt ${marker}`;
    const yamlContent = `e2e yaml prompt ${marker}`;

    await openPromptEditor(page);
    const editor = page.getByTestId("resource-editor");

    await editor.getByLabel(/^(Resource name|资源名称)\s*\*$/i)
      .fill(promptName);
    await editor.getByLabel(/^(Display name|显示名称)$/i)
      .fill("E2E resource prompt");
    await editor.getByLabel(/^(Prompt content|Prompt 内容)\s*\*$/i)
      .fill(formContent);

    await editor.getByRole("tab", { name: "YAML" }).click();
    const yamlEditor = editor.getByTestId("resource-yaml-editor");
    const generatedYaml = await yamlEditor.inputValue();
    expect(generatedYaml).toContain(`name: ${promptName}`);
    expect(generatedYaml).toContain(formContent);

    await yamlEditor.fill(generatedYaml.replace(formContent, yamlContent));
    await editor.getByRole("tab", {
      name: /^(Plan & diff|计划与差异)$/i,
    }).click();
    await expect(editor.getByRole("tab", {
      name: /^(Plan & diff|计划与差异)$/i,
    })).toHaveAttribute("aria-selected", "true");

    await editor.getByRole("button", {
      name: /^(Validate|校验)$/i,
    }).click();
    await expect(editor.getByText(
      /The resource is valid|资源校验通过/i,
    )).toBeVisible();

    await editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    }).click();
    await expect(editor.getByText("Plan ID")).toBeVisible();

    await editor.getByRole("tab", {
      name: /^(Configuration|配置)$/i,
    }).click();
    await expect(editor.getByLabel(/^(Prompt content|Prompt 内容)\s*\*$/i))
      .toHaveValue(yamlContent);
    await editor.getByRole("tab", {
      name: /^(Plan & diff|计划与差异)$/i,
    }).click();
    await expect(editor.getByRole("tab", {
      name: /^(Plan & diff|计划与差异)$/i,
    })).toHaveAttribute("aria-selected", "true");
    await expect(editor.getByText(
      /Semantic changes|语义变更/i,
    )).toBeVisible();

    await editor.getByRole("button", {
      name: /^(Apply resource|应用资源)$/i,
    }).click();
    await expect(editor.getByText(/^(Applied|已应用)$/i)).toBeVisible();
    await expect(editor.getByRole("alert")).toContainText(
      /Revision \d+|版本 \d+/i,
    );
    await expect(editor.getByRole("button", {
      name: /^(Apply resource|应用资源)$/i,
    })).toBeDisabled();
  });

  test("keeps invalid YAML local and blocks submission on mobile", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await openPromptEditor(page);
    const editor = page.getByTestId("resource-editor");
    const mobileActions = [
      editor.getByRole("button", { name: /^(Validate|校验)$/i }),
      editor.getByRole("button", {
        name: /^(Generate plan|生成计划)$/i,
      }),
      editor.getByRole("button", {
        name: /^(Apply resource|应用资源)$/i,
      }),
    ];
    for (const action of mobileActions) {
      const box = await action.boundingBox();
      expect(box?.height).toBeGreaterThanOrEqual(44);
    }
    let validationRequests = 0;
    page.on("request", (request) => {
      if (/OrchestrationResourceService\/ValidateResource$/.test(
        new URL(request.url()).pathname,
      )) {
        validationRequests += 1;
      }
    });

    await editor.getByRole("tab", { name: "YAML" }).click();
    await editor.getByTestId("resource-yaml-editor").fill([
      "apiVersion: agentcloud.io/v1alpha1",
      "kind: Prompt",
      "metadata:",
      `  name: ${promptName}`,
      "  name: duplicate-name",
      "spec:",
      "  content: blocked",
    ].join("\n"));
    await editor.getByRole("tab", {
      name: /^(Configuration|配置)$/i,
    }).click();

    await expect(editor.getByText(
      /Fix YAML before returning|请先修复 YAML/i,
    )).toBeVisible();
    await expect(editor.getByRole("tab", { name: "YAML" }))
      .toHaveAttribute("aria-selected", "true");
    await expect(editor.getByRole("button", {
      name: /^(Validate|校验)$/i,
    })).toBeDisabled();
    await expect(editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    })).toBeDisabled();
    expect(validationRequests).toBe(0);
  });

  test("shows permission denial without exposing the control-plane URL", async ({
    page,
  }) => {
    await page.route(planEndpoint, (route) => route.fulfill({
      status: 403,
      contentType: "application/json",
      body: JSON.stringify({
        code: "permission_denied",
        message: "You cannot plan this resource.",
      }),
    }), { times: 1 });
    await openPromptEditor(page);
    const editor = page.getByTestId("resource-editor");
    await fillPromptDraft(editor, `permission-${Date.now()}`);

    await editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    }).click();

    await expect(editor.getByRole("alert")).toContainText(
      "You cannot plan this resource.",
    );
    await expect(editor).not.toContainText(
      "OrchestrationResourceService/PlanResource",
    );
  });

  test("keeps blocking plan issues visible and Apply disabled", async ({
    page,
  }) => {
    await page.route(planEndpoint, (route) => fulfillPlan(route, {
      issues: [{
        severity: IssueSeverity.BLOCKING,
        code: "REFERENCE_FORBIDDEN",
        path: "/spec/modelRef",
        message: "The model binding is not readable.",
      }],
    }), { times: 1 });
    await openPromptEditor(page);
    const editor = page.getByTestId("resource-editor");
    await fillPromptDraft(editor, `blocked-${Date.now()}`);

    await editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    }).click();

    await expect(editor).toContainText(
      "/spec/modelRef: The model binding is not readable.",
    );
    await expect(editor.getByRole("button", {
      name: /^(Apply resource|应用资源)$/i,
    })).toBeDisabled();
  });

  test("expires a reviewed plan and recovers by generating a new one", async ({
    page,
  }) => {
    const resourceName = `expiry-${Date.now()}`;
    await page.route(planEndpoint, (route) => fulfillPlan(route, {
      planId: "44444444-4444-4444-8444-444444444444",
      expiresAt: new Date(Date.now() + 2_000).toISOString(),
      resourceName,
    }), { times: 1 });
    await openPromptEditor(page);
    const editor = page.getByTestId("resource-editor");
    await fillPromptDraft(editor, resourceName);
    const apply = editor.getByRole("button", {
      name: /^(Apply resource|应用资源)$/i,
    });

    await editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    }).click();
    await expect(apply).toBeEnabled();
    await expect(editor.getByText(
      /This plan has expired|该计划已过期/i,
    )).toHaveCount(1);
    await expect(editor.getByText(/^(Plan expired|计划已过期)$/i))
      .toBeVisible();
    await expect(editor.getByText(
      "44444444-4444-4444-8444-444444444444",
    )).toBeVisible();
    await expect(apply).toBeDisabled();

    await editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    }).click();
    await expect(apply).toBeEnabled();
  });

  test("recovers an Apply conflict only through re-plan", async ({ page }) => {
    await openPromptEditor(page);
    const editor = page.getByTestId("resource-editor");
    await fillPromptDraft(editor, `conflict-${Date.now()}`);
    await editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    }).click();
    await expect(editor.getByText("Plan ID")).toBeVisible();
    await page.route(applyPromptEndpoint, (route) => route.fulfill({
      status: 409,
      contentType: "application/json",
      body: JSON.stringify({
        code: "plan_conflict",
        message: "The resource changed after this plan was generated.",
      }),
    }), { times: 1 });
    const apply = editor.getByRole("button", {
      name: /^(Apply resource|应用资源)$/i,
    });

    await apply.click();

    await expect(editor.getByRole("alert")).toContainText(
      "The resource changed after this plan was generated.",
    );
    await expect(apply).toBeDisabled();

    await editor.getByRole("button", {
      name: /^(Generate plan|生成计划)$/i,
    }).click();
    await expect(apply).toBeEnabled();
  });
});

async function openPromptEditor(page: import("@playwright/test").Page) {
  await page.goto(`/${TEST_ORG_SLUG}/workers/new?nosw=1`);
  await page.getByTestId("pill-tab-resources").click();
  await expect(page.getByTestId("resource-editor")).toContainText(
    /Prompt/i,
  );
}

async function fillPromptDraft(
  editor: import("@playwright/test").Locator,
  name: string,
) {
  await editor.getByLabel(/^(Resource name|资源名称)\s*\*$/i).fill(name);
  await editor.getByLabel(/^(Prompt content|Prompt 内容)\s*\*$/i)
    .fill(`Prompt content for ${name}`);
}

async function fulfillPlan(
  route: import("@playwright/test").Route,
  input: {
    planId?: string;
    expiresAt?: string;
    resourceName?: string;
    issues?: Array<{
      severity: IssueSeverity;
      code: string;
      path: string;
      message: string;
    }>;
  },
) {
  const response = create(PlanResourceResponseSchema, {
    operation: ResourceOperation.CREATE,
    ...(input.planId ? {
      canonicalJson: new TextEncoder().encode(JSON.stringify({
        apiVersion: "agentcloud.io/v1alpha1",
        kind: "Prompt",
        metadata: {
          name: input.resourceName ?? promptName,
          namespace: TEST_ORG_SLUG,
          displayName: "E2E resource prompt",
          labels: {},
        },
        spec: {
          content: `Prompt content for ${input.resourceName ?? promptName}`,
          variables: {},
        },
      })),
      plan: {
        planId: input.planId,
        operation: ResourceOperation.CREATE,
        expiresAt: input.expiresAt ?? "2099-07-16T00:00:00Z",
        status: PlanStatus.PENDING,
      },
    } : {}),
    issues: input.issues ?? [],
  });
  await route.fulfill({
    status: 200,
    headers: { "content-type": "application/proto" },
    body: Buffer.from(toBinary(PlanResourceResponseSchema, response)),
  });
}
