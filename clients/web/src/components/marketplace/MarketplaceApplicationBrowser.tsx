"use client";

import { useDeferredValue, useState } from "react";
import { Search, Sparkles } from "lucide-react";

import { Input } from "@/components/ui/input";
import type { PublicMarketApplication } from "@/lib/public-market-api";
import { MarketplaceApplicationCard } from "./MarketplaceApplicationCard";

interface Props {
  applications: PublicMarketApplication[];
  loadError?: string;
}

export function MarketplaceApplicationBrowser({ applications, loadError }: Props) {
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("全部");
  const deferredQuery = useDeferredValue(query.trim().toLowerCase());
  const categories = ["全部", ...Array.from(new Set(applications.map((app) => app.category)))];
  const visibleApplications = applications.filter((app) => {
    const matchesCategory = category === "全部" || app.category === category;
    const haystack = [
      app.name,
      app.summary,
      app.description,
      ...app.tags,
      ...app.skill_slugs,
    ].join(" ").toLowerCase();
    return matchesCategory && (!deferredQuery || haystack.includes(deferredQuery));
  });

  return (
    <main>
      <section className="border-b border-border bg-background">
        <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <div className="flex items-center gap-2 text-sm font-medium text-primary">
                <Sparkles className="h-4 w-4" />
                平台精选专家
              </div>
              <h1 className="mt-3 text-3xl font-semibold text-foreground sm:text-4xl">
                专家应用市场
              </h1>
              <p className="mt-3 max-w-2xl text-base leading-7 text-muted-foreground">
                选择一个已经装配好的 AI 专家，直接带着任务目标、工作方式和能力组件进入工作。
              </p>
            </div>
            <div className="text-sm text-muted-foreground">
              已上架 <span className="font-semibold text-foreground">{applications.length}</span> 个专家应用
            </div>
          </div>

          <div className="mt-8 flex flex-col gap-4 lg:flex-row lg:items-center">
            <div className="relative flex-1">
              <Search className="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="搜索专家、任务场景或 Skill"
                aria-label="搜索专家应用"
                className="h-11 bg-background pl-10"
              />
            </div>
            <div className="flex flex-wrap gap-2" aria-label="专家应用分类">
              {categories.map((item) => (
                <button
                  key={item}
                  type="button"
                  onClick={() => setCategory(item)}
                  className={
                    category === item
                      ? "h-9 rounded-md bg-foreground px-4 text-sm font-medium text-background"
                      : "h-9 rounded-md border border-border bg-background px-4 text-sm text-muted-foreground hover:border-border-strong hover:text-foreground"
                  }
                >
                  {item}
                </button>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        {loadError ? (
          <MarketplaceMessage
            title="专家应用暂时无法加载"
            description="市场服务返回异常，请稍后刷新页面。"
          />
        ) : visibleApplications.length === 0 ? (
          <MarketplaceMessage
            title="没有匹配的专家应用"
            description="调整搜索词或切换分类后再试。"
          />
        ) : (
          <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
            {visibleApplications.map((application) => (
              <MarketplaceApplicationCard key={application.slug} application={application} />
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

function MarketplaceMessage({ title, description }: { title: string; description: string }) {
  return (
    <div className="rounded-md border border-border bg-card px-6 py-16 text-center">
      <h2 className="text-lg font-semibold text-foreground">{title}</h2>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
    </div>
  );
}
