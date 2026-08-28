interface GalleryMeResponse {
  isAdministrator: boolean;
  roles: string[];
  capabilities: string[];
}

const managementCapabilities = [
  "gallery.dashboard.read",
  "gallery.image.read",
  "gallery.image.update",
  "gallery.image.hide",
  "gallery.comment.read",
  "gallery.comment.moderate",
  "gallery.comment.delete",
  "gallery.submission.read",
  "gallery.submission.review",
  "gallery.collection.read",
  "gallery.collection.manage",
  "gallery.classification.read",
  "gallery.classification.proposal_review",
  "gallery.classification.govern",
  "gallery.case.read",
  "gallery.case.resolve",
  "gallery.discovery.read",
  "gallery.asset_settings.manage",
  "authorization.manage",
] as const;

export function useGalleryMe() {
  const { call } = useGalleryApi();
  const me = useState<GalleryMeResponse | null>("gallery-me", () => null);

  async function refreshMe() {
    try {
      me.value = await call<GalleryMeResponse>("/me");
    } catch (error: unknown) {
      const status =
        (error as { failure?: { status?: number } })?.failure?.status ??
        (error as { statusCode?: number })?.statusCode;
      if (status === 401 || status === 403) me.value = null;
    }
  }

  const isAdministrator = computed(() => me.value?.isAdministrator ?? false);
  const can = (capability: string) =>
    me.value?.capabilities.includes(capability) ?? false;
  const canManage = computed(() => managementCapabilities.some(can));

  return {
    me,
    roles: computed(() => me.value?.roles ?? []),
    isAdministrator,
    can,
    canManage,
    refreshMe,
  };
}
