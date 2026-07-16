export function useClientHydrated() {
  const hydrated = ref(false);
  onMounted(() => {
    hydrated.value = true;
  });
  return readonly(hydrated);
}
