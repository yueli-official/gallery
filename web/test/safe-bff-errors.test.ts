import { describe, expect, it } from "vitest";
import { sanitizeBffError } from "../server/utils/bffError";

describe("Gallery BFF error boundary", () => {
  it("preserves a safe client status without returning stack or request data", () => {
    const result = sanitizeBffError({
      statusCode: 400,
      statusMessage: "Invalid BFF request path",
      url: "http://127.0.0.1/private",
      stack: "E:\\private\\gallery\\node_modules\\runtime.js",
    });

    expect(result).toEqual({
      error: true,
      statusCode: 400,
      statusMessage: "Invalid BFF request path",
      message: "Invalid BFF request path",
    });
    expect(JSON.stringify(result)).not.toContain("127.0.0.1");
    expect(JSON.stringify(result)).not.toContain("node_modules");
  });

  it("replaces server failures with a fixed public message", () => {
    expect(
      sanitizeBffError({
        statusCode: 500,
        statusMessage: "database password is secret",
      }),
    ).toEqual({
      error: true,
      statusCode: 500,
      statusMessage: "BFF request failed",
      message: "BFF request failed",
    });
  });
});
