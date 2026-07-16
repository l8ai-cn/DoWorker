import type { Metadata, Viewport } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import { ThemeProvider, ThemeColorMeta } from "@/components/theme";
import { PWAProvider } from "@/components/pwa";
import { PostHogProvider } from "@/providers/PostHogProvider";
import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";
import { Toaster } from "sonner";
import "@fontsource-variable/space-grotesk/wght.css";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL("https://agentsmesh.ai"),
  title: {
    default: "Do Worker - The AI Agent Workforce Platform",
    template: "%s | Do Worker",
  },
  description: "Ship like a team of fifty. With a team of five. Give every team member an AI agent squad — assign tasks, track progress, and let them collaborate autonomously.",
  keywords: [
    "do-worker", "do worker", "l8ai",
    "AI agent workforce platform", "agent team management", "AI agent team",
    "AI agents", "AI coding", "Claude Code", "Codex CLI", "Gemini CLI", "Aider",
    "multi-agent collaboration", "agent coordination", "terminal AI", "code automation",
    "developer tools", "enterprise development", "self-hosted", "agent fleet",
    "AI developer tools", "coding agents", "agent management",
    "multi-agent orchestration", "team productivity",
  ],
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "default",
    title: "Do Worker",
  },
  formatDetection: {
    telephone: false,
  },
  openGraph: {
    type: "website",
    siteName: "Do Worker",
    title: "Do Worker - The AI Agent Workforce Platform",
    description: "Ship like a team of fifty. With a team of five. Give every team member an AI agent squad — assign tasks, track progress, and let them collaborate autonomously.",
    url: "https://agentsmesh.ai",
  },
  twitter: {
    card: "summary_large_image",
    title: "Do Worker - The AI Agent Workforce Platform",
    description: "Ship like a team of fifty. With a team of five. Give every team member an AI agent squad — assign tasks, track progress, and let them collaborate autonomously.",
  },
  alternates: {
    canonical: "https://agentsmesh.ai",
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#ffffff" },
    { media: "(prefers-color-scheme: dark)", color: "#16130f" },
  ],
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = await getLocale();
  const messages = await getMessages();

  return (
    <html lang={locale} suppressHydrationWarning>
      <body
        className={`${GeistSans.variable} ${GeistMono.variable} antialiased bg-background text-foreground`}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
          themes={["light", "dark", "solarized-light", "solarized-dark"]}
        >
          <PostHogProvider>
            <NextIntlClientProvider locale={locale} messages={messages}>
              <PWAProvider>
                {children}
              </PWAProvider>
            </NextIntlClientProvider>
          </PostHogProvider>
          <ThemeColorMeta />
          <Toaster richColors position="top-right" />
        </ThemeProvider>
      </body>
    </html>
  );
}
