import { createApp, watch } from "vue";
import { createRouter, createWebHashHistory } from "vue-router";
import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import AppRoot from "./App.vue";
import { getInitialUILanguageCode } from "./bootstrap/initialUiLanguage";
import { syncDocumentTitle } from "./bootstrap/syncDocumentTitle";
import { createAppI18n } from "./i18n";
import { appRoutes } from "./router/routes";
import "./assets/style.css";

async function bootstrap() {
  const code = await getInitialUILanguageCode();
  const i18n = createAppI18n(code);
  const router = createRouter({
    history: createWebHashHistory(),
    routes: appRoutes,
  });

  router.afterEach((to) => {
    syncDocumentTitle((key) => i18n.global.t(key), to.meta);
  });

  const app = createApp(AppRoot);
  app.use(i18n);
  app.use(router);
  watch(i18n.global.locale, () => {
    syncDocumentTitle(
      (key) => i18n.global.t(key),
      router.currentRoute.value.meta,
    );
  });
  app.use(ElementPlus);
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component);
  }
  app.mount("#app");
}

void bootstrap().catch((e: unknown) => {
  console.error("[VRChat Tweaker] bootstrap failed", e);
  const root = document.getElementById("app");
  const msg = e instanceof Error ? e.message : String(e);
  if (root) {
    root.textContent = `起動に失敗しました: ${msg}`;
  }
});
