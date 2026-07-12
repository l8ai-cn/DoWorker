import { Box } from "lucide-react";
import Link from "next/link";

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
            <small>应用市场</small>
          </span>
        </Link>
        <nav aria-label="主导航">
          <Link href="/">市场首页</Link>
          <Link href="/catalog">全部应用</Link>
        </nav>
      </div>
    </header>
  );
}
