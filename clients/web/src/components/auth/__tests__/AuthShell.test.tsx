import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AuthShell } from "@/components/auth/AuthShell";

describe("AuthShell", () => {
  it("allows a paragraph footer without nesting paragraphs", () => {
    const { container } = render(
      <AuthShell
        title="Title"
        subtitle="Subtitle"
        footer={<p>Footer</p>}
      >
        <div>Content</div>
      </AuthShell>,
    );

    expect(screen.getByText("Footer")).toBeInTheDocument();
    expect(container.querySelector("p > p")).toBeNull();
  });
});
