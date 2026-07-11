import type { ReactNode } from "react";

export function DocsArticle({ children }: { children: ReactNode }) {
  return (
    <article className="[&>div>section]:mb-10 sm:[&>div>section]:mb-12">
      {children}
    </article>
  );
}
