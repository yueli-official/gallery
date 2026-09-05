import type { RemoteFailure } from "@yueli/http-runtime";
export function remoteFailure(
  code: string,
  status: number,
  extra: Partial<RemoteFailure> = {},
) {
  return {
    failure: {
      kind: "remote",
      status,
      code,
      params: {},
      violations: [],
      traceId: "test-trace",
      reauth: "not-attempted",
      ...extra,
    } satisfies RemoteFailure,
  };
}
