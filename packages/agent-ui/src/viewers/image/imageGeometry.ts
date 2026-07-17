export interface ImagePoint {
  x: number;
  y: number;
}

export interface ImageDimensions {
  width: number;
  height: number;
}

export function normalizePoint(
  point: ImagePoint,
  dimensions: ImageDimensions,
): ImagePoint {
  assertPositiveDimensions(dimensions);
  return {
    x: clampNormalized(point.x / dimensions.width),
    y: clampNormalized(point.y / dimensions.height),
  };
}

export function denormalizePoint(
  point: ImagePoint,
  dimensions: ImageDimensions,
): ImagePoint {
  assertPositiveDimensions(dimensions);
  return {
    x: clampNormalized(point.x) * dimensions.width,
    y: clampNormalized(point.y) * dimensions.height,
  };
}

function clampNormalized(value: number): number {
  return Math.min(1, Math.max(0, value));
}

function assertPositiveDimensions(dimensions: ImageDimensions): void {
  if (dimensions.width <= 0 || dimensions.height <= 0) {
    throw new Error("image_dimensions_must_be_positive");
  }
}
