<template>
  <el-config-provider :locale="elementLocale">
    <div class="app-layout">
      <TitleBar />
      <div class="app-body">
        <Sidebar v-if="!bareLayout" />
        <main
          class="main-content"
          :class="{ 'main-content--bare': bareLayout }"
        >
          <div class="router-outlet-host">
            <!-- デフォルト描画: v-slot + <transition mode="out-in"> は WebView で遅延ルートが
                 真っ白になることがあるため使わない（Vue Router が渡す Component は VNode 扱い）。 -->
            <router-view :key="route.fullPath" />
          </div>
        </main>
      </div>
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import TitleBar from "./components/TitleBar.vue";
import Sidebar from "./components/Sidebar.vue";
import { useSessionUnlock } from "./composables/useSessionUnlock";
import { elementPlusLocaleFor } from "./i18n/elementPlusLocale";

const route = useRoute();
const bareLayout = computed(() => route.meta.bare === true);
const { locale } = useI18n();
const elementLocale = computed(() => elementPlusLocaleFor(locale.value));

const { beginStartupUnlock } = useSessionUnlock();

onMounted(() => {
  // Best-effort: attempt to restore the previous session via the credential blob.
  // The result (unlocked / needs-relogin) is reflected in Go-side IsLoggedIn state
  // which individual views query via App.isLoggedIn().
  beginStartupUnlock().catch(() => undefined);
});
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
   component roots, so the combinator resolves to one element. Gallery manages its own scroll,
   so its root `.gallery-view` is excluded here. */
/* アクティビティは遭遇ログカード内でスクロールするためルートははみ出し抑制 */
.router-outlet-host > *:not(.gallery-view):not(.activity-view) {
  overflow-y: auto;
}
</style>
