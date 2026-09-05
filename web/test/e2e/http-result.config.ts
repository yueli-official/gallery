import { defineConfig } from "@playwright/test";
import base from "./playwright.config";

export default defineConfig({
  ...base,
  testMatch: "http-result.spec.ts",
  outputDir: "../../test-results/http-result-contract",
  reporter: "list",
});
