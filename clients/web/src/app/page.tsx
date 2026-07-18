import { Navbar, FinalCTA, Footer } from "@/components/landing";
import { HomeSessionRedirect } from "@/components/landing/HomeSessionRedirect";
import { ExpertHome } from "@/components/landing/expert-home/ExpertHome";

export default function Home() {
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "Do Worker",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web, Linux, macOS, Windows",
    description:
      "Do Worker builds, governs, distributes, and operates reusable Agents for enterprise teams, OPC founders, and higher-education digital employee pilots.",
    alternateName: ["AgentMesh", "Agents Mesh"],
    url: "https://agentsmesh.ai",
    keywords:
      "enterprise Agent supply, OPC incubation, higher-education digital employees, internal Agent marketplace, AI partners, self-hosted AI agents",
    publisher: {
      "@type": "Organization",
      name: "Do Worker",
      url: "https://agentsmesh.ai",
      logo: "https://agentsmesh.ai/icons/icon-512.png",
      sameAs: [
        "https://github.com/l8ai-cn/DoWorker",
        "https://x.com/agentsmeshai",
        "https://discord.gg/3RcX7VBbH9",
      ],
    },
  };

  return (
    <div className="azure-theme expert-home min-h-screen bg-[var(--expert-bg)]">
      <HomeSessionRedirect />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <Navbar />
      <main>
        <ExpertHome />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
