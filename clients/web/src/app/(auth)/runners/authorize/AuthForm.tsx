import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select";
import type { LightOrganization } from "@/lib/light-auth";

interface AuthFormProps {
  isSignedIn: boolean;
  userEmail?: string;
  authKey: string;
  organizations: LightOrganization[] | null;
  selectedOrg: LightOrganization | null;
  onSelectOrg: (org: LightOrganization | null) => void;
  nodeIdInput: string;
  onNodeIdChange: (val: string) => void;
  authorizing: boolean;
  onAuthorize: () => void;
  error: string;
  t: (key: string, params?: Record<string, string | number>) => string;
  tCommon: (key: string) => string;
}

export function AuthForm({
  isSignedIn,
  userEmail,
  authKey,
  organizations,
  selectedOrg,
  onSelectOrg,
  nodeIdInput,
  onNodeIdChange,
  authorizing,
  onAuthorize,
  error,
  t,
  tCommon,
}: AuthFormProps) {
  return (
    <div className="space-y-4">
      <div className="flex justify-center">
        <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center">
          <svg className="w-8 h-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
            />
          </svg>
        </div>
      </div>

      {error && (
        <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">{error}</div>
      )}

      {!isSignedIn ? (
        <UnauthenticatedPrompt authKey={authKey} t={t} />
      ) : (
        <AuthenticatedForm
          userEmail={userEmail}
          organizations={organizations}
          selectedOrg={selectedOrg}
          onSelectOrg={onSelectOrg}
          nodeIdInput={nodeIdInput}
          onNodeIdChange={onNodeIdChange}
          authorizing={authorizing}
          onAuthorize={onAuthorize}
          t={t}
          tCommon={tCommon}
        />
      )}
    </div>
  );
}

function UnauthenticatedPrompt({
  authKey,
  t,
}: {
  authKey: string;
  t: (k: string, p?: Record<string, string | number>) => string;
}) {
  return (
    <div className="space-y-3">
      <p className="text-sm text-center text-muted-foreground">{t("signInRequired")}</p>
      <Link href={`/login?redirect=/runners/authorize?key=${authKey}`}>
        <Button className="w-full">{t("signInToAuthorize")}</Button>
      </Link>
      <p className="text-sm text-center text-muted-foreground">
        {t("noAccount")}{" "}
        <Link
          href={`/register?redirect=/runners/authorize?key=${authKey}`}
          className="text-primary hover:underline"
        >
          {t("signUp")}
        </Link>
      </p>
    </div>
  );
}

function AuthenticatedForm({
  userEmail,
  organizations,
  selectedOrg,
  onSelectOrg,
  nodeIdInput,
  onNodeIdChange,
  authorizing,
  onAuthorize,
  t,
  tCommon,
}: {
  userEmail?: string;
  organizations: LightOrganization[] | null;
  selectedOrg: LightOrganization | null;
  onSelectOrg: (org: LightOrganization | null) => void;
  nodeIdInput: string;
  onNodeIdChange: (val: string) => void;
  authorizing: boolean;
  onAuthorize: () => void;
  t: (k: string, p?: Record<string, string | number>) => string;
  tCommon: (k: string) => string;
}) {
  return (
    <div className="space-y-4">
      {userEmail && (
        <p className="text-sm text-center text-muted-foreground">
          {t("signedInAs")} <strong>{userEmail}</strong>
        </p>
      )}

      {organizations && organizations.length > 0 ? (
        <div className="space-y-2">
          <label className="text-sm font-medium text-foreground">{t("selectOrganization")}</label>
          <Select
            value={selectedOrg ? String(selectedOrg.id) : ""}
            onValueChange={(value) => {
              const org = organizations.find((o) => o.id === parseInt(value, 10));
              onSelectOrg(org || null);
            }}
          >
            <SelectTrigger>
              <span className={selectedOrg ? "" : "text-muted-foreground"}>
                {selectedOrg?.name ?? t("selectOrgPlaceholder")}
              </span>
            </SelectTrigger>
            <SelectContent>
              {organizations.map((org) => (
                <SelectItem key={org.id} value={String(org.id)}>
                  {org.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      ) : (
        <div className="p-3 text-sm text-warning bg-warning-bg rounded-md">{t("noOrganizations")}</div>
      )}

      <div className="space-y-2">
        <label className="text-sm font-medium text-foreground">
          {t("nodeIdLabel")}{" "}
          <span className="text-muted-foreground">({tCommon("optional")})</span>
        </label>
        <Input
          placeholder={t("nodeIdPlaceholder")}
          value={nodeIdInput}
          onChange={(e) => onNodeIdChange(e.target.value)}
        />
        <p className="text-xs text-muted-foreground">{t("nodeIdHint")}</p>
      </div>

      <Button className="w-full" onClick={onAuthorize} disabled={authorizing || !selectedOrg}>
        {authorizing ? t("authorizing") : t("authorizeButton")}
      </Button>
    </div>
  );
}
