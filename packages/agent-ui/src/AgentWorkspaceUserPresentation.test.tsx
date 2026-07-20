import { fireEvent, render, screen } from "@testing-library/react";

import type { AgentArtifactItem } from "./agentArtifactContracts";
import { AgentWorkspace } from "./AgentWorkspace";
import {
  agentWorkspaceRuntime as runtime,
  agentWorkspaceSnapshot as sessionSnapshot,
} from "./AgentWorkspace.test-fixture";

describe("AgentWorkspace user presentation", () => {
  it("shows verified video execution steps without raw runtime details", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "completed";
    snapshot.plan = [];
    snapshot.permissions = [];
    snapshot.items = [
      {
        id: "user-1",
        kind: "message",
        role: "user",
        text: "生成一个灯笼升空的视频",
        status: "completed",
      },
      {
        id: "system-1",
        kind: "system",
        title: "System",
        detail: "[stderr] tools registered: Bash, Read, WebSearch",
        status: "completed",
      },
      {
        id: "tool-1",
        identity: {
          namespace: "agentsmesh.acp",
          schemaVersion: "1",
          semanticKey: "shell",
        },
        kind: "tool",
        results: [],
        title: "shell",
        output: "internal protocol output",
        status: "completed",
      },
      verifiedVideoArtifact(),
    ];
    const { agentRuntime, terminalRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
        terminalRuntime={terminalRuntime}
      />,
    );

    expect(await screen.findByText("生成一个灯笼升空的视频")).toBeVisible();
    expect(screen.getByLabelText("视频创作进度")).toBeVisible();
    expect(screen.getByText("接收创作请求")).toBeVisible();
    expect(screen.getByText("生成视频画面")).toBeVisible();
    expect(screen.getByText("准备播放文件")).toBeVisible();
    expect(screen.getByText("校验并发布结果")).toBeVisible();
    expect(screen.getByText("文件已校验并发布")).toBeVisible();
    expect(screen.getByText("视频文件已发布并通过完整性校验")).toBeVisible();
    expect(screen.getByRole("tab", { name: "成果" })).toBeVisible();
    expect(screen.queryByText(/tools registered/)).not.toBeInTheDocument();
    expect(screen.queryByText("internal protocol output")).not.toBeInTheDocument();
    expect(screen.queryByRole("tab", { name: "终端" })).not.toBeInTheDocument();
    expect(screen.queryByText("智能体模式")).not.toBeInTheDocument();
    expect(screen.queryByText("dev-runner-codex")).not.toBeInTheDocument();
    expect(screen.queryByText("生成过程")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("tab", { name: "成果" }));
    expect(screen.queryByText("video/mp4")).not.toBeInTheDocument();
    expect(screen.queryByText("playable")).not.toBeInTheDocument();
  });

  it("does not expose raw execution errors", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "failed";
    snapshot.error = "server error 503: credential path=/internal/secret";
    snapshot.items = [];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(
      await screen.findByText("任务执行失败"),
    ).toBeVisible();
    expect(screen.queryByText(/credential path/)).not.toBeInTheDocument();
  });

  it.each([
    {
      detail: "第三方视频账号池当前没有可用账号，未生成视频文件。",
      error:
        "[CREATIVE_NO_ACCOUNT_AVAILABLE] pool exhausted account_id=internal-7",
      hidden: /CREATIVE_NO_ACCOUNT_AVAILABLE|account_id/,
      title: "视频服务暂不可用",
      trace: "视频服务当前没有可用账号",
    },
    {
      detail: "供应商拒绝了当前视频生成凭据，未生成视频文件。",
      error: "[INVALID_API_KEY] provider rejected key path=/internal/secret",
      hidden: /INVALID_API_KEY|internal\/secret/,
      title: "视频服务凭据无效",
      trace: "视频服务凭据校验失败",
    },
    {
      detail: "未取得可验证视频文件。",
      error: "[API_KEY_QUOTA_EXHAUSTED] completion review failed token=secret",
      hidden: /API_KEY_QUOTA_EXHAUSTED|token=secret/,
      title: "智能体主模型额度耗尽",
      trace: "智能体主模型额度不足；未取得可验证视频文件",
    },
  ])(
    "projects task failures without raw technical details: $title",
    async ({ detail, error, hidden, title, trace }) => {
      const snapshot = sessionSnapshot();
      snapshot.status = "failed";
      snapshot.error = error;
      snapshot.latestUserCommandId = "command-current";
      snapshot.items = [
        {
          id: "user-1",
          kind: "message",
          role: "user",
          text: "生成一个湖面灯笼视频",
          status: "completed",
        },
      ];
      snapshot.plan = [];
      snapshot.permissions = [];
      const { agentRuntime } = runtime(snapshot);

      render(
        <AgentWorkspace
          locale="zh-CN"
          presentation="user"
          runtime={agentRuntime}
          sessionId={snapshot.sessionId}
        />,
      );

      expect(await screen.findByText(title)).toBeVisible();
      expect(screen.getByText(detail)).toBeVisible();
      expect(screen.getByLabelText("视频创作进度")).toBeVisible();
      expect(screen.getByText("接收创作请求")).toBeVisible();
      expect(screen.getByText("生成视频画面")).toBeVisible();
      expect(screen.getByText(trace)).toBeVisible();
      expect(screen.queryByText(hidden)).not.toBeInTheDocument();
      expect(screen.queryByRole("tab", { name: "成果" })).not.toBeInTheDocument();
      expect(
        screen.queryByText("视频文件已发布并通过完整性校验"),
      ).not.toBeInTheDocument();
    },
  );

  it("labels a verified video from a failed session as partial", async () => {
    const snapshot = sessionSnapshot();
    snapshot.status = "failed";
    snapshot.error = "post-processing command failed";
    snapshot.items = [verifiedVideoArtifact()];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(
      await screen.findByText("视频文件可用，但任务未完整结束"),
    ).toBeVisible();
    expect(screen.getByRole("tab", { name: "成果" })).toBeVisible();
    expect(
      screen.queryByText("视频文件已发布并通过完整性校验"),
    ).not.toBeInTheDocument();
    expect(screen.queryByText(/post-processing/)).not.toBeInTheDocument();
  });

  it("hides a completed video that fails file verification", async () => {
    const snapshot = sessionSnapshot();
    const invalidVideo = verifiedVideoArtifact();
    invalidVideo.representations[0]!.byteSize = BigInt(0);
    snapshot.status = "completed";
    snapshot.items = [invalidVideo];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(await screen.findByText("视频成果校验未通过")).toBeVisible();
    expect(screen.queryByRole("tab", { name: "成果" })).not.toBeInTheDocument();
    expect(
      screen.queryByText("视频文件已发布并通过完整性校验"),
    ).not.toBeInTheDocument();
  });

  it("keeps a verified result when an older video revision is invalid", async () => {
    const snapshot = sessionSnapshot();
    const invalidVideo = verifiedVideoArtifact();
    invalidVideo.id = "artifact-old";
    invalidVideo.artifactId = "seedance-video-old";
    invalidVideo.representations[0]!.digest = undefined;
    snapshot.status = "completed";
    snapshot.items = [invalidVideo, verifiedVideoArtifact()];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(
      await screen.findByText("视频文件已发布并通过完整性校验"),
    ).toBeVisible();
    expect(screen.getByRole("tab", { name: "成果" })).toBeVisible();
    expect(screen.queryByText("视频成果校验未通过")).not.toBeInTheDocument();
  });

  it("does not claim the current task succeeded from another command's video", async () => {
    const snapshot = sessionSnapshot();
    const previousVideo = verifiedVideoArtifact();
    previousVideo.provenance = {
      ...previousVideo.provenance,
      commandId: "command-previous",
    };
    snapshot.latestUserCommandId = "command-current";
    snapshot.status = "completed";
    snapshot.items = [previousVideo];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(await screen.findByText("视频成果校验未通过")).toBeVisible();
    expect(
      screen.queryByText("视频文件已发布并通过完整性校验"),
    ).not.toBeInTheDocument();
  });

  it("does not claim the current task succeeded from an unbound historical video", async () => {
    const snapshot = sessionSnapshot();
    snapshot.latestUserCommandId = "command-current";
    snapshot.status = "completed";
    snapshot.items = [verifiedVideoArtifact()];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(await screen.findByText("视频成果校验未通过")).toBeVisible();
    expect(
      screen.queryByText("视频文件已发布并通过完整性校验"),
    ).not.toBeInTheDocument();
  });

  it("accepts a verified video from the latest user command", async () => {
    const snapshot = sessionSnapshot();
    const currentVideo = verifiedVideoArtifact();
    currentVideo.provenance = {
      ...currentVideo.provenance,
      commandId: "command-current",
    };
    snapshot.latestUserCommandId = "command-current";
    snapshot.status = "completed";
    snapshot.items = [currentVideo];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(
      await screen.findByText("视频文件已发布并通过完整性校验"),
    ).toBeVisible();
  });

  it("does not leave a completed session labeled as still generating", async () => {
    const snapshot = sessionSnapshot();
    const processingVideo = verifiedVideoArtifact();
    processingVideo.status = "processing";
    if (processingVideo.manifest?.kind === "video") {
      processingVideo.manifest.stage = "rendering";
    }
    snapshot.status = "completed";
    snapshot.items = [processingVideo];
    snapshot.plan = [];
    snapshot.permissions = [];
    const { agentRuntime } = runtime(snapshot);

    render(
      <AgentWorkspace
        locale="zh-CN"
        presentation="user"
        runtime={agentRuntime}
        sessionId={snapshot.sessionId}
      />,
    );

    expect(await screen.findByText("视频成果校验未通过")).toBeVisible();
    expect(screen.queryByText("正在生成视频")).not.toBeInTheDocument();
  });
});

function verifiedVideoArtifact(): AgentArtifactItem {
  return {
    actions: [],
    artifactId: "seedance-video",
    filename: "seedance-video.mp4",
    grants: [],
    id: "artifact-video",
    kind: "artifact",
    manifest: {
      derivativeRepresentationIds: [],
      kind: "video",
      playableRepresentationId: "playable",
      stage: "ready",
      thumbnailRepresentationIds: [],
    },
    mimeType: "video/mp4",
    provenance: {
      publicationToolExecutionId: "workbench-publish:1",
      producerId: "seedance-task-1",
      producerNamespace: "seedance",
      producerType: "video.generate",
    },
    representations: [
      {
        byteSize: 1_700_000n,
        digest: `sha256:${"b".repeat(64)}`,
        mediaType: "video/mp4",
        representationId: "playable",
        revision: 1n,
        role: "playable",
        status: "ready",
      },
    ],
    revision: 1n,
    role: "video",
    schemaVersion: "1",
    selectedRepresentationId: "playable",
    status: "completed",
  };
}
