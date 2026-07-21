import { ArrowDownRight, ShieldCheck } from "lucide-react";

import type { Market } from "@/lib/marketplace-types";

export function MarketIntro({ market }: { market: Market }) {
  return (
    <section className="market-intro">
      <div>
        <span className="eyebrow">应用市场</span>
        <h1>把可交付的业务结果带进团队。</h1>
        <p>{market.summary}</p>
      </div>
      <div className="market-intro-aside">
        <span className="market-name">{market.name}</span>
        <div className="trust-note">
          <ShieldCheck aria-hidden="true" size={20} />
          <span>
            <strong>先评估，再启用</strong>
            <small>查看结果、接入条件、权限和市场额度后，再跳转至 Agent Cloud 完成启用。</small>
          </span>
        </div>
        <span className="intro-direction">
          从工作场景开始
          <ArrowDownRight aria-hidden="true" size={18} />
        </span>
      </div>
    </section>
  );
}
