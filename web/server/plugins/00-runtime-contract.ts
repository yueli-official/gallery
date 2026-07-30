import { validateRuntimeContract } from "../utils/runtimeContract";

export default defineNitroPlugin(() => {
  validateRuntimeContract(useRuntimeConfig());
});
