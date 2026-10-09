// Product card density and badge placement; collection state stays in Foundation.
export const compactCollectionClasses = [
  "gallery-compact-collection",
  "[&_[data-collection-grid]]:grid-cols-[repeat(auto-fill,minmax(min(100%,12rem),1fr))]",
  "[&_[data-collection-grid]]:gap-3",
  "[@container_collection_(max-width:30rem)]:[&_[data-collection-grid]]:grid-cols-[repeat(auto-fill,minmax(min(100%,9rem),1fr))]",
  "[@container_collection_(max-width:30rem)]:[&_[data-collection-grid]]:gap-2",
  "[&_[data-collection-grid]>article>[data-slot=root]]:p-0",
  "[&_[data-collection-grid]>article>[data-slot=root]]:bg-transparent",
  "[&_[data-collection-grid]>article>[data-slot=root]]:shadow-none",
  "[&_[data-collection-grid]>article>[data-slot=root]]:[backdrop-filter:none]",
  "[&_[data-collection-grid]>article>[data-slot=root]]:rounded-none",
  "[&_[data-collection-grid]>article>[data-slot=root]]:left-2.5",
  "[&_[data-collection-grid]>article>[data-slot=root]]:top-2.5",
].join(" ");

export const imageManagePageClasses = [
  "[&_[data-manage-page-actions]]:w-full",
  "[&_[data-manage-page-actions]]:min-w-0",
  "max-xl:[&_[data-manage-page-header]]:grid",
  "max-xl:[&_[data-manage-page-header]]:grid-cols-[minmax(0,1fr)_auto]",
  "max-xl:[&_[data-manage-page-header]]:items-center",
  "max-xl:[&_[data-manage-page-actions]]:contents",
  "xl:[&_[data-manage-page-actions]]:w-auto",
  "xl:[&_[data-manage-page-actions]]:max-w-[calc(100%_-_8rem)]",
].join(" ");
