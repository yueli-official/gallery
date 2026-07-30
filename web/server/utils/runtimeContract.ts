interface GalleryRuntimeConfig {
  apiBase?: unknown;
  assetBase?: unknown;
  downstreamBase?: unknown;
  sealSecret?: unknown;
  cookieSecure?: unknown;
  authCookieSecure?: unknown;
  guestCookieSecure?: unknown;
  public?: Record<string, unknown>;
}

export function validateRuntimeContract(config: Readonly<GalleryRuntimeConfig>) {
  const sealSecret = String(config.sealSecret || "").trim();
  if (Buffer.byteLength(sealSecret, "utf8") < 32) {
    throw new Error("NUXT_SEAL_SECRET must contain at least 32 bytes");
  }

  const cookieSecure = runtimeBoolean(config.cookieSecure, "NUXT_COOKIE_SECURE");
  const authCookieSecure = runtimeBoolean(
    config.authCookieSecure,
    "NUXT_AUTH_COOKIE_SECURE",
  );
  const guestCookieSecure = runtimeBoolean(
    config.guestCookieSecure,
    "NUXT_GUEST_COOKIE_SECURE",
  );
  if (authCookieSecure !== cookieSecure || guestCookieSecure !== cookieSecure) {
    throw new Error("Gallery cookie security settings must use the same value");
  }

  for (const [name, value] of [
    ["NUXT_API_BASE", config.apiBase],
    ["NUXT_ASSET_BASE", config.assetBase],
    ["NUXT_DOWNSTREAM_BASE", config.downstreamBase],
    ["NUXT_PUBLIC_OIDC_ISSUER", config.public?.oidcIssuer],
    ["NUXT_PUBLIC_OIDC_REDIRECT_URI", config.public?.oidcRedirectUri],
    ["NUXT_PUBLIC_OIDC_POST_LOGOUT_REDIRECT_URI", config.public?.oidcPostLogoutRedirectUri],
    ["NUXT_PUBLIC_ACCOUNT_URL", config.public?.accountUrl],
  ] as const) {
    requireAbsoluteHTTPURL(name, value);
  }
  const redirect = new URL(String(config.public?.oidcRedirectUri));
  if (redirect.protocol === "https:" && !cookieSecure) {
    throw new Error("NUXT_COOKIE_SECURE must be true for an HTTPS OIDC redirect URI");
  }
}

function runtimeBoolean(value: unknown, name: string) {
  if (typeof value === "boolean") return value;
  if (value === "true") return true;
  if (value === "false") return false;
  throw new Error(`${name} must be true or false`);
}

function requireAbsoluteHTTPURL(name: string, value: unknown) {
  let parsed: URL;
  try {
    parsed = new URL(String(value || ""));
  }
  catch {
    throw new Error(`${name} must be an absolute HTTP(S) URL`);
  }
  if ((parsed.protocol !== "http:" && parsed.protocol !== "https:") || parsed.username || parsed.password) {
    throw new Error(`${name} must be an absolute HTTP(S) URL without credentials`);
  }
}
