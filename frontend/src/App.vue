<template>
  <el-config-provider :locale="elLocale">
    <div class="app-layout">
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
    <AppUpdateDialog />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import TitleBar from "./components/TitleBar.vue";
import Sidebar from "./components/Sidebar.vue";
import AppUpdateDialog from "./components/AppUpdateDialog.vue";
import OnboardingWizard from "./components/OnboardingWizard.vue";
import { useFirstLaunchGuide } from "./composables/useFirstLaunchGuide";
import { useSessionUnlock } from "./composables/useSessionUnlock";
import { elLocale, isAppLocale, setLanguage } from "./i18n";
import { App as WailsApp } from "./wails/app";

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

const route = useRoute();
const { t, locale } = useI18n();
const bareLayout = computed(() => route.meta.bare === true);

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

onMounted(async () => {
  void initLocaleFromBackend();
  await loadFirstLaunchGuideState();
  if (firstLaunchGuideCompleted.value && !manualOpen.value) {
    beginStartupUnlock().catch(() => undefined);
  }
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
  padding: var(--space-section);
}

.main-content--bare {
  padding: var(--space-block);
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
   DOM node. Gallery manages its own scroll, so its root `.gallery-view` is excluded here.
   Settings scrolls an inner element so the nav breakpoint can sit on the outer frame.
   Automation scrolls inside the list, editor, and run log. */
/* アクティビティは遭遇ログカード内でスクロールするためルートははみ出し抑制 */
.router-outlet-host
  > *:not(.gallery-view):not(.activity-view):not(.settings-frame):not(
    .automation-view
  ) {
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
