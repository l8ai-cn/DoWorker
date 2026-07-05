"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  destroyAfterOptions,
  destroyPolicyOptions,
  type DestroyPolicy,
} from "./podLifecycleOptions";

interface PodLifecycleSectionProps {
  destroyPolicy: DestroyPolicy;
  destroyAfterMinutes: number;
  onPolicyChange: (policy: DestroyPolicy) => void;
  onAfterChange: (minutes: number) => void;
}

export function PodLifecycleSection({
  destroyPolicy,
  destroyAfterMinutes,
  onPolicyChange,
  onAfterChange,
}: PodLifecycleSectionProps) {
  const selected = destroyPolicyOptions.find((o) => o.value === destroyPolicy);

  return (
    <section className="rounded-lg border border-border bg-surface-muted/35 p-3">
      <div className="mb-3">
        <h3 className="text-sm font-medium text-foreground">生命周期策略</h3>
        <p className="text-xs leading-5 text-muted-foreground">
          配置实例启动后的保留方式，避免临时任务长期占用 Runner 资源。
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            销毁策略
          </label>
          <Select
            value={destroyPolicy}
            onValueChange={(value) => onPolicyChange(value as DestroyPolicy)}
          >
            <SelectTrigger>
              <SelectValue placeholder="选择销毁策略" />
            </SelectTrigger>
            <SelectContent>
              {destroyPolicyOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="mt-1 text-xs text-muted-foreground">{selected?.description}</p>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            销毁时间
          </label>
          <Select
            value={String(destroyAfterMinutes)}
            onValueChange={(value) => onAfterChange(Number(value))}
            disabled={destroyPolicy === "manual"}
          >
            <SelectTrigger>
              <SelectValue placeholder="选择时间" />
            </SelectTrigger>
            <SelectContent>
              {destroyAfterOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <p className="mt-1 text-xs text-muted-foreground">
            手动销毁模式下不会设置自动销毁时间。
          </p>
        </div>
      </div>
    </section>
  );
}
