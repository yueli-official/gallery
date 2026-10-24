export function useGalleryAdministratorClaim() {
  const { call } = useApi("gallery-authorization");
  const status = useState<{ claimed: boolean; canClaim: boolean } | null>("gallery-administrator-claim", () => null);
  async function refresh() {
    status.value = await call<{ claimed: boolean; canClaim: boolean }>("/setup");
    return status.value;
  }
  return { status, refresh };
}
