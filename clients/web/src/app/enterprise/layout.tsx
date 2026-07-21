import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Enterprise",
  description: "Your agents. Your infrastructure. Your rules. Enterprise-grade security and governance for organizations managing agent workforces at scale.",
  alternates: {
    canonical: "https://agentcloud.ai/enterprise",
  },
  openGraph: {
    title: "Enterprise | Agent Cloud",
    description: "Your agents. Your infrastructure. Your rules. Enterprise-grade security and governance for organizations managing agent workforces at scale.",
    url: "https://agentcloud.ai/enterprise",
  },
};

export default function EnterpriseLayout({ children }: { children: React.ReactNode }) {
  return children;
}
