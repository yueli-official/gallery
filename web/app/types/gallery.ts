import type { Facet, FacetValue } from "@platform/facet";

export interface GallerySite {
  name: string;
  title: string;
  description: string;
  searchPlaceholder: string;
  footerTagline: string;
}

export interface GalleryCreator {
  id: string;
  handle: string;
  displayName: string;
}

export interface GalleryArtworkCard {
  id: string;
  title: string;
  description: string;
  coverUrl: string;
  placeholderUrl: string;
  width: number;
  height: number;
  contentRating: "general" | "sensitive" | "adult";
  aiUsage: "none" | "assistive" | "mostly_generated";
  publishedAt?: string;
  creator: GalleryCreator;
}

export type GalleryFacet = Facet;
export type GalleryFacetValue = FacetValue;

export interface GalleryAsset {
  id: string;
  assetId: string;
  sortOrder: number;
  width: number;
  height: number;
  altText: string;
  placeholderUrl: string;
  thumbnailUrl: string;
  cardUrl: string;
  detailUrl: string;
  originalUrl: string;
}

export interface GalleryArtwork extends GalleryArtworkCard {
  license: string;
  rightsBasis: string;
  aiTrainingPermission: string;
  tags: string[];
  assets: GalleryAsset[];
}

export interface GalleryDiscovery {
  site: GallerySite;
  featured: GalleryArtworkCard[];
  latest: GalleryArtworkCard[];
  facets: GalleryFacet[];
  facetValues: GalleryFacetValue[];
}

export interface GalleryCreatorProfile {
  id: string;
  handle: string;
  displayName: string;
  bio: string;
  coverAssetId: string;
  status: "pending" | "active" | "rejected" | "suspended";
  applicationNote: string;
  reviewNote: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface GalleryPublicCreator {
  id: string;
  handle: string;
  displayName: string;
  bio: string;
  coverAssetId: string;
  createdAt?: string;
}

export interface GalleryFacetAssignment {
  facetId: string;
  valueId: string;
}

export interface GalleryStudioArtwork {
  id: string;
  creatorId: string;
  title: string;
  description: string;
  status: "draft" | "pending_review" | "published" | "rejected" | "restricted" | "archived";
  visibility: "public" | "unlisted" | "private";
  contentRating: "general" | "sensitive" | "adult";
  aiUsage: "none" | "assistive" | "mostly_generated";
  aiTrainingPermission: "unspecified" | "allow" | "disallow";
  rightsBasis: "original" | "authorized_repost" | "public_domain" | "licensed_material";
  license: string;
  reviewNote: string;
  createdAt?: string;
  updatedAt?: string;
  publishedAt?: string;
  creator?: GalleryCreatorProfile;
  assets: GalleryAsset[];
  facetAssignments: GalleryFacetAssignment[];
}

export interface GalleryUploadedAsset {
  id: string;
  filename: string;
  mime: string;
  size: number;
  width?: number;
  height?: number;
}
