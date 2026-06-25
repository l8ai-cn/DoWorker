"use client";

import { useServerUrl } from "@/hooks/useServerUrl";
import { useTranslations } from "next-intl";
import { DocNavigation } from "@/components/docs/DocNavigation";
import { EnvVarsSection } from "./_sections/EnvVarsSection";

export default function RunnerSetupPage() {
  const serverUrl = useServerUrl();
  const t = useTranslations();

  return (
    <div>
      <h1 className="text-4xl font-bold mb-8">
        {t("docs.runners.setup.title")}
      </h1>

      <p className="text-muted-foreground leading-relaxed mb-8">
        {t("docs.runners.setup.description")}
      </p>

      {/* Requirements */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.runners.setup.requirements.title")}
        </h2>
        <ul className="list-disc list-inside text-muted-foreground space-y-2">
          <li>{t("docs.runners.setup.requirements.item1")}</li>
          <li>{t("docs.runners.setup.requirements.item2")}</li>
          <li>{t("docs.runners.setup.requirements.item3")}</li>
          <li>{t("docs.runners.setup.requirements.item4")}</li>
          <li>{t("docs.runners.setup.requirements.item5")}</li>
        </ul>
      </section>

      {/* Quick Install */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.runners.setup.quickInstall.title")}
        </h2>

        <h3 className="text-lg font-medium mb-2 mt-6">
          {t("docs.runners.setup.quickInstall.oneLineTitle")}
        </h3>
        <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm overflow-x-auto">
          <pre className="text-success">{`# macOS / Linux
curl -fsSL ${serverUrl}/install.sh | sh

# Windows (PowerShell)
irm ${serverUrl}/install.ps1 | iex`}</pre>
        </div>

        <h3 className="text-lg font-medium mb-2 mt-6">
          {t("docs.runners.setup.quickInstall.linuxTitle")}
        </h3>
        <p className="text-sm text-muted-foreground mb-2">
          The one-line script above also works on Linux. Alternatively, download .deb/.rpm/.apk packages from{" "}
          <a
            href="https://github.com/AgentsMesh/AgentsMesh/releases/latest"
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary underline"
          >
            GitHub Releases
          </a>.
        </p>

        <h3 className="text-lg font-medium mb-2 mt-6">
          {t("docs.runners.setup.quickInstall.afterInstall")}
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
            <h4 className="font-medium mb-2 text-sm">
              {t("docs.runners.setup.quickInstall.methodToken")}
            </h4>
            <div className="font-mono text-sm overflow-x-auto">
              <pre className="text-success">{`agentsmesh-runner register \\
  --server ${serverUrl} \\
  --token <YOUR_TOKEN>
agentsmesh-runner run`}</pre>
            </div>
          </div>
          <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4">
            <h4 className="font-medium mb-2 text-sm">
              {t("docs.runners.setup.quickInstall.methodLogin")}
            </h4>
            <div className="font-mono text-sm overflow-x-auto">
              <pre className="text-success">{`agentsmesh-runner login
agentsmesh-runner run`}</pre>
            </div>
          </div>
        </div>

        <p className="text-sm text-muted-foreground mt-4">
          {t("docs.runners.setup.quickInstall.tokenHint")}
        </p>
      </section>

      {/* Docker Installation */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.runners.setup.docker.title")}
        </h2>
        <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm overflow-x-auto">
          <pre className="text-success">{`# Run with Docker
docker run -d \\
  --name agentsmesh-runner \\
  -e AGENTSMESH_TOKEN=<YOUR_TOKEN> \\
  -e AGENTSMESH_URL=${serverUrl} \\
  -v /var/run/docker.sock:/var/run/docker.sock \\
  -v ~/.ssh:/root/.ssh:ro \\
  agentsmesh/runner:latest`}</pre>
        </div>
      </section>

      {/* Docker Compose */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.runners.setup.dockerCompose.title")}
        </h2>
        <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 font-mono text-sm overflow-x-auto">
          <pre className="text-success">{`# docker-compose.yml
version: '3.8'
services:
  runner:
    image: agentsmesh/runner:latest
    container_name: agentsmesh-runner
    restart: unless-stopped
    environment:
      - AGENTSMESH_TOKEN=\${AGENTSMESH_TOKEN}
      - AGENTSMESH_URL=\${AGENTSMESH_URL:-${serverUrl}}
      - MAX_CONCURRENT_PODS=5
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ~/.ssh:/root/.ssh:ro
      - runner-data:/data
volumes:
  runner-data:`}</pre>
        </div>
      </section>

      <EnvVarsSection serverUrl={serverUrl} />

      {/* Registration Token */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.runners.setup.registrationToken.title")}
        </h2>
        <ol className="list-decimal list-inside text-muted-foreground space-y-2">
          <li>{t("docs.runners.setup.registrationToken.step1")}</li>
          <li>{t("docs.runners.setup.registrationToken.step2")}</li>
          <li>{t("docs.runners.setup.registrationToken.step3")}</li>
          <li>{t("docs.runners.setup.registrationToken.step4")}</li>
          <li>{t("docs.runners.setup.registrationToken.step5")}</li>
        </ol>
        <div className="rounded-lg bg-surface-muted ring-1 ring-border/15 p-4 mt-4">
          <p className="text-sm text-muted-foreground">
            {t("docs.runners.setup.registrationToken.securityNote")}
          </p>
        </div>
      </section>

      {/* Verifying Installation */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.runners.setup.verifying.title")}
        </h2>
        <p className="text-muted-foreground mb-4">
          {t("docs.runners.setup.verifying.description")}
        </p>
        <ol className="list-decimal list-inside text-muted-foreground space-y-2">
          <li>{t("docs.runners.setup.verifying.step1")}</li>
          <li>{t("docs.runners.setup.verifying.step2")}</li>
          <li>{t("docs.runners.setup.verifying.step3")}</li>
        </ol>
      </section>

      {/* Troubleshooting */}
      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4 text-foreground">
          {t("docs.runners.setup.troubleshooting.title")}
        </h2>
        <div className="space-y-4">
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.runners.setup.troubleshooting.offline")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.runners.setup.troubleshooting.offlineDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.runners.setup.troubleshooting.podsFail")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.runners.setup.troubleshooting.podsFailDesc")}
            </p>
          </div>
          <div className="surface-card p-4">
            <h3 className="font-medium mb-2">
              {t("docs.runners.setup.troubleshooting.tokenInvalid")}
            </h3>
            <p className="text-sm text-muted-foreground">
              {t("docs.runners.setup.troubleshooting.tokenInvalidDesc")}
            </p>
          </div>
        </div>
      </section>

      <DocNavigation />
    </div>
  );
}
