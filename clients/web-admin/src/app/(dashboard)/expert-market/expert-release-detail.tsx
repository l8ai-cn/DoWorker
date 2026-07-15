"use client";

import { Check, X } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import type { ExpertMarketRelease } from "@/lib/api/admin";
import { ExpertReleaseStatus } from "./expert-release-status";
import { ReleaseJsonSnapshot } from "./release-json-snapshot";
import { SkillDependencies } from "./skill-dependencies";

interface ExpertReleaseDetailProps {
  release: ExpertMarketRelease;
  isActing: boolean;
  rejectionReason: string;
  rejectionError: string;
  isRejecting: boolean;
  onApprove: () => void;
  onRejectStart: () => void;
  onRejectCancel: () => void;
  onRejectConfirm: () => void;
  onReasonChange: (value: string) => void;
}

export function ExpertReleaseDetail(props: ExpertReleaseDetailProps) {
  const { release } = props;
  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-xl font-semibold">{release.name}</h2>
            <ExpertReleaseStatus status={release.status} />
            {release.featured && <Badge variant="outline">精选</Badge>}
          </div>
          <p className="mt-1 text-sm text-muted-foreground">{release.summary}</p>
        </div>
        {release.status === "pending" && !props.isRejecting && (
          <div className="flex shrink-0 gap-2">
            <Button
              variant="outline"
              disabled={props.isActing}
              onClick={props.onRejectStart}
            >
              <X />
              驳回
            </Button>
            <Button disabled={props.isActing} onClick={props.onApprove}>
              <Check />
              批准发布
            </Button>
          </div>
        )}
      </div>

      <dl className="grid gap-3 rounded-md border border-border p-4 text-sm sm:grid-cols-3">
        <Meta label="发布 ID" value={`#${release.id}`} />
        <Meta label="申请 ID" value={`#${release.application_id}`} />
        <Meta label="来源专家 ID" value={`#${release.source_expert_id}`} />
        <Meta label="发布组织 ID" value={`#${release.publisher_organization_id}`} />
        <Meta label="发布用户 ID" value={`#${release.publisher_user_id}`} />
        <Meta label="版本" value={`v${release.version}`} />
      </dl>

      <section>
        <h3 className="mb-2 text-sm font-semibold">发布描述</h3>
        <p className="whitespace-pre-wrap text-sm text-muted-foreground">
          {release.description || "未提供描述"}
        </p>
      </section>

      <SkillDependencies value={release.skill_dependencies_json} />
      <ReleaseJsonSnapshot title="专家快照" value={release.expert_snapshot_json} />
      <ReleaseJsonSnapshot
        title="Worker Spec 快照"
        value={release.worker_spec_snapshot_json}
      />

      {release.rejection_reason && (
        <section className="rounded-md border border-destructive/40 bg-destructive/5 p-4">
          <h3 className="text-sm font-semibold text-destructive">驳回理由</h3>
          <p className="mt-1 text-sm">{release.rejection_reason}</p>
        </section>
      )}

      {props.isRejecting && (
        <section className="rounded-md border border-destructive/40 p-4">
          <label htmlFor="rejection-reason" className="text-sm font-medium">
            驳回理由
          </label>
          <Textarea
            id="rejection-reason"
            className="mt-2 min-h-24"
            value={props.rejectionReason}
            error={props.rejectionError}
            onChange={(event) => props.onReasonChange(event.target.value)}
          />
          <div className="mt-3 flex justify-end gap-2">
            <Button
              variant="ghost"
              disabled={props.isActing}
              onClick={props.onRejectCancel}
            >
              取消
            </Button>
            <Button
              variant="destructive"
              disabled={props.isActing}
              onClick={props.onRejectConfirm}
            >
              确认驳回
            </Button>
          </div>
        </section>
      )}
    </div>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-1 font-medium">{value}</dd>
    </div>
  );
}
