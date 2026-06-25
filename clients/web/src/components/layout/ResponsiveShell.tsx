"use client";

import React from "react";
import { usePathname } from "next/navigation";
import { useBreakpoint } from "./useBreakpoint";
import { IDEShell } from "@/components/ide";
import { MobileShell } from "@/components/mobile";
import { hideIdeChrome, hideMobileTabBar as shouldHideMobileTabBar } from "@/lib/ide-chrome";

interface ResponsiveShellProps {
  children: React.ReactNode;
  sidebarContent?: React.ReactNode;
  mobileTitle?: string;
  mobileHeaderActions?: React.ReactNode;
  hideMobileTabBar?: boolean;
}

export function ResponsiveShell({
  children,
  sidebarContent,
  mobileTitle,
  mobileHeaderActions,
  hideMobileTabBar = false,
}: ResponsiveShellProps) {
  const pathname = usePathname();
  const { isMobile } = useBreakpoint();
  const standalone = hideIdeChrome(pathname);

  if (standalone) {
    return (
      <div className="app-shell flex h-screen flex-col bg-background overflow-hidden">
        {children}
      </div>
    );
  }

  if (isMobile) {
    return (
      <MobileShell
        title={mobileTitle}
        headerActions={mobileHeaderActions}
        hideTabBar={hideMobileTabBar || shouldHideMobileTabBar(pathname)}
      >
        {children}
      </MobileShell>
    );
  }

  return (
    <IDEShell sidebarContent={sidebarContent}>
      {children}
    </IDEShell>
  );
}

export default ResponsiveShell;
