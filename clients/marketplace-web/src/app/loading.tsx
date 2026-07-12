export default function Loading() {
  return (
    <main className="shell page-main" aria-label="正在加载市场内容">
      <div className="skeleton skeleton-intro" />
      <div className="skeleton skeleton-search" />
      <div className="skeleton-grid">
        {Array.from({ length: 6 }, (_, index) => (
          <div className="skeleton skeleton-card" key={index} />
        ))}
      </div>
    </main>
  );
}
