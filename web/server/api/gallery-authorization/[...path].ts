import { createBffHandler } from "@yueli/nuxt-runtime/server";

const galleryAuthorizationBff = createBffHandler({
  mountPath: "/api/gallery-authorization",
  resolveTarget({ event }) {
    const configured = new URL(String(useRuntimeConfig(event).apiBase));
    if (
      (configured.protocol !== "http:" && configured.protocol !== "https:") ||
      configured.username ||
      configured.password ||
      configured.search ||
      configured.hash
    ) {
      throw new TypeError("Invalid Gallery authorization target");
    }
    const basePath =
      configured.pathname === "/"
        ? ""
        : configured.pathname.replace(/\/$/u, "");
    return {
      origin: configured.origin,
      pathPrefix: `${basePath}/api/v1/authorization`,
    };
  },
  credential: {
    async resolve({ event }) {
      const headers = await sessionAuthHeaders(event);
      const authorization = headers.authorization;
      return authorization?.startsWith("Bearer ")
        ? { kind: "bearer", token: authorization.slice("Bearer ".length) }
        : { kind: "anonymous" };
    },
  },
});

export default withSafeBffErrors(galleryAuthorizationBff);
