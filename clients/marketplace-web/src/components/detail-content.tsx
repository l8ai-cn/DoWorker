import { Check, ExternalLink } from "lucide-react";

import type { ListingDetail } from "@/lib/marketplace-types";

function ListSection({
  title,
  items,
}: {
  title: string;
  items: string[];
}) {
  if (!items.length) return null;
  return (
    <section className="detail-section">
      <h2>{title}</h2>
      <ul className="check-list">
        {items.map((item) => (
          <li key={item}>
            <Check aria-hidden="true" size={17} />
            {item}
          </li>
        ))}
      </ul>
    </section>
  );
}

export function DetailContent({ listing }: { listing: ListingDetail }) {
  return (
    <div className="detail-content">
      <main>
        <section className="detail-section">
          <h2>能力说明</h2>
          <p className="long-copy">{listing.description}</p>
        </section>
        <ListSection title="可以完成什么" items={listing.outcomes || []} />
        <ListSection title="适用场景" items={listing.use_cases || []} />
        <ListSection title="适用对象" items={listing.target_audience || []} />
        <ListSection title="应用包含什么" items={listing.package_summary || []} />
      </main>
      <aside>
        {listing.first_task && (
          <section className="detail-section first-task">
            <h2>启用后从这里开始</h2>
            <strong>{listing.first_task.title}</strong>
            <p className="long-copy">{listing.first_task.description}</p>
          </section>
        )}
        <ListSection title="启用要求" items={listing.requirements || []} />
        <ListSection title="所需权限" items={listing.permissions || []} />
        <section className="detail-section">
          <h2>版本说明</h2>
          <p className="long-copy">{listing.release_notes || "暂无版本说明。"}</p>
          <div className="support-links">
            {listing.documentation_url && (
              <a href={listing.documentation_url}>
                查看文档
                <ExternalLink aria-hidden="true" size={15} />
              </a>
            )}
            {listing.support_url && (
              <a href={listing.support_url}>
                获取支持
                <ExternalLink aria-hidden="true" size={15} />
              </a>
            )}
          </div>
        </section>
      </aside>
    </div>
  );
}
