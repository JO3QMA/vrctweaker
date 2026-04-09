<template>
  <el-config-provider :locale="elLocale">
    <div class="app-layout">
      <TitleBar />
      <div class="app-body">
        <Sidebar v-if="!bareLayout" />
        <main
          class="main-content"
          :class="{ 'main-content--bare': bareLayout }"
        >
          <div class="router-outlet-host">
            <router-view v-slot="{ Component }">
              <transition name="fade" mode="out-in">
                <component :is="Component" />
              </transition>
            </router-view>
          </div>
        </main>
      </div>
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import TitleBar from "./components/TitleBar.vue";
import Sidebar from "./components/Sidebar.vue";
import { useSessionUnlock } from "./composables/useSessionUnlock";
import { elLocale, isAppLocale, setLanguage } from "./i18n";
import { App as WailsApp } from "./wails/app";

const route = useRoute();
const { t, locale } = useI18n();
const bareLayout = computed(() => route.meta.bare === true);

const { beginStartupUnlock } = useSessionUnlock();

async function initLocaleFromBackend(): Promise<void> {
  try {
    let lang = (await WailsApp.getLanguage()).trim();
    if (!lang) {
      const detected = await WailsApp.getSystemLocale();
      const next = isAppLocale(detected) ? detected : "en";
      await WailsApp.setLanguage(next);
      lang = next;
    }
    setLanguage(isAppLocale(lang) ? lang : "en");
  } catch {
    // Match i18n fallbackLocale and invalid-language branches above.
    setLanguage("en");
  }
}

onMounted(() => {
  void initLocaleFromBackend();
  // Best-effort: attempt to restore the previous session via the credential blob.
  // The result (unlocked / needs-relogin) is reflected in Go-side IsLoggedIn state
  // which individual views query via App.isLoggedIn().
  beginStartupUnlock().catch(() => undefined);
});

watch(
  () => [route.meta.titleKey, locale.value] as const,
  () => {
    const key = route.meta.titleKey;
    if (typeof key === "string" && key.length > 0) {
      document.title = `${t(key)} — ${t("app.name")}`;
    } else {
      document.title = t("app.name");
    }
  },
  { immediate: true },
);
</script>

<style scoped>
.app-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
}

.app-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.main-content {
  flex: 1;
  min-height: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 1.5rem;
}

.main-content--bare {
  padding: 1rem 1.25rem;
}

.router-outlet-host {
  flex: 1;
  min-height: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.router-outlet-host > * {
  flex: 1;
  min-height: 0;
  min-width: 0;
}

/* `> *` matches the routed SFC root: Vue 3 applies this parent’s scope attribute to child
   component roots, so the combinator resolves to one element. `<Transition>` adds no wrapper
   DOM node. Gallery manages its own scroll, so its root `.gallery-view` is excluded here. */
/* アクティビティは遭遇ログカード内でスクロールするためルートははみ出し抑制 */
.router-outlet-host > *:not(.gallery-view):not(.activity-view) {
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
