interface ReleaseJsonSnapshotProps {
  title: string;
  value: string;
}

export function ReleaseJsonSnapshot({
  title,
  value,
}: ReleaseJsonSnapshotProps) {
  const result = parseJson(value);
  return (
    <section>
      <h3 className="mb-2 text-sm font-semibold">{title}</h3>
      {result.error ? (
        <div className="rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">
          快照 JSON 无法解析：{result.error}
        </div>
      ) : (
        <pre className="max-h-72 overflow-auto rounded-md bg-muted p-3 text-xs leading-5">
          {JSON.stringify(result.value, null, 2)}
        </pre>
      )}
    </section>
  );
}

function parseJson(value: string): { value?: unknown; error?: string } {
  try {
    return { value: JSON.parse(value) };
  } catch (error) {
    return {
      error: error instanceof Error ? error.message : "未知解析错误",
    };
  }
}
