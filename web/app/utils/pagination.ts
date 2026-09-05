export function galleryPageCount(
  page: { total: number; size: number } | null | undefined,
): number {
  return page && page.size > 0 ? Math.ceil(page.total / page.size) : 0;
}
