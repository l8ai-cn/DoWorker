import { Button } from "@/components/ui/button";
import type { BillingOverview } from "@/lib/api";
import type { TranslationFn } from "../GeneralSettings";

interface CurrentPlanCardProps {
  plan: BillingOverview["plan"];
  status: string;
  billing_cycle: string;
  current_period_end?: string;
  cancelAtPeriodEnd?: boolean;
  onChangePlan: () => void;
  onCancelPlan: () => void;
  onReactivate: () => void;
  reactivating?: boolean;
  t: TranslationFn;
}

export function CurrentPlanCard({
  plan,
  status,
  billing_cycle,
  current_period_end,
  cancelAtPeriodEnd,
  onChangePlan,
  onCancelPlan,
  onReactivate,
  reactivating,
  t,
}: CurrentPlanCardProps) {
  const isPaidPlan = plan?.price_per_seat_monthly && plan.price_per_seat_monthly > 0;
  const isFrozen = status === "frozen";
  const isCanceled = status === "canceled";
  const isInactive = isFrozen || isCanceled;

  return (
    <div className="surface-card p-6">
      <h2 className="text-lg font-semibold mb-4">{t("settings.billingPage.currentPlan")}</h2>
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h3 className="text-2xl font-bold">{plan?.display_name || plan?.name || t("settings.billingPage.plansDialog.free")}</h3>
            <span className={`text-xs px-2 py-0.5 rounded ${
              status === "active" ? "bg-success-bg text-success" :
              status === "past_due" ? "bg-warning-bg text-warning" :
              status === "frozen" ? "bg-info-bg text-info" :
              status === "canceled" ? "bg-muted text-muted-foreground" :
              "bg-danger-bg text-danger"
            }`}>
              {status === "frozen" ? t("settings.billingPage.frozen") :
               status === "canceled" ? t("settings.billingPage.canceled") :
               status.charAt(0).toUpperCase() + status.slice(1)}
            </span>
            {cancelAtPeriodEnd && !isInactive && (
              <span className="text-xs px-2 py-0.5 rounded bg-warning-bg text-warning">
                {t("settings.billingPage.cancellingAtPeriodEnd")}
              </span>
            )}
          </div>
          <p className="text-muted-foreground">
            {billing_cycle === "yearly" ? t("settings.billingPage.yearly") : t("settings.billingPage.monthly")} billing
            {current_period_end && !isInactive && (
              <> · {cancelAtPeriodEnd ? t("settings.billingPage.endsOn") : t("settings.billingPage.renews")} {new Date(current_period_end).toLocaleDateString()}</>
            )}
            {isFrozen && current_period_end && (
              <> · {t("settings.billingPage.expiredOn")} {new Date(current_period_end).toLocaleDateString()}</>
            )}
            {isCanceled && current_period_end && (
              <> · {t("settings.billingPage.canceledOn")} {new Date(current_period_end).toLocaleDateString()}</>
            )}
          </p>
          {isPaidPlan && (
            <p className="text-sm text-muted-foreground mt-1">
              ${plan.price_per_seat_monthly}/seat/month
            </p>
          )}
          {isFrozen && (
            <p className="text-sm text-info mt-2">
              {t("settings.billingPage.frozenMessage")}
            </p>
          )}
          {isCanceled && (
            <p className="text-sm text-muted-foreground mt-2">
              {t("settings.billingPage.canceledMessage")}
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          {/* Frozen or Canceled state: show resubscribe button */}
          {isInactive && isPaidPlan && (
            <Button variant="default" onClick={onChangePlan}>
              {t("settings.billingPage.resubscribe")}
            </Button>
          )}
          {/* Cancel pending: show reactivate button */}
          {!isInactive && isPaidPlan && cancelAtPeriodEnd && (
            <Button variant="default" onClick={onReactivate} disabled={reactivating}>
              {reactivating ? t("settings.billingPage.reactivating") : t("settings.billingPage.reactivate")}
            </Button>
          )}
          {/* Active state: show cancel button */}
          {!isInactive && isPaidPlan && !cancelAtPeriodEnd && (
            <Button variant="outline" onClick={onCancelPlan}>
              {t("settings.billingPage.cancelPlan")}
            </Button>
          )}
          {/* Active state: show change plan button */}
          {!isInactive && (
            <Button onClick={onChangePlan}>
              {plan?.name === "free" ? t("settings.billingPage.upgrade") : t("settings.billingPage.changePlan")}
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
