import { StatePanel } from "@/components/state-panel";

export default function NotFound() {
  return (
    <main className="shell page-main">
      <StatePanel
        kind="empty"
        title="此内容当前不可获取"
        description="它可能已下架、暂停发布或更换了访问地址。"
        action={{ href: "/catalog", label: "浏览全部内容" }}
      />
    </main>
  );
}
