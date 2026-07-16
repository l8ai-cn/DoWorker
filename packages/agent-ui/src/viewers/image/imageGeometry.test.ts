import { denormalizePoint, normalizePoint } from "./imageGeometry";

describe("imageGeometry", () => {
  it("normalizes source pixels into normalized coordinates", () => {
    expect(
      normalizePoint(
        { x: 250, y: 125 },
        { width: 1000, height: 500 },
      ),
    ).toEqual({ x: 0.25, y: 0.25 });
  });

  it("denormalizes coordinates for a different rendered size", () => {
    expect(
      denormalizePoint(
        { x: 0.25, y: 0.25 },
        { width: 2000, height: 1000 },
      ),
    ).toEqual({ x: 500, y: 250 });
  });

  it("clamps points to the normalized image bounds", () => {
    expect(
      normalizePoint(
        { x: -20, y: 750 },
        { width: 1000, height: 500 },
      ),
    ).toEqual({ x: 0, y: 1 });
    expect(
      denormalizePoint(
        { x: 1.4, y: -0.2 },
        { width: 1000, height: 500 },
      ),
    ).toEqual({ x: 1000, y: 0 });
  });

  it.each([
    ["zero width", { width: 0, height: 100 }],
    ["zero height", { width: 100, height: 0 }],
  ])("rejects %s", (_label, dimensions) => {
    expect(() => normalizePoint({ x: 1, y: 1 }, dimensions)).toThrow(
      "image_dimensions_must_be_positive",
    );
    expect(() => denormalizePoint({ x: 0.5, y: 0.5 }, dimensions)).toThrow(
      "image_dimensions_must_be_positive",
    );
  });
});
