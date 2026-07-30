import { createBffHandler } from "@yueli/nuxt-runtime/server";

const safeMethods = new Set(["GET", "HEAD", "OPTIONS"]);
const guestScopedPath = "/api/gallery/me/";

const galleryBff = createBffHandler({
  mountPath: "/api/gallery",
  resolveTarget({ event }) {
    const configured = new URL(String(useRuntimeConfig(event).apiBase));
    if (
      (configured.protocol !== "http:" && configured.protocol !== "https:") ||
      configured.username ||
      configured.password ||
      configured.search ||
      configured.hash
    ) {
      throw new TypeError("Invalid Gallery API target");
    }
    const basePath = configured.pathname === "/"
      ? ""
      : configured.pathname.replace(/\/$/u, "");
    return {
      origin: configured.origin,
      pathPrefix: `${basePath}/api/v1/gallery`,
    };
  },
  credential: {
    async resolve({ event }) {
      const config = useRuntimeConfig(event);
      let headers = await sessionAuthHeaders(event);
      if (!headers.authorization) {
        const clientId = String(config.public.oidcClientId || "");
        const createGuest =
          !safeMethods.has(event.method) ||
          getRequestURL(event).pathname.startsWith(guestScopedPath);
        headers = await guestSessionAuthHeaders(event, clientId, createGuest);
      }
      const authorization = headers.authorization;
      return authorization?.startsWith("Bearer ")
        ? { kind: "bearer", token: authorization.slice("Bearer ".length) }
        : { kind: "anonymous" };
    },
  },
});

export default withSafeBffErrors(galleryBff);
