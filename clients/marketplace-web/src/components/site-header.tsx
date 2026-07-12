import { Box, CircleUserRound } from "lucide-react";
import Link from "next/link";

function PendingEntry({ label }: { label: string }) {
  return (
    <span className="pending-entry">
      {label}
      <small>即将开放</small>
    </span>
  );
}

export function SiteHeader() {
  return (
    <header className="site-header">
      <div className="shell header-inner">
        <Link className="brand" href="/">
          <span className="brand-mark">
            <Box aria-hidden="true" size={19} strokeWidth={2.2} />
          </span>
          <span>
            <strong>Do Worker</strong>
            <small>专家应用市场</small>
          </span>
        </Link>
        <nav aria-label="主导航">
          <Link href="/#spaces">专区</Link>
          <Link href="/catalog">全部内容</Link>
          <PendingEntry label="我的应用" />
          <PendingEntry label="额度" />
        </nav>
        <span className="account-entry" aria-label="账户入口未配置">
          <CircleUserRound aria-hidden="true" size={20} />
          <PendingEntry label="账户" />
        </span>
      </div>
    </header>
  );
}
