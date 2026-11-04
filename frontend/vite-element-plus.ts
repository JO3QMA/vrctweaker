import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";
import type { PluginOption } from "vite";

const elementPlusResolver = ElementPlusResolver({ importStyle: "css" });

/** Element Plus on-demand (components + function APIs such as ElMessage). */
export function elementPlusPlugins(): PluginOption[] {
  return [
    AutoImport({
      resolvers: [elementPlusResolver],
      dts: "src/auto-imports.d.ts",
    }),
    Components({
      resolvers: [elementPlusResolver],
      dts: "src/components.d.ts",
    }),
  ];
}
