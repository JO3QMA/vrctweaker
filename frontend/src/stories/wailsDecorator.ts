import type { Decorator } from "@storybook/vue3-vite";
import { useRouter } from "vue-router";

/** Loosely typed Wails App method bag (wailsjs models include convertValues). */
type AppStub = Record<string, (...args: never[]) => unknown>;

type WailsDecoratorHooks = {
  created?: () => void;
  beforeUnmount?: () => void;
};

/** Storybook: stub window.go.main.App for the story lifecycle. */
export function withWailsApp(
  app: AppStub,
  hooks?: WailsDecoratorHooks,
): Decorator {
  return (story) => {
    let prevGo: typeof window.go;
    return {
      components: { story },
      template: "<story />",
      created() {
        hooks?.created?.();
        prevGo = window.go;
        window.go = {
          main: {
            App: app as NonNullable<
              NonNullable<typeof window.go>["main"]
            >["App"],
          },
        };
      },
      beforeUnmount() {
        window.go = prevGo;
        hooks?.beforeUnmount?.();
      },
    };
  };
}

/** Storybook: navigate memory router before render. */
export function withRouter(route: {
  name: string;
  query?: Record<string, string>;
}): Decorator {
  return (story) => ({
    components: { story },
    template: "<story />",
    async mounted() {
      await useRouter().push(route);
    },
  });
}
