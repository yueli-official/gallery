import type { ViewerNavigationSession } from "~/utils/viewerSequence";

export function useGalleryViewerNavigation(currentId = "") {
  return useState<ViewerNavigationSession>("gallery-viewer-navigation", () =>
    createViewerNavigationSession(currentId),
  );
}
