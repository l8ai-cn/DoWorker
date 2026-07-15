"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { LoopRuntimeSnapshot } from "@/lib/viewModels/loop-program";

interface LoopRuntimeDialogProps {
  error?: string;
  loading: boolean;
  open: boolean;
  running: boolean;
  snapshots: LoopRuntimeSnapshot[];
  onOpenChange: (open: boolean) => void;
  onRetry: () => void;
  onRun: (snapshotId: string) => void;
}

function runtimeLabel(snapshot: LoopRuntimeSnapshot): string {
  const name = snapshot.alias || "未命名环境";
  return `${name} · ${snapshot.workerType} · 快照 ${snapshot.id}`;
}

export function LoopRuntimeDialog({
  error,
  loading,
  open,
  running,
  snapshots,
  onOpenChange,
  onRetry,
  onRun,
}: LoopRuntimeDialogProps) {
  const [selectedId, setSelectedId] = useState("");
  const selected = snapshots.find(({ id }) => id === selectedId);

  function changeOpen(nextOpen: boolean) {
    if (!nextOpen) setSelectedId("");
    onOpenChange(nextOpen);
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen} overlayClassName="z-[100001]">
      <DialogContent
        className="max-w-md"
        title="选择运行环境"
        description="运行环境只在本次启动时绑定，不属于循环编排。"
      >
        <DialogBody className="space-y-3">
          <Label>运行环境</Label>
          <Select
            disabled={loading || snapshots.length === 0 || running}
            value={selectedId}
            onValueChange={setSelectedId}
          >
            <SelectTrigger>
              {selected ? runtimeLabel(selected) : <SelectValue placeholder="选择运行环境" />}
            </SelectTrigger>
            <SelectContent>
              {snapshots.map((snapshot) => (
                <SelectItem key={snapshot.id} value={snapshot.id}>
                  {runtimeLabel(snapshot)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {loading && (
            <p className="text-sm text-muted-foreground">正在加载运行环境</p>
          )}
          {!loading && error && (
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm text-destructive">{error}</p>
              <Button onClick={onRetry} variant="outline">
                重新加载
              </Button>
            </div>
          )}
          {!loading && !error && snapshots.length === 0 && (
            <p className="text-sm text-destructive">当前组织没有可用的运行环境</p>
          )}
        </DialogBody>
        <DialogFooter>
          <Button disabled={running} onClick={() => changeOpen(false)} variant="outline">
            取消
          </Button>
          <Button
            disabled={loading || !selected || running}
            loading={running}
            onClick={() => {
              if (!selected) return;
              setSelectedId("");
              onRun(selected.id);
            }}
          >
            启动循环
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
