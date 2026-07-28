export interface GallerySite {
  name: string;
  title: string;
  description: string;
  searchPlaceholder: string;
  footerTagline: string;
  randomBatchSize: number;
  randomCandidateSize: number;
  homeSections: GalleryHomeSection[];
}

export type GalleryHomeSectionKey =
  | "random"
  | "collections"
  | "latest"
  | "trending";

export interface GalleryHomeSection {
  key: GalleryHomeSectionKey;
  enabled: boolean;
  position: number;
  title: string;
  description: string;
  actionLabel: string;
  itemLimit: number;
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
  primaryCategory: string;
  primaryCategorySlug: string;
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

export interface GalleryRelatedImageReason {
  kind: "primary_category" | "facet" | "tag";
  label: string;
}

export interface GalleryRelatedImage extends GalleryImageCard {
  reasons: GalleryRelatedImageReason[];
}

export interface GalleryDiscovery {
  site: GallerySite;
  seed: string;
  images: GalleryImageCard[];
  categories: GalleryClassificationNode[];
  facets: GalleryFacet[];
}

export interface GallerySubmissionOptions {
  categories: GalleryClassificationNode[];
  facets: GalleryFacet[];
}

export interface GalleryClassificationNode {
  id: string;
  parentId?: string;
  slug: string;
  name: string;
  count: number;
  selected: boolean;
}

export interface GalleryFacet {
  id: string;
  slug: string;
  name: string;
  values: GalleryClassificationNode[];
}

export interface GalleryImagePage {
  items: GalleryImageCard[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  diagnostics: GalleryClassificationDiagnostic[];
  categories: GalleryClassificationNode[];
  facets: GalleryFacet[];
}

export interface GalleryAdminImage extends GalleryImageCard {
  primaryCategoryId: string;
  description: string;
  sourceUrl: string;
  processingState: "queued" | "processing" | "ready" | "failed";
  reviewState: "not_required" | "pending" | "approved" | "rejected";
  publicationState: "draft" | "published" | "hidden" | "deleted";
  safetyState: "pending" | "safe" | "uncertain" | "blocked" | "unavailable";
  publicRenditionReady: boolean;
  facets: GalleryFacetAssignment[];
  tags: Array<{ id: string; name: string }>;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryAdminImagePage {
  items: GalleryAdminImage[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  counts: Record<string, number>;
}

export interface GallerySearchSuggestion {
  key: string;
  label: string;
  context: string;
}

export interface GalleryClassificationDiagnostic {
  code: string;
  path: string[];
  reference?: string;
  params?: Record<string, string>;
}

export type GalleryClassificationStatus =
  "draft" | "active" | "inactive" | "replaced";

export interface GalleryClassificationCatalogNode {
  id: string;
  parentId: string;
  slug: string;
  name: string;
  status: GalleryClassificationStatus;
  editorialPosition?: number;
  replacementId: string;
}

export interface GalleryClassificationCatalogFacet {
  id: string;
  slug: string;
  name: string;
  status: GalleryClassificationStatus;
  editorialPosition?: number;
  replacementId: string;
  values: GalleryClassificationCatalogNode[];
}

export interface GalleryClassificationCatalog {
  revision: number;
  categories: GalleryClassificationCatalogNode[];
  facets: GalleryClassificationCatalogFacet[];
}

export interface GalleryClassificationChildMove {
  childId: string;
  parentId: string;
}

export interface GalleryClassificationGovernanceCommand {
  operation: "set_status" | "reparent" | "merge" | "delete";
  kind: "category" | "facet" | "facet_value" | "tag";
  id: string;
  targetId?: string;
  parentId?: string;
  status?: GalleryClassificationStatus;
  childPlan: GalleryClassificationChildMove[];
  deleteAllRelated: boolean;
}

export interface GalleryClassificationGovernanceStep {
  kind: string;
  identityKind: string;
  sourceId: string;
  targetId: string;
  parentId: string;
  status: string;
  affectedCount: number;
}

export interface GalleryClassificationGovernancePreview {
  catalogRevision: number;
  outcome: "planned" | "rejected";
  diagnostics: GalleryClassificationDiagnostic[];
  plan: {
    expectedCatalogRevision: number;
    expectedRequestToken: string;
    expectedImpactToken: string;
    steps: GalleryClassificationGovernanceStep[];
  };
}

export interface GalleryClassificationTag {
  id: string;
  name: string;
  slug: string;
  status: GalleryClassificationStatus;
  replacementId: string;
  assignmentCount: number;
  aliasCount: number;
}

export interface GalleryClassificationTagPage {
  items: GalleryClassificationTag[];
  nextCursor: string;
}

export interface GalleryClassificationTagProposal {
  id: string;
  submissionId: string;
  inputValue: string;
  lookupKey: string;
  status: "pending" | "approved" | "rejected";
  resolvedTagId: string;
  createdAt?: string;
  reviewedAt?: string;
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
  coverAssetId?: string;
  coverAltText?: string;
  coverWidth?: number;
  coverHeight?: number;
  coverColor?: string;
  seoTitle?: string;
  seoDescription?: string;
  itemCount: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryCollectionDetail extends GalleryCollection {
  images: GalleryImageCard[];
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface GallerySubmission {
  id: string;
  assetId: string;
  imageId: string;
  title: string;
  description: string;
  sourceUrl: string;
  altText: string;
  primaryCategoryId: string;
  processingState: "queued" | "processing" | "ready" | "failed";
  reviewState: "not_required" | "pending" | "approved" | "rejected";
  safetyState: "pending" | "safe" | "uncertain" | "blocked" | "unavailable";
  outcome:
    "pending" | "published" | "duplicate" | "rejected" | "withdrawn" | "failed";
  failureCode: string;
  reviewNote: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryAdminSubmissionPage {
  items: GallerySubmission[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
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
  kind:
    | "report"
    | "source_correction"
    | "safety_uncertain"
    | "near_duplicate"
    | "takedown";
  status: "open" | "reviewing" | "resolved" | "dismissed";
  reason: string;
  description: string;
  proposedSourceUrl: string;
  resolutionNote: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryAdminCasePage {
  items: GalleryCase[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

export interface GalleryAdminOverview {
  pendingSubmissions: number;
  openCases: number;
  publishedImages: number;
  failedProcessing: number;
}

export interface GalleryUploadedAsset {
  id: string;
  filename: string;
  mime: string;
  size: number;
  width?: number;
  height?: number;
  securityState?: string;
  scanStatus?: string;
  scanFailureCode?: string;
}
