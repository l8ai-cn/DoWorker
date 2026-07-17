import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { AttachmentCard } from "../AttachmentCard";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("AttachmentCard", () => {
  it("renders remote html as an explicit external link without embedding it", () => {
    const url = "https://cdn.example.test/report.html";
    const { container } = render(<AttachmentCard url={url} />);

    expect(screen.getByText("report.html")).toBeVisible();
    const link = screen.getByRole("link", { name: "preview" });
    expect(link).toHaveAttribute("href", url);
    expect(link).not.toHaveAttribute("download");
    expect(container.querySelector("iframe")).toBeNull();
  });
});
