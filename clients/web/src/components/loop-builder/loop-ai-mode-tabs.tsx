"use client";

import { FileText, Sparkles } from "lucide-react";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { LoopAIMode } from "./loop-ai-assistant-types";
import type { LoopAIMessages } from "./loop-workbench-messages";

interface LoopAIModeTabsProps {
  mode: LoopAIMode;
  busy: boolean;
  messages: LoopAIMessages;
  onModeChange: (mode: LoopAIMode) => void;
}

export function LoopAIModeTabs({
  mode,
  busy,
  messages,
  onModeChange,
}: LoopAIModeTabsProps) {
  return (
    <Tabs value={mode} onValueChange={(value) => onModeChange(value as LoopAIMode)}>
      <TabsList className="grid w-full grid-cols-2">
        <TabsTrigger disabled={busy} value="generate">
          <Sparkles className="mr-1.5 h-4 w-4" />
          {messages.generateMode}
        </TabsTrigger>
        <TabsTrigger disabled={busy} value="explain">
          <FileText className="mr-1.5 h-4 w-4" />
          {messages.explainMode}
        </TabsTrigger>
      </TabsList>
    </Tabs>
  );
}
