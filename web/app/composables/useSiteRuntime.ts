export function useSiteRuntime() {
  const config = useRuntimeConfig();
  return {
    slug: computed(() => config.public.siteSlug),
    brand: computed(() => config.public.siteBrand),
  };
}
