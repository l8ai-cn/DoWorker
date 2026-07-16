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
    default: "Do Worker - Enterprise Agent Supply and OPC Incubation",
    template: "%s | Do Worker",
  },
  description: "Build, govern, distribute, and operate reusable Agents for enterprise teams, OPC founders, and higher-education digital employee programs.",
  keywords: [
    "do-worker", "do worker", "l8ai",
    "enterprise agent supply", "OPC incubation", "higher-education digital employees",
    "AI partner platform", "internal agent marketplace",
    "AI agents", "AI coding", "Claude Code", "Codex CLI", "Gemini CLI", "Aider",
    "multi-agent orchestration", "AI skills", "terminal AI", "code automation",
    "developer tools", "enterprise development", "self-hosted", "agent fleet",
    "AI developer tools", "coding agents", "agent workflows",
    "organizational knowledge", "human checkpoints",
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
    title: "Do Worker - Enterprise Agent Supply and OPC Incubation",
    description: "Build, govern, distribute, and operate reusable Agents across organizations.",
    url: "https://agentsmesh.ai",
  },
  twitter: {
    card: "summary_large_image",
    title: "Do Worker - Enterprise Agent Supply and OPC Incubation",
    description: "Build, govern, distribute, and operate reusable Agents across organizations.",
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
