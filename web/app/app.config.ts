import { createUiPreset } from "@yueli/ui/theme";

const preset = createUiPreset({ primary: "blue" });

export default defineAppConfig({
  ui: {
    ...preset.ui,
    toast: {
      slots: {
        root: "gallery-toast gap-2.5 rounded-xl p-3 ring-0",
        wrapper: "flex min-h-5 min-w-0 flex-1 flex-col justify-center",
        title:
          "line-clamp-2 text-[0.8125rem] font-semibold leading-5 text-highlighted",
        description:
          "mt-0.5 line-clamp-3 text-xs leading-5 text-muted",
        icon: "size-[1.125rem] shrink-0 self-center",
        close:
          "self-center rounded-md p-1 text-dimmed transition-colors hover:bg-elevated hover:text-highlighted",
        progress: "hidden",
      },
    },
    input: {
      slots: {
        root: "relative flex w-full items-center",
        base: "focus-visible:outline-none focus-visible:ring-inset",
      },
    },
    textarea: {
      slots: {
        root: "relative flex w-full items-center",
        base: "focus-visible:outline-none focus-visible:ring-inset",
      },
    },
    select: {
      slots: {
        base: "flex w-full focus-visible:outline-none focus-visible:ring-inset",
      },
    },
    selectMenu: {
      slots: {
        base: "flex w-full focus-visible:outline-none focus-visible:ring-inset",
      },
    },
  },
});
