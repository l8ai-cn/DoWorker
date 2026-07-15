type SkillDependency = {
  slug: string;
  version: number;
};

export function SkillDependencies({ value }: { value: string }) {
  const result = parseDependencies(value);
  return (
    <section>
      <h3 className="mb-2 text-sm font-semibold">Skill 依赖</h3>
      {result.error ? (
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          Skill 依赖 JSON 无法解析：{result.error}
        </div>
      ) : result.items.length === 0 ? (
        <p className="rounded-md border border-border p-3 text-sm text-muted-foreground">
          无 Skill 依赖
        </p>
      ) : (
        <div className="divide-y divide-border rounded-md border border-border">
          {result.items.map((item, index) => (
            <div
              key={`${item.slug}-${item.version}-${index}`}
              className="flex items-center justify-between gap-4 px-3 py-2"
            >
              <span className="text-sm font-medium">{item.slug}</span>
              <span className="text-xs text-muted-foreground">
                版本 {item.version}
              </span>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function parseDependencies(
  value: string,
): { items: SkillDependency[]; error?: string } {
  try {
    const parsed: unknown = JSON.parse(value);
    if (!Array.isArray(parsed)) {
      return { items: [], error: "根节点必须是数组" };
    }
    const items = parsed.map((item, index) => {
      if (!item || typeof item !== "object") {
        throw new Error(`第 ${index + 1} 项必须是对象`);
      }
      const record = item as Record<string, unknown>;
      if (typeof record.slug !== "string" || !record.slug.trim()) {
        throw new Error(`第 ${index + 1} 项缺少 slug`);
      }
      if (!Number.isInteger(record.version) || Number(record.version) <= 0) {
        throw new Error(`第 ${index + 1} 项的 version 无效`);
      }
      return {
        slug: record.slug,
        version: Number(record.version),
      };
    });
    return { items };
  } catch (error) {
    return {
      items: [],
      error: error instanceof Error ? error.message : "未知解析错误",
    };
  }
}
