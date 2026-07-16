import type { MarketplaceToolModelGroup } from "@/lib/marketplace-tool-model-resources";

export function MarketplaceToolModelFields({
  groups,
  values,
  onChange,
  disabled,
}: {
  groups: MarketplaceToolModelGroup[];
  values: Record<string, string>;
  onChange: (role: string, value: string) => void;
  disabled?: boolean;
}) {
  return groups.map((group) => (
    <label key={group.role} className="block space-y-2">
      <span className="text-sm font-medium text-foreground">
        {toolModelLabel(group.role)}
      </span>
      <select
        value={values[group.role] ?? ""}
        onChange={(event) => onChange(group.role, event.target.value)}
        disabled={disabled}
        className="h-12 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
        aria-label={`选择${toolModelLabel(group.role)}`}
      >
        <option value="">请选择工具模型</option>
        {group.resources.map((resource) => (
          <option key={resource.id} value={resource.id}>
            {resource.label}
          </option>
        ))}
      </select>
    </label>
  ));
}

function toolModelLabel(role: string): string {
  if (role === "seedance-video") return "视频生成模型";
  return role
    .split("-")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
