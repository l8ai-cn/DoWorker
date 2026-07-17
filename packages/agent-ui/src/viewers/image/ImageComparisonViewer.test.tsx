import { fireEvent, render, screen } from "@testing-library/react";

import { ImageComparisonViewer } from "./ImageComparisonViewer";

const source = { alt: "产品源图", src: "/source.png" };
const result = { alt: "产品结果图", src: "/result.png" };

describe("ImageComparisonViewer", () => {
  it("offers accessible controls for every comparison mode", () => {
    render(<ImageComparisonViewer result={result} source={source} />);

    expect(screen.getByRole("button", { name: "查看源图" })).toBeVisible();
    expect(screen.getByRole("button", { name: "查看结果图" })).toBeVisible();
    expect(screen.getByRole("button", { name: "并排比较" })).toBeVisible();
    expect(screen.getByRole("button", { name: "滑块比较" })).toBeVisible();
  });

  it("switches between source, result, and side-by-side views", () => {
    render(<ImageComparisonViewer result={result} source={source} />);

    expect(screen.getByRole("img", { name: "产品源图" })).toBeVisible();
    expect(
      screen.queryByRole("img", { name: "产品结果图" }),
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "查看结果图" }));
    expect(screen.getByRole("img", { name: "产品结果图" })).toBeVisible();
    expect(
      screen.queryByRole("img", { name: "产品源图" }),
    ).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "并排比较" }));
    expect(screen.getByRole("img", { name: "产品源图" })).toBeVisible();
    expect(screen.getByRole("img", { name: "产品结果图" })).toBeVisible();
  });

  it("uses a range control to clip the result layer", () => {
    render(<ImageComparisonViewer result={result} source={source} />);

    fireEvent.click(screen.getByRole("button", { name: "滑块比较" }));
    const slider = screen.getByRole("slider", { name: "比较位置" });
    const resultLayer = screen.getByTestId("image-comparison-result-layer");

    expect(slider).toHaveValue("50");
    expect(resultLayer).toHaveStyle({
      clipPath: "inset(0 50% 0 0)",
    });

    fireEvent.change(slider, { target: { value: "30" } });

    expect(slider).toHaveValue("30");
    expect(resultLayer).toHaveStyle({
      clipPath: "inset(0 70% 0 0)",
    });
  });

  it("keeps the slider viewport constrained by narrow hosts", () => {
    render(
      <ImageComparisonViewer
        defaultMode="slider"
        result={result}
        source={source}
      />,
    );

    const viewport = screen.getByTestId("image-comparison-slider-viewport");
    expect(viewport).toHaveClass("w-full");
    expect(viewport).not.toHaveClass("min-h-56");
  });
});
