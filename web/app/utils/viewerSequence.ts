import type { GalleryImageCard } from "../types/gallery";
import type { GalleryCatalogRequest } from "./catalog";

export interface ViewerSequenceState {
  ids: string[];
  index: number;
}

export interface ViewerCatalogContinuation {
  request: Omit<GalleryCatalogRequest, "page">;
  previousPage?: number;
  nextPage?: number;
}

export interface ViewerNavigationSession {
  sequence: ViewerSequenceState;
  candidates: GalleryImageCard[];
  catalog?: ViewerCatalogContinuation;
}

function uniqueIds(ids: string[]): string[] {
  const seen = new Set<string>();
  return ids.filter((id) => {
    const value = id.trim();
    if (!value || seen.has(value)) return false;
    seen.add(value);
    return true;
  });
}

export function createViewerSequence(
  currentId: string,
  seedIds: string[] = [],
): ViewerSequenceState {
  const ids = uniqueIds(
    seedIds.includes(currentId) ? seedIds : [currentId, ...seedIds],
  );
  return { ids, index: Math.max(0, ids.indexOf(currentId)) };
}

export function extendViewerSequence(
  state: ViewerSequenceState,
  candidateIds: string[],
): ViewerSequenceState {
  const ids = uniqueIds([...state.ids, ...candidateIds]);
  return { ids, index: Math.min(state.index, Math.max(0, ids.length - 1)) };
}

export function prependViewerSequence(
  state: ViewerSequenceState,
  candidateIds: string[],
): ViewerSequenceState {
  const existing = new Set(state.ids);
  const preceding = uniqueIds(candidateIds).filter((id) => !existing.has(id));
  return {
    ids: [...preceding, ...state.ids],
    index: state.index + preceding.length,
  };
}

export function moveViewerSequence(
  state: ViewerSequenceState,
  targetId: string,
): ViewerSequenceState {
  const existing = state.ids.indexOf(targetId);
  if (existing >= 0) return { ids: state.ids, index: existing };
  return { ids: [...state.ids, targetId], index: state.ids.length };
}

export function viewerCloseTarget(back?: string): string {
  if (!back || /^\/images\/[^/?#]+(?:[?#]|$)/.test(back)) return "/images";
  return back.startsWith("/") ? back : "/images";
}

export function createViewerNavigationSession(
  currentId = "",
): ViewerNavigationSession {
  return {
    sequence: createViewerSequence(currentId),
    candidates: [],
  };
}

export function createCatalogViewerNavigationSession(
  items: GalleryImageCard[],
  request: GalleryCatalogRequest,
  totalPages: number,
): ViewerNavigationSession {
  const { page, ...continuationRequest } = request;
  const firstId = items[0]?.id || "";
  return {
    sequence: createViewerSequence(
      firstId,
      items.map((item) => item.id),
    ),
    candidates: [...items],
    catalog: {
      request: continuationRequest,
      previousPage: page > 1 ? page - 1 : undefined,
      nextPage: page < totalPages ? page + 1 : undefined,
    },
  };
}
