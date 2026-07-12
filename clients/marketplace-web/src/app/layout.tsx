import type { Metadata } from "next";

import { SiteHeader } from "@/components/site-header";

import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "Do Worker 应用市场",
    template: "%s | Do Worker 市场",
  },
  description: "发现可直接启用的专家应用，确认条件后在 Do Worker 中开始任务。",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="zh-CN">
      <body>
        <SiteHeader />
        {children}
        <footer className="site-footer">
          <div className="shell">
            <span>Do Worker 应用市场</span>
            <span>发现、评估、启用，然后开始真实工作</span>
          </div>
        </footer>
      </body>
    </html>
  );
}
