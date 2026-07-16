export interface GallerySubmissionMetadata {
  title: string;
  description: string;
  sourceUrl: string;
  primaryCategoryId: string;
  sceneValueIds: string[];
  tags: string;
}

export interface GallerySubmissionBatchItem extends GallerySubmissionMetadata {
  status: string;
  customized: boolean;
}

export function applyGallerySubmissionDefaults(
  item: GallerySubmissionBatchItem,
  defaults: GallerySubmissionMetadata,
  preserveTitle = false,
) {
  if (item.status === "completed") return false;
  if (!preserveTitle && defaults.title.trim())
    item.title = defaults.title.trim();
  item.description = defaults.description;
  item.sourceUrl = defaults.sourceUrl;
  item.primaryCategoryId = defaults.primaryCategoryId;
  item.sceneValueIds = [...defaults.sceneValueIds];
  item.tags = defaults.tags;
  item.customized = false;
  return true;
}

export function gallerySubmissionMetadataValid(
  item: GallerySubmissionMetadata,
) {
  return Boolean(
    item.title.trim() && item.primaryCategoryId && item.sceneValueIds.length,
  );
}
