import React from "react";
import { BlocksSidebar } from "@/components/blocks/BlocksSidebar";
import { ExpertsSidebarContent } from "@/components/experts/ExpertsSidebarContent";
import type { ActivityType } from "@/stores/ide";
import { ChannelsSidebarContent } from "./sidebar/ChannelsSidebarContent";
import { InfraSidebarContent } from "./sidebar/InfraSidebarContent";
import { MeshSidebarContent } from "./sidebar/MeshSidebarContent";
import { MarketplaceSidebarContent } from "./sidebar/MarketplaceSidebarContent";
import { RepositoriesSidebarContent } from "./sidebar/RepositoriesSidebarContent";
import { RunnersSidebarContent } from "./sidebar/RunnersSidebarContent";
import { SettingsSidebarContent } from "./sidebar/SettingsSidebarContent";
import { SkillsSidebarContent } from "./sidebar/SkillsSidebarContent";
import { TicketsSidebarContent } from "./sidebar/TicketsSidebarContent";
import { WorkflowsSidebarContent } from "./sidebar/WorkflowsSidebarContent";
import { WorkspaceSidebarContent } from "./sidebar/WorkspaceSidebarContent";

export interface SidebarCallbacks {
  onCreatePod?: () => void;
  onAddRunner?: () => void;
  onImportRepo?: () => void;
}

export function getSidebarContent(
  activity: ActivityType,
  callbacks: SidebarCallbacks,
): React.ReactNode {
  switch (activity) {
    case "workspace":
      return <WorkspaceSidebarContent onCreatePod={callbacks.onCreatePod} />;
    case "tickets":
      return <TicketsSidebarContent />;
    case "channels":
      return <ChannelsSidebarContent />;
    case "mesh":
      return <MeshSidebarContent />;
    case "workflows":
      return <WorkflowsSidebarContent />;
    case "experts":
      return <ExpertsSidebarContent />;
    case "blocks":
      return <BlocksSidebar />;
    case "infra":
      return (
        <InfraSidebarContent
          onImportRepo={callbacks.onImportRepo}
          onAddRunner={callbacks.onAddRunner}
        />
      );
    case "repositories":
      return <RepositoriesSidebarContent onImportRepo={callbacks.onImportRepo} />;
    case "runners":
      return <RunnersSidebarContent onAddRunner={callbacks.onAddRunner} />;
    case "skills":
      return <SkillsSidebarContent />;
    case "marketplace":
      return <MarketplaceSidebarContent />;
    case "settings":
      return <SettingsSidebarContent />;
    default:
      return null;
  }
}
