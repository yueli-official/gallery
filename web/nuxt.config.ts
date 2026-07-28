const siteBrand = process.env.NUXT_PUBLIC_SITE_BRAND || "月离图库";

export default defineNuxtConfig({
  extends: [
    "@yueli/identity-nuxt",
    "@platform/site",
    "@platform/manage",
    "@yueli/asset-nuxt",
  ],
  modules: ["@nuxt/ui", "@yueli/ui"],
  css: ["~/assets/css/main.css"],
  app: {
    head: {
      meta: [
        { property: "og:site_name", content: siteBrand },
        { property: "og:type", content: "website" },
        { name: "twitter:card", content: "summary_large_image" },
      ],
    },
  },
  buildDir: process.env.NUXT_BUILD_DIR || ".nuxt",
  devServer: { port: Number(process.env.NUXT_DEV_PORT || "3007") },
  vite: {
    optimizeDeps: {
      include: ["@platform/ui > vue-picture-cropper"],
    },
  },
  runtimeConfig: {
    apiBase: process.env.NUXT_API_BASE || "http://127.0.0.1:8091",
    assetBase: process.env.NUXT_ASSET_BASE || "http://127.0.0.1:8082",
    downstreamBase: process.env.NUXT_DOWNSTREAM_BASE || "http://127.0.0.1:8091",
    guestSessionTtlSeconds: Number(
      process.env.NUXT_GUEST_SESSION_TTL_SECONDS || 60 * 60 * 24 * 30,
    ),
    guestCookieSecure:
      process.env.NUXT_GUEST_COOKIE_SECURE === "true" ||
      process.env.NODE_ENV === "production",
    assetAudience: "asset-api",
    guestClaimTargets: [
      {
        audience: process.env.NUXT_PUBLIC_OIDC_CLIENT_ID || "gallery-main-web",
        base: process.env.NUXT_DOWNSTREAM_BASE || "http://127.0.0.1:8091",
        path: "/api/v1/gallery/guest-claims",
      },
      {
        audience: "asset-api",
        base: process.env.NUXT_ASSET_BASE || "http://127.0.0.1:8082",
        path: "/api/v1/assets/guest-claims",
      },
    ],
    sealSecret:
      process.env.NUXT_SEAL_SECRET ||
      "dev-gallery-seal-secret-change-me-0123456789ab",
    public: {
      oidcIssuer:
        process.env.NUXT_PUBLIC_OIDC_ISSUER || "http://localhost:8081",
      oidcClientId:
        process.env.NUXT_PUBLIC_OIDC_CLIENT_ID || "gallery-main-web",
      oidcRedirectUri:
        process.env.NUXT_PUBLIC_OIDC_REDIRECT_URI ||
        "http://localhost:3007/auth/callback",
      oidcScopes:
        process.env.NUXT_PUBLIC_OIDC_SCOPES ||
        "openid profile email roles offline_access",
      accountUrl:
        process.env.NUXT_PUBLIC_ACCOUNT_URL || "http://localhost:3000",
      siteSlug: process.env.NUXT_PUBLIC_SITE_SLUG || "gallery-main",
      siteBrand,
      siteDomain: process.env.NUXT_PUBLIC_SITE_DOMAIN || "gallery.localhost",
      assetSpace: process.env.NUXT_PUBLIC_ASSET_SPACE || "yueli",
      assetNamespace: process.env.NUXT_PUBLIC_ASSET_NAMESPACE || "yueli",
      assetProfile: process.env.NUXT_PUBLIC_ASSET_PROFILE || "gallery-default",
    },
  },
  devtools: { enabled: true },
});
