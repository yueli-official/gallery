export interface GalleryReauthOptions {
  requireLoggedIn?: boolean;
}

export type GalleryReauthHandler = (
  options?: GalleryReauthOptions,
) => Promise<boolean>;

export function getOptionalGalleryReauth(): GalleryReauthHandler | undefined {
  const nuxtApp = tryUseNuxtApp() as
    | (ReturnType<typeof useNuxtApp> & {
        $platformReauth?: GalleryReauthHandler;
      })
    | null;

  return nuxtApp?.$platformReauth;
}
