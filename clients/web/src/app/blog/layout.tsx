import type { Metadata } from "next";

export const metadata: Metadata = {
  title: {
    template: "%s | Do Worker Blog",
    default: "Blog",
  },
};

export default function BlogLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
