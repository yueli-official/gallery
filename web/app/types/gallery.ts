import type { Facet, FacetValue } from "@platform/facet";

export interface GallerySite {
  name: string;
  title: string;
  description: string;
  searchPlaceholder: string;
  footerTagline: string;
  randomBatchSize: number;
  randomCandidateSize: number;
}

export interface GalleryMetrics {
  views: number;
  favorites: number;
}

export interface GalleryImageCard {
  id: string;
  assetId: string;
  title: string;
  altText: string;
  width: number;
  height: number;
  dominantColor: string;
  topic: string;
  topicSlug: string;
  publishedAt?: string;
  metrics: GalleryMetrics;
}

export interface GalleryFacetAssignment {
  facetId: string;
  valueId: string;
}

export interface GalleryImage extends GalleryImageCard {
  description: string;
  sourceUrl: string;
  focusX: number;
  focusY: number;
  tags: string[];
  facets: GalleryFacetAssignment[];
  favorited: boolean;
}

export interface GalleryDiscovery {
  site: GallerySite;
  seed: string;
  images: GalleryImageCard[];
  facets: Facet[];
  facetValues: FacetValue[];
}

export interface GalleryImagePage {
  items: GalleryImageCard[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

export interface GalleryCollection {
  id: string;
  kind: "gallery.editorial" | "gallery.favorites";
  resourceKind: "gallery.image";
  ownerKind: "site" | "user";
  visibility: "private" | "public";
  name: string;
  description: string;
  version: number;
  slug?: string;
  coverImageId?: string;
  itemCount: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryCollectionDetail extends GalleryCollection {
  images: GalleryImageCard[];
}

export interface GallerySubmission {
  id: string;
  assetId: string;
  imageId: string;
  title: string;
  description: string;
  sourceUrl: string;
  altText: string;
  topicId: string;
  processingState: "queued" | "processing" | "ready" | "failed";
  reviewState: "not_required" | "pending" | "approved" | "rejected";
  safetyState: "pending" | "safe" | "uncertain" | "blocked" | "unavailable";
  outcome: "pending" | "published" | "duplicate" | "rejected" | "withdrawn" | "failed";
  failureCode: string;
  reviewNote: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryRanking {
  kind: "trending" | "most_viewed" | "most_favorited";
  window: "24h" | "7d" | "30d" | "all";
  generatedAt?: string;
  images: GalleryImageCard[];
}

export interface GalleryCase {
  id: string;
  imageId: string;
  submissionId: string;
  kind: string;
  status: string;
  reason: string;
  description: string;
  proposedSourceUrl: string;
  resolutionNote: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryAdminOverview {
  pendingSubmissions: number;
  openCases: number;
  publishedImages: number;
  failedProcessing: number;
}

export type GalleryFacet = Facet;
export type GalleryFacetValue = FacetValue;

export interface GalleryUploadedAsset {
  id: string;
  filename: string;
  mime: string;
  size: number;
  width?: number;
  height?: number;
}
