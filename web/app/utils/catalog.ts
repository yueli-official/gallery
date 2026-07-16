export type GalleryCatalogSort =
  "newest" | "oldest" | "title_asc" | "title_desc";
export type GalleryCatalogView = "grid" | "masonry";

export interface GalleryCatalogState {
  q: string;
  sort: GalleryCatalogSort;
  page: number;
  categories: string[];
  facets: string[];
  tag: string;
  preview: string;
  view: GalleryCatalogView;
}

export interface GalleryCatalogRequest {
  q?: string;
  sort: GalleryCatalogSort;
  page: number;
  size: number;
  categories?: string;
  facets?: string;
  tag?: string;
}

const allowedSorts = new Set<GalleryCatalogSort>([
  "newest",
  "oldest",
  "title_asc",
  "title_desc",
]);

function scalar(value: unknown): string {
  if (Array.isArray(value)) return scalar(value[0]);
  if (typeof value === "string" || typeof value === "number") {
    return String(value).trim();
  }
  return "";
}

function list(value: unknown): string[] {
  return [
    ...new Set(
      scalar(value)
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
    ),
  ];
}

export function parseGalleryCatalogState(
  query: Record<string, unknown>,
): GalleryCatalogState {
  const requestedSort = scalar(query.sort) as GalleryCatalogSort;
  const requestedPage = Number.parseInt(scalar(query.page), 10);

  return {
    q: scalar(query.q),
    sort: allowedSorts.has(requestedSort) ? requestedSort : "newest",
    page:
      Number.isFinite(requestedPage) && requestedPage > 0 ? requestedPage : 1,
    categories: list(query.categories),
    facets: list(query.facets),
    tag: scalar(query.tag),
    preview: scalar(query.preview),
    view: scalar(query.view) === "masonry" ? "masonry" : "grid",
  };
}

export function galleryCatalogRequest(
  state: GalleryCatalogState,
  size = 24,
): GalleryCatalogRequest {
  return {
    q: state.q || undefined,
    sort: state.sort,
    page: state.page,
    size,
    categories: state.categories.join(",") || undefined,
    facets: state.facets.join(",") || undefined,
    tag: state.tag || undefined,
  };
}

export function galleryCatalogQuery(
  state: GalleryCatalogState,
): Record<string, string | number> {
  const query: Record<string, string | number> = {};
  if (state.q) query.q = state.q;
  if (state.sort !== "newest") query.sort = state.sort;
  if (state.page > 1) query.page = state.page;
  if (state.categories.length) query.categories = state.categories.join(",");
  if (state.facets.length) query.facets = state.facets.join(",");
  if (state.tag) query.tag = state.tag;
  if (state.preview) query.preview = state.preview;
  if (state.view === "masonry") query.view = state.view;
  return query;
}

export function galleryCatalogFilterCount(state: GalleryCatalogState): number {
  return (
    state.categories.length +
    state.facets.length +
    (state.q ? 1 : 0) +
    (state.tag ? 1 : 0)
  );
}
