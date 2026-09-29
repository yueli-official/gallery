import { realpathSync } from "node:fs";
import { resolve, sep } from "node:path";

const siteBrand = process.env.NUXT_PUBLIC_SITE_BRAND || "月离图库";
const cookieSecure =
  process.env.NUXT_COOKIE_SECURE === undefined
    ? process.env.NODE_ENV === "production"
    : process.env.NUXT_COOKIE_SECURE === "true";
const resolvedNuxt = realpathSync(resolve(process.cwd(), "node_modules/nuxt"));
const pnpmStoreMarker = `${sep}.pnpm${sep}`;
const pnpmStoreIndex = resolvedNuxt.indexOf(pnpmStoreMarker);
const dependencyRoot =
  pnpmStoreIndex >= 0
    ? resolvedNuxt.slice(0, pnpmStoreIndex)
    : resolve(process.cwd(), "node_modules");

export default defineNuxtConfig({
  extends: ["@yueli/identity-nuxt", "@yueli/asset-nuxt"],
  modules: ["@nuxt/ui", "@yueli/ui", "@yueli/nuxt-runtime"],
  yueliUi: {
    tablerIcons: [
      "i-tabler-dots-vertical",
      "i-tabler-arrow-bar-to-down",
      "i-tabler-arrow-bar-to-up",
      "i-tabler-arrow-down",
      "i-tabler-arrow-up",
      "i-tabler-file-text",
      "i-tabler-grip-vertical",
      "i-tabler-layout-grid",
      "i-tabler-list",
      "i-tabler-photo-cog",
      "i-tabler-search",
    ],
  },
  icon: {
    provider: "none",
    fallbackToApi: false,
    serverBundle: { collections: ["tabler"] },
    clientBundle: {
      scan: {
        globInclude: [
          "app/**/*.{vue,js,mjs,ts,jsx,tsx}",
          "node_modules/@yueli/**/*.{vue,js,mjs,ts,jsx,tsx}",
        ],
        globExclude: [
          "test/**",
          "tests/**",
          "coverage/**",
          "dist/**",
          ".nuxt/**",
          ".output/**",
          ".*",
        ],
      },
      sizeLimitKb: 256,
    },
  },
  yueliRuntime: {
    defaultTarget: "platform",
    targets: {
      platform: {
        path: "/",
        ssr: {
          cookies: [],
          headers: ["accept-language", "user-agent"],
        },
      },
      asset: {
        path: "/asset-api",
        ssr: {
          cookies: [],
          headers: ["accept-language", "user-agent"],
        },
      },
      gallery: {
        path: "/api/gallery",
        ssr: {
          cookies: [],
          headers: ["accept-language", "user-agent"],
        },
      },
      "gallery-authorization": {
        path: "/api/gallery-authorization",
        ssr: {
          cookies: [],
          headers: ["accept-language", "user-agent"],
        },
      },
      identity: {
        path: "/identity-api",
        ssr: {
          cookies: [],
          headers: ["accept-language", "user-agent"],
        },
      },
    },
  },
  css: ["~/assets/css/main.css"],
  app: {
    head: {
      link: [
        { rel: "icon", type: "image/x-icon", href: "/favicon.ico?v=20260929" },
        { rel: "icon", type: "image/svg+xml", sizes: "any", href: "/favicon.svg?v=20260929" },
        { rel: "apple-touch-icon", sizes: "180x180", href: "/apple-touch-icon.png?v=20260929" },
      ],
      htmlAttrs: { lang: "zh-CN" },
      meta: [
        { property: "og:site_name", content: siteBrand },
        { property: "og:type", content: "website" },
        { name: "twitter:card", content: "summary_large_image" },
      ],
    },
  },
  buildDir: process.env.NUXT_BUILD_DIR || ".nuxt",
  devServer: {
    host: "127.0.0.1",
    port: Number(process.env.NUXT_DEV_PORT || "3007"),
  },
  fonts: {
    providers: {
      google: false,
      googleicons: false,
      bunny: false,
      fontshare: false,
      fontsource: false,
    },
  },
  nitro: {
    esbuild: {
      options: {
        // 当前 @yueli 正式包仍包含供 Nuxt/Nitro 消费的 TypeScript 源码。
        // 只放行该作用域，其余 node_modules 继续保持 Nitro 默认的外部依赖处理。
        exclude: /node_modules(?!.*(?:@yueli\+|@yueli[\\/]))/,
      },
    },
  },
  runtimeConfig: {
    apiBase: process.env.NUXT_API_BASE || "http://127.0.0.1:8091",
    assetBase: process.env.NUXT_ASSET_BASE || "http://127.0.0.1:8082",
    identityBase: process.env.NUXT_IDENTITY_BASE || "http://127.0.0.1:8081",
    downstreamBase: process.env.NUXT_DOWNSTREAM_BASE || "http://127.0.0.1:8091",
    guestSessionTtlSeconds: Number(
      process.env.NUXT_GUEST_SESSION_TTL_SECONDS || 60 * 60 * 24 * 30,
    ),
    cookieSecure,
    guestCookieSecure: cookieSecure,
    authCookieSecure: cookieSecure,
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
    sealSecret: process.env.NUXT_SEAL_SECRET || "",
    public: {
      oidcIssuer:
        process.env.NUXT_PUBLIC_OIDC_ISSUER || "http://localhost:8081",
      oidcClientId:
        process.env.NUXT_PUBLIC_OIDC_CLIENT_ID || "gallery-main-web",
      oidcRedirectUri:
        process.env.NUXT_PUBLIC_OIDC_REDIRECT_URI ||
        "http://localhost:3007/auth/callback",
      oidcPostLogoutRedirectUri:
        process.env.NUXT_PUBLIC_OIDC_POST_LOGOUT_REDIRECT_URI ||
        "http://localhost:3007/",
      oidcScopes:
        process.env.NUXT_PUBLIC_OIDC_SCOPES ||
        "openid profile email roles offline_access",
      accountUrl:
        process.env.NUXT_PUBLIC_ACCOUNT_URL || "http://localhost:3000",
      siteSlug: process.env.NUXT_PUBLIC_SITE_SLUG || "gallery-main",
      assetNamespace: process.env.NUXT_PUBLIC_ASSET_NAMESPACE || "gallery",
      siteBrand,
    },
  },
  vite: {
    resolve: {
      dedupe: ["vue", "vue-router", "@vue/runtime-core", "@vue/runtime-dom"],
    },
    server: {
      fs: {
        allow: [resolve(process.cwd(), "../../foundation"), dependencyRoot],
      },
    },
  },
  devtools: { enabled: process.env.NUXT_DEVTOOLS === "true" },
});
