import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";
import type { PluginOption } from "vite";

const elementPlusResolver = ElementPlusResolver({ importStyle: "css" });

export type ElementPlusPluginOptions = {
  /** When false, skip writing src/auto-imports.d.ts and src/components.d.ts (Storybook). */
  emitDts?: boolean;
};

/** Element Plus on-demand (components + optional auto-import for EP function APIs). */
export function elementPlusPlugins(
  options: ElementPlusPluginOptions = {},
): PluginOption[] {
  const emitDts = options.emitDts !== false;
  return [
    // Insurance if a future edit adds EP APIs without explicit imports + function-api-styles.
    AutoImport({
      resolvers: [elementPlusResolver],
      dts: emitDts ? "src/auto-imports.d.ts" : false,
    }),
    Components({
      resolvers: [elementPlusResolver],
      dts: emitDts ? "src/components.d.ts" : false,
    }),
  ];
}
