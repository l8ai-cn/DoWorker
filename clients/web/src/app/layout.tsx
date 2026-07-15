import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono, Space_Grotesk } from "next/font/google";
import { ThemeProvider, ThemeColorMeta } from "@/components/theme";
import { PWAProvider } from "@/components/pwa";
import { PostHogProvider } from "@/providers/PostHogProvider";
import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages } from "next-intl/server";
import { Toaster } from "sonner";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const spaceGrotesk = Space_Grotesk({
  variable: "--font-space-grotesk",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL("https://agentsmesh.ai"),
  title: {
    default: "Do Worker - Build AI Experts That Work Across Teams",
    template: "%s | Do Worker",
  },
  description: "Do Worker organizes AI capabilities, organizational knowledge, tools, and execution runtimes behind one Expert that carries cross-functional work through to inspectable delivery.",
  keywords: [
    "do-worker", "do worker", "l8ai",
    "AI expert platform", "cross-functional AI", "AI expert orchestration",
    "AI agents", "AI coding", "Claude Code", "Codex CLI", "Gemini CLI", "Aider",
    "multi-agent orchestration", "AI skills", "terminal AI", "code automation",
    "developer tools", "enterprise development", "self-hosted", "agent fleet",
    "AI developer tools", "coding agents", "expert workflows",
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
    title: "Do Worker - Build AI Experts That Work Across Teams",
    description: "Organize AI capabilities, knowledge, tools, and execution runtimes behind one Expert that delivers cross-functional work.",
    url: "https://agentsmesh.ai",
  },
  twitter: {
    card: "summary_large_image",
    title: "Do Worker - Build AI Experts That Work Across Teams",
    description: "Organize AI capabilities, knowledge, tools, and execution runtimes behind one Expert that delivers cross-functional work.",
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
        className={`${geistSans.variable} ${geistMono.variable} ${spaceGrotesk.variable} antialiased bg-background text-foreground`}
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
