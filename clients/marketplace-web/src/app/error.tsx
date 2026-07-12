"use client";

import { AlertCircle } from "lucide-react";

export default function ErrorPage({ reset }: { reset: () => void }) {
  return (
    <main className="shell page-main">
      <section className="state-panel state-error">
        <span className="state-icon">
          <AlertCircle aria-hidden="true" size={24} />
        </span>
        <h1>市场内容加载失败</h1>
        <p>服务暂时不可用，请稍后重试。如果问题持续，请联系市场管理员。</p>
        <button className="button button-secondary" type="button" onClick={reset}>
          重新加载
        </button>
      </section>
    </main>
  );
}
