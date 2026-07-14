const siteBrand = process.env.NUXT_PUBLIC_SITE_BRAND || "月离图库";

export default defineNuxtConfig({
  extends: [
    "@platform/auth",
    "@platform/site",
    "@platform/manage",
    "@platform/asset",
  ],
  modules: ["@nuxt/ui"],
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
  runtimeConfig: {
    apiBase: process.env.NUXT_API_BASE || "http://127.0.0.1:8091",
    assetBase: process.env.NUXT_ASSET_BASE || "http://127.0.0.1:8082",
    downstreamBase:
      process.env.NUXT_DOWNSTREAM_BASE || "http://127.0.0.1:8091",
    sealSecret:
      process.env.NUXT_SEAL_SECRET ||
      "dev-gallery-seal-secret-change-me-0123456789ab",
    public: {
      oidcIssuer:
        process.env.NUXT_PUBLIC_OIDC_ISSUER || "http://localhost:8081",
      oidcClientId:
        process.env.NUXT_PUBLIC_OIDC_CLIENT_ID || "gallery-ae-web",
      oidcRedirectUri:
        process.env.NUXT_PUBLIC_OIDC_REDIRECT_URI ||
        "http://localhost:3007/auth/callback",
      oidcScopes:
        process.env.NUXT_PUBLIC_OIDC_SCOPES ||
        "openid profile email roles offline_access",
      accountUrl:
        process.env.NUXT_PUBLIC_ACCOUNT_URL || "http://localhost:3000",
      siteSlug: process.env.NUXT_PUBLIC_SITE_SLUG || "gallery-ae",
      siteBrand,
      siteDomain: process.env.NUXT_PUBLIC_SITE_DOMAIN || "gallery.localhost",
      assetSpace: process.env.NUXT_PUBLIC_ASSET_SPACE || "ae",
      assetNamespace: process.env.NUXT_PUBLIC_ASSET_NAMESPACE || "ae",
      assetProfile:
        process.env.NUXT_PUBLIC_ASSET_PROFILE || "gallery-default",
    },
  },
  devtools: { enabled: true },
});
