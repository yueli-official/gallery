import {
  failureFromProblemResponse,
  readTextWithinLimit,
  type ApiFailure,
  type HttpMethod,
  type QueryValue,
  type RequestOptions,
} from "@yueli/http-runtime";

interface LegacyEnvelope<T> {
  readonly code: string;
  readonly data: T;
  readonly params?: unknown;
  readonly traceId?: unknown;
}

export interface GalleryApiCallOptions {
  readonly method?: string;
  readonly query?: Readonly<Record<string, QueryValue>>;
  readonly body?: unknown;
  readonly headers?: HeadersInit;
  readonly signal?: AbortSignal;
  readonly timeout?: number;
}

function isLegacyEnvelope(value: unknown): value is LegacyEnvelope<unknown> {
  return (
    typeof value === "object" &&
    value !== null &&
    "code" in value &&
    typeof value.code === "string" &&
    "data" in value
  );
}

function problemParams(value: unknown) {
  if (typeof value !== "object" || value === null || Array.isArray(value))
    return {};
  return Object.fromEntries(
    Object.entries(value).filter(([, item]) => {
      if (["string", "number", "boolean"].includes(typeof item)) return true;
      return (
        Array.isArray(item) &&
        item.every((part) =>
          ["string", "number", "boolean"].includes(typeof part),
        )
      );
    }),
  );
}

function safeTraceId(value: unknown, response: Response) {
  const candidate =
    response.headers.get("x-trace-id") ??
    (typeof value === "string" ? value : undefined);
  return candidate && /^[A-Za-z0-9._:-]{1,128}$/.test(candidate)
    ? candidate
    : `legacy-http-${response.status}`;
}

/** Gallery 完成全部 API raw-success 切换后删除 legacy envelope 分支。 */
export async function decodeGalleryLegacyFailure(
  response: Response,
  limit: number,
): Promise<ApiFailure> {
  if (
    response.headers
      .get("content-type")
      ?.split(";", 1)[0]
      ?.trim()
      .toLowerCase() === "application/problem+json"
  ) {
    return failureFromProblemResponse(response, limit);
  }
  const text = await readTextWithinLimit(response, limit);
  if (text === undefined)
    return {
      kind: "protocol",
      code: "foundation.problem.body_too_large",
      reauth: "not-attempted",
    };

  let value: unknown;
  try {
    value = JSON.parse(text);
  } catch {
    return {
      kind: "protocol",
      code: "foundation.problem.invalid_body",
      reauth: "not-attempted",
    };
  }
  if (!isLegacyEnvelope(value) || !/^[a-z][a-z0-9._-]+$/.test(value.code))
    return {
      kind: "protocol",
      code: "foundation.problem.invalid_body",
      reauth: "not-attempted",
    };

  return {
    kind: "remote",
    status: response.status,
    code: value.code,
    params: problemParams(value.params),
    violations: [],
    traceId: safeTraceId(value.traceId, response),
    reauth: "not-attempted",
  };
}

export function decodeGalleryApiResponse<T>(value: unknown): T {
  if (!isLegacyEnvelope(value)) return value as T;
  if (value.code !== "ok")
    throw new Error("foundation.response.legacy_envelope_error");
  return value.data as T;
}

function queryFromUrl(search: string) {
  const query: Record<string, string | string[]> = {};
  for (const [name, value] of new URLSearchParams(search)) {
    const current = query[name];
    if (current === undefined) query[name] = value;
    else if (Array.isArray(current)) query[name] = [...current, value];
    else query[name] = [current, value];
  }
  return query;
}

export function splitGalleryApiUrl(url: string) {
  if (url.includes("#")) throw new Error("foundation.request.invalid_path");
  const queryIndex = url.indexOf("?");
  const path = queryIndex === -1 ? url : url.slice(0, queryIndex);
  const search = queryIndex === -1 ? "" : url.slice(queryIndex + 1);
  return {
    path: path as `/${string}`,
    query: queryFromUrl(search),
  };
}

export function toGalleryApiRequest<T>(
  url: string,
  options: GalleryApiCallOptions = {},
): { path: `/${string}`; options: RequestOptions<T> } {
  const parsed = splitGalleryApiUrl(url);
  const method = String(options.method ?? "GET").toUpperCase() as HttpMethod;
  const headers = options.headers
    ? (Object.fromEntries(
        new Headers(options.headers),
      ) as RequestOptions<T>["headers"])
    : undefined;

  return {
    path: parsed.path,
    options: {
      method,
      query: { ...parsed.query, ...options.query },
      body: options.body as RequestOptions<T>["body"],
      headers,
      signal: options.signal,
      timeoutMs: options.timeout,
      decode: decodeGalleryApiResponse<T>,
      decodeFailure: decodeGalleryLegacyFailure,
    },
  };
}
