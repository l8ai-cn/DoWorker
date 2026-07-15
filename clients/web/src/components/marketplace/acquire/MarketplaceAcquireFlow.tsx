"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { ArrowLeft, ArrowRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useLightSession } from "@/hooks/useLightSession";
import {
  lightListOrganizations,
  type LightOrganization,
} from "@/lib/light-auth";
import {
  applyInstallationPlan,
  createInstallationPlan,
  fetchMarketplaceListing,
  MarketplaceAcquireError,
  type InstallationPlan,
  type MarketplaceListingDetail,
} from "@/lib/marketplace/acquire-api";
import { MarketplaceAcquireSummary } from "./MarketplaceAcquireSummary";
import {
  AcquireShell,
  ErrorState,
  InlineError,
  LoadingState,
  OrganizationStep,
  SuccessState,
} from "./MarketplaceAcquireStates";
import { useMarketplaceRuntimeModels } from "./useMarketplaceRuntimeModels";

type Step = "select" | "confirm" | "installing" | "success";

export function MarketplaceAcquireFlow({
  organizationSlug,
}: {
  organizationSlug?: string;
}) {
  const router = useRouter();
  const params = useSearchParams();
  const { session, hydrated } = useLightSession();
  const marketSlug = params.get("market") ?? "";
  const listingSlug = params.get("listing") ?? "";
  const requestedVersion = params.get("version") ?? "";
  const [listing, setListing] = useState<MarketplaceListingDetail | null>(null);
  const [organizations, setOrganizations] = useState<LightOrganization[]>([]);
  const [loadingOrganizations, setLoadingOrganizations] = useState(true);
  const [organizationID, setOrganizationID] = useState("");
  const [plan, setPlan] = useState<InstallationPlan | null>(null);
  const [installationID, setInstallationID] = useState("");
  const [step, setStep] = useState<Step>("select");
  const [error, setError] = useState("");
  const selectedOrganization = organizations.find(
    (organization) => String(organization.id) === organizationID,
  );
  const runtimeModels = useMarketplaceRuntimeModels(
    selectedOrganization?.slug,
    listing?.agent_slug,
  );

  useEffect(() => {
    if (!marketSlug || !listingSlug) {
      setError("启用链接不完整，请返回市场重新选择应用。");
      return;
    }
    fetchMarketplaceListing(marketSlug, listingSlug)
      .then(setListing)
      .catch((cause) => setError(errorMessage(cause)));
  }, [marketSlug, listingSlug]);

  useEffect(() => {
    if (!hydrated || !session?.isAuthenticated) return;
    setLoadingOrganizations(true);
    lightListOrganizations()
      .then(setOrganizations)
      .catch(() => setError("组织列表加载失败，请刷新后重试。"))
      .finally(() => setLoadingOrganizations(false));
  }, [hydrated, session?.isAuthenticated]);

  useEffect(() => {
    if (!organizationSlug || organizations.length === 0) return;
    const organization = organizations.find((item) => item.slug === organizationSlug);
    if (!organization) {
      setError("你没有在当前组织启用市场内容的权限。");
      return;
    }
    setOrganizationID(String(organization.id));
  }, [organizationSlug, organizations]);

  if (!hydrated || (!listing && !error)) {
    return <AcquireShell><LoadingState /></AcquireShell>;
  }
  if (!listing) {
    return <AcquireShell><ErrorState message={error} /></AcquireShell>;
  }
  if (!session?.isAuthenticated) {
    const redirect = organizationSlug
      ? `/${organizationSlug}/marketplace/acquire?${params.toString()}`
      : `/marketplace/acquire?${params.toString()}`;
    router.replace(`/login?redirect=${encodeURIComponent(redirect)}`);
    return <AcquireShell><LoadingState label="正在前往登录" /></AcquireShell>;
  }
  if (error && step === "select") {
    return <AcquireShell><ErrorState message={error} /></AcquireShell>;
  }

  async function preparePlan() {
    if (!selectedOrganization || !listing || !runtimeModels.modelResourceID) return;
    setError("");
    try {
      const result = await createInstallationPlan(
        marketSlug,
        listingSlug,
        requestedVersion || listing.listing_version_id,
        selectedOrganization.id,
        Number(runtimeModels.modelResourceID),
      );
      setPlan(result);
      setStep("confirm");
    } catch (cause) {
      setError(errorMessage(cause));
    }
  }

  async function install() {
    if (!plan) return;
    setStep("installing");
    setError("");
    try {
      const result = await applyInstallationPlan(plan);
      if (result.status !== "succeeded") {
        throw new Error("启用操作尚未完成，请稍后查看操作状态。");
      }
      setInstallationID(result.installation_id);
      setStep("success");
    } catch (cause) {
      setStep("confirm");
      setError(errorMessage(cause));
    }
  }

  return (
    <AcquireShell>
      <header className="border-b border-border pb-6">
        <Link
          href={organizationSlug ? `/${organizationSlug}/marketplace/${listingSlug}` : "/marketplace"}
          className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          返回应用市场
        </Link>
        <p className="mt-6 text-sm font-medium text-primary">专家应用启用向导</p>
        <h1 className="mt-2 text-3xl font-semibold text-foreground">{listing.display_name}</h1>
        <p className="mt-3 text-base leading-7 text-muted-foreground">{listing.tagline}</p>
      </header>

      {step === "select" ? (
        <OrganizationStep
          organizations={organizations}
          loadingOrganizations={loadingOrganizations}
          value={organizationID}
          onChange={setOrganizationID}
          onContinue={preparePlan}
          fixedOrganization={organizationSlug ? selectedOrganization : undefined}
          modelResources={runtimeModels.modelResources}
          modelResourceID={runtimeModels.modelResourceID}
          onModelChange={runtimeModels.setModelResourceID}
          loadingModels={runtimeModels.loadingModels}
          modelError={runtimeModels.modelError}
          incompatibleListing={runtimeModels.incompatibleListing}
          onReloadModels={runtimeModels.reloadModels}
          settingsHref={
            selectedOrganization
              ? `/${selectedOrganization.slug}/settings?tab=ai-resources`
              : ""
          }
        />
      ) : null}
      {step === "confirm" && plan && selectedOrganization ? (
        <div className="space-y-6">
          <MarketplaceAcquireSummary listing={listing} organizationName={selectedOrganization.name} plan={plan} />
          {error ? <InlineError message={error} /> : null}
          <Button className="w-full gap-2" size="lg" onClick={install}>
            确认并启用
            <ArrowRight className="h-4 w-4" />
          </Button>
        </div>
      ) : null}
      {step === "installing" ? <LoadingState label="正在创建专家应用实例" /> : null}
      {step === "success" && selectedOrganization && installationID ? (
        <SuccessState organization={selectedOrganization} installationID={installationID} />
      ) : null}
    </AcquireShell>
  );
}

function errorMessage(cause: unknown): string {
  if (cause instanceof MarketplaceAcquireError) return cause.message;
  if (cause instanceof Error) return cause.message;
  return "启用失败，请稍后重试。";
}
