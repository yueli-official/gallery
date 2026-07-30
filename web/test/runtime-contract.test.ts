import { describe, expect, it } from "vitest";
import { validateRuntimeContract } from "../server/utils/runtimeContract";

function validConfig(overrides: Record<string, unknown> = {}) {
  return {
    apiBase: "http://gallery-api:8091",
    assetBase: "http://asset:8082",
    downstreamBase: "http://gallery-api:8091",
    sealSecret: "a".repeat(32),
    cookieSecure: false,
    authCookieSecure: false,
    guestCookieSecure: false,
    public: {
      oidcIssuer: "http://localhost:8081",
      oidcRedirectUri: "http://localhost:3007/auth/callback",
      oidcPostLogoutRedirectUri: "http://localhost:3007/",
      accountUrl: "http://localhost:3000",
    },
    ...overrides,
  };
}

describe("Gallery runtime contract", () => {
  it("validates a read-only runtime config without mutating it", () => {
    const config = validConfig({ cookieSecure: "false" });
    Object.freeze(config);
    expect(() => validateRuntimeContract(config)).not.toThrow();
    expect(config.authCookieSecure).toBe(false);
    expect(config.guestCookieSecure).toBe(false);
  });

  it("rejects the old fixed development secret", () => {
    expect(() => validateRuntimeContract(validConfig({ sealSecret: "" })))
      .toThrow(/at least 32 bytes/);
  });

  it("rejects inconsistent cookie security settings", () => {
    expect(() => validateRuntimeContract(validConfig({ authCookieSecure: true })))
      .toThrow(/same value/);
  });

  it("requires secure cookies for HTTPS callbacks", () => {
    const config = validConfig();
    config.public.oidcRedirectUri = "https://gallery.example.com/auth/callback";
    expect(() => validateRuntimeContract(config)).toThrow(/COOKIE_SECURE/);
  });
});
