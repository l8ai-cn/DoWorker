import type { Metadata } from "next";

import { SiteHeader } from "@/components/site-header";

import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "Do Worker 专家应用市场",
    template: "%s | Do Worker 市场",
  },
  description: "发现经过验证的 AI 专家应用、Skill、系统连接与资源。",
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
            <span>Do Worker 专家应用市场</span>
            <span>在启用前确认权限、额度与运行要求</span>
          </div>
        </footer>
      </body>
    </html>
  );
}
