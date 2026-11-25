import { createUiPreset } from "@yueli/ui/theme";

export default defineAppConfig(createUiPreset({ primary: "amber", neutral: "zinc" }, { icons: { dark: "i-tabler-moon-stars", light: "i-tabler-sun-high" } }));
