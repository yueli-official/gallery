export function artworkAspectRatio(width: number, height: number): string {
  return width > 0 && height > 0 ? `${width} / ${height}` : "4 / 5";
}

export function contentRatingLabel(
  rating: "general" | "sensitive" | "adult",
): string {
  return {
    general: "全年龄",
    sensitive: "敏感内容",
    adult: "成人内容",
  }[rating];
}
