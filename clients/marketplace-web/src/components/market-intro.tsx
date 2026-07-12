import { ShieldCheck, Sparkles } from "lucide-react";

import type { Market } from "@/lib/marketplace-types";

export function MarketIntro({ market }: { market: Market }) {
  return (
    <section className="market-intro">
      <div>
        <span className="eyebrow">DO WORKER MARKETPLACE</span>
        <h1>{market.name}</h1>
        <p>{market.summary}</p>
      </div>
      <div className="trust-note">
        <ShieldCheck aria-hidden="true" size={20} />
        <span>
          <strong>经过验证的工作能力</strong>
          <small>清晰查看权限、额度与运行要求后再启用</small>
        </span>
      </div>
      <Sparkles className="intro-mark" aria-hidden="true" size={64} />
    </section>
  );
}
