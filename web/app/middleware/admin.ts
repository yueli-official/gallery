const routeCapabilities: Array<{
  matches: (path: string) => boolean;
  capabilities: string[];
}> = [
  {
    matches: (path) => path === "/manage/authorization",
    capabilities: ["authorization.manage"],
  },
  {
    matches: (path) => path === "/manage/assets",
    capabilities: ["gallery.asset_settings.manage"],
  },
  {
    matches: (path) => path === "/manage/discovery",
    capabilities: ["gallery.discovery.read", "gallery.discovery.manage"],
  },
  {
    matches: (path) => path === "/manage/cases",
    capabilities: ["gallery.case.read", "gallery.case.resolve"],
  },
  {
    matches: (path) => path === "/manage/classification",
    capabilities: [
      "gallery.classification.read",
      "gallery.classification.proposal_review",
      "gallery.classification.govern",
    ],
  },
  {
    matches: (path) => path.startsWith("/manage/collections"),
    capabilities: ["gallery.collection.read", "gallery.collection.manage"],
  },
  {
    matches: (path) => path === "/manage/comments",
    capabilities: [
      "gallery.comment.read",
      "gallery.comment.moderate",
      "gallery.comment.delete",
      "authorization.manage",
    ],
  },
  {
    matches: (path) => path === "/manage/submissions",
    capabilities: ["gallery.submission.read", "gallery.submission.review"],
  },
  {
    matches: (path) => path === "/manage/images",
    capabilities: [
      "gallery.image.read",
      "gallery.image.update",
      "gallery.image.hide",
    ],
  },
  {
    matches: (path) => path === "/manage",
    capabilities: ["gallery.dashboard.read"],
  },
];

const managementLandingPages = [
  { path: "/manage", capabilities: ["gallery.dashboard.read"] },
  {
    path: "/manage/images",
    capabilities: [
      "gallery.image.read",
      "gallery.image.update",
      "gallery.image.hide",
    ],
  },
  {
    path: "/manage/comments",
    capabilities: [
      "gallery.comment.read",
      "gallery.comment.moderate",
      "gallery.comment.delete",
      "authorization.manage",
    ],
  },
  {
    path: "/manage/submissions",
    capabilities: ["gallery.submission.read", "gallery.submission.review"],
  },
  {
    path: "/manage/collections",
    capabilities: ["gallery.collection.read", "gallery.collection.manage"],
  },
  {
    path: "/manage/classification",
    capabilities: [
      "gallery.classification.read",
      "gallery.classification.proposal_review",
      "gallery.classification.govern",
    ],
  },
  {
    path: "/manage/cases",
    capabilities: ["gallery.case.read", "gallery.case.resolve"],
  },
  {
    path: "/manage/discovery",
    capabilities: ["gallery.discovery.read", "gallery.discovery.manage"],
  },
  {
    path: "/manage/assets",
    capabilities: ["gallery.asset_settings.manage"],
  },
  { path: "/manage/authorization", capabilities: ["authorization.manage"] },
] as const;

export default defineNuxtRouteMiddleware(async (to) => {
  const { can, canManage, refreshMe } = useGalleryMe();
  await refreshMe();

  const requirement = routeCapabilities.find((item) => item.matches(to.path));
  if (requirement?.capabilities.some(can)) return;

  if (!canManage.value) return navigateTo("/");
  const landing = managementLandingPages.find((item) =>
    item.capabilities.some(can),
  );
  if (landing && landing.path !== to.path) return navigateTo(landing.path);
  return navigateTo("/");
});
