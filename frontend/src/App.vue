<template>
  <el-config-provider
    :locale="elLocale"
    size="default"
    :z-index="elementPlusZIndex"
  >
    <AppStartupScreen
      v-if="!startupReady"
      :status="startupStatus"
      :retrying="startupRetrying"
      @retry="onStartupRetry"
      @open-data-folder="handleOpenDataFolder"
    />
    <div v-else class="app-layout">
      <TitleBar />
      <OnboardingWizard
        v-if="showOnboardingWizard"
        @finished="onOnboardingFinished"
      />
      <div class="app-body">
        <Sidebar v-if="!bareLayout" />
        <main
          class="main-content"
          :class="{ 'main-content--bare': bareLayout }"
        >
          <div
            class="router-outlet-host"
            :class="{ 'router-outlet-host--self-scroll': routeSelfScroll }"
          >
            <router-view v-slot="{ Component }">
              <transition name="fade" mode="out-in">
                <component :is="Component" />
              </transition>
            </router-view>
          </div>
        </main>
      </div>
    </div>
    <AppUpdateDialog v-if="startupReady" />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import TitleBar from "./components/TitleBar.vue";
import Sidebar from "./components/Sidebar.vue";
import AppUpdateDialog from "./components/AppUpdateDialog.vue";
import OnboardingWizard from "./components/OnboardingWizard.vue";
import AppStartupScreen from "./components/AppStartupScreen.vue";
import { useStartupInit } from "./composables/useStartupInit";
import { useFirstLaunchGuide } from "./composables/useFirstLaunchGuide";
import { useSessionUnlock } from "./composables/useSessionUnlock";
import { Z_INDEX_ELEMENT_PLUS_BASE } from "./constants/elementPlusStack";
import { elLocale, isAppLocale, setLanguage } from "./i18n";

const elementPlusZIndex = Z_INDEX_ELEMENT_PLUS_BASE;
import { App as WailsApp } from "./wails/app";
import { showToast } from "./utils/showToast";

const {
  status: startupStatus,
  retrying: startupRetrying,
  retryInit: onStartupRetry,
  openDataFolder: onOpenDataFolder,
} = useStartupInit();

const startupReady = computed(() => startupStatus.value.ready);

async function handleOpenDataFolder(): Promise<void> {
  try {
    await onOpenDataFolder();
  } catch {
    showToast.error(t("app.startup.openDataFolderFailed"));
  }
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

const route = useRoute();
const { t, locale } = useI18n();
const bareLayout = computed(() => route.meta.bare === true);

const routeSelfScroll = computed(() =>
  route.matched.some((record) => record.meta.selfScroll === true),
);

const { beginStartupUnlock } = useSessionUnlock();
const { manualOpen, clearManualFirstLaunchRequest } = useFirstLaunchGuide();
const firstLaunchGuideCompleted = ref(true);
const showOnboardingWizard = computed(
  () => !firstLaunchGuideCompleted.value || manualOpen.value,
);

async function loadFirstLaunchGuideState(): Promise<void> {
  const maxAttempts = 12;
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      firstLaunchGuideCompleted.value =
        await WailsApp.getFirstLaunchGuideCompleted();
      return;
    } catch {
      if (attempt < maxAttempts - 1) {
        await delay(50 * (attempt + 1));
        continue;
      }
      // Backend still unavailable: show wizard so first launch is not skipped.
      firstLaunchGuideCompleted.value = false;
      return;
    }
  }
}

function onOnboardingFinished(): void {
  firstLaunchGuideCompleted.value = true;
  clearManualFirstLaunchRequest();
  beginStartupUnlock().catch(() => undefined);
}

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

watch(
  startupReady,
  (ready) => {
    if (!ready) {
      return;
    }
    void initLocaleFromBackend();
    void loadFirstLaunchGuideState().then(() => {
      if (firstLaunchGuideCompleted.value && !manualOpen.value) {
        beginStartupUnlock().catch(() => undefined);
      }
    });
  },
  { immediate: true },
);

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
  background: var(--color-bg-base);
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

/* Routed view root scrolls unless the route sets meta.selfScroll (inner scroller). */
.router-outlet-host:not(.router-outlet-host--self-scroll) > * {
  overflow-x: hidden;
  overflow-y: auto;
  padding: var(--space-section);
  box-sizing: border-box;
  scrollbar-gutter: stable;
}

.main-content--bare
  .router-outlet-host:not(.router-outlet-host--self-scroll)
  > * {
  padding: var(--space-block);
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
