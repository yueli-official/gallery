export interface ViewerSequenceState {
  ids: string[];
  index: number;
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
