export interface GallerySubmissionEditor {
  title: string;
  description: string;
  sourceUrl: string;
  primaryCategoryId: string;
  sceneValueIds: string[];
  tags: string;
  facetValues: Record<string, string[]>;
}
