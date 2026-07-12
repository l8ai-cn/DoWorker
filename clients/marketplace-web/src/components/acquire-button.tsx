import { ArrowUpRight } from "lucide-react";

import { buildAcquireLink } from "@/lib/acquire-link";
import { acquireLabels } from "@/lib/listing-presentation";
import type { ResourceType } from "@/lib/marketplace-types";

interface AcquireButtonProps {
  coreWebUrl: string | undefined;
  resourceType: ResourceType;
  target: {
    market: string;
    listing: string;
    version: string;
  };
}

export function AcquireButton({
  coreWebUrl,
  resourceType,
  target,
}: AcquireButtonProps) {
  const label = acquireLabels[resourceType];
  const href = buildAcquireLink(coreWebUrl, target);

  if (resourceType !== "application") {
    return (
      <p className="helper-text">
        该资源用于应用装配，当前不提供单独启用。
      </p>
    );
  }

  if (!href) {
    return (
      <div className="acquire-action">
        <button className="button button-primary" type="button" disabled>
          {label}
        </button>
        <span className="helper-text">获取入口尚未配置</span>
      </div>
    );
  }

  return (
    <a className="button button-primary" href={href}>
      {label}
      <ArrowUpRight aria-hidden="true" size={17} />
    </a>
  );
}
