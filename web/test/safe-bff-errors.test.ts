import { describe, expect, it } from "vitest";
import { sanitizeBffError } from "../server/utils/bffError";
describe("Gallery BFF Problem boundary", () => {
  it("keeps supported HTTP status and emits a safe traceable Problem", () => {
    const result = sanitizeBffError(
      {
        statusCode: 400,
        statusMessage: "private text",
        stack: "E:/private",
        url: "http://127.0.0.1",
      },
      "test-trace",
    );
    expect(result).toEqual({
      type: "https://errors.yueli.dev/problems/gallery.request.invalid",
      status: 400,
      code: "gallery.request.invalid",
      traceId: "test-trace",
    });
  });
  it("reduces unknown server failures to internal and distinguishes timeout", () => {
    const result = sanitizeBffError(
      { statusCode: 500, statusMessage: "database password=secret" },
      "trace",
    );
    expect(result).toMatchObject({
      status: 500,
      code: "common.internal",
      traceId: "trace",
    });
    expect(JSON.stringify(result)).not.toContain("secret");
    expect(sanitizeBffError({ statusCode: 504 }, "timeout")).toMatchObject({
      status: 504,
      code: "gallery.gateway.timeout",
    });
  });
});
