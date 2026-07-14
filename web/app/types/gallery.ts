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

export interface GalleryCategory {
  id: string;
  slug: string;
  name: string;
  description: string;
  artworkCount: number;
}

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
  categories: GalleryCategory[];
}

