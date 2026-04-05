<template>
  <div class="app-layout">
    <TitleBar />
    <div class="app-body">
      <Sidebar v-if="!bareLayout" />
      <main class="main-content" :class="{ 'main-content--bare': bareLayout }">
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
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import TitleBar from "./components/TitleBar.vue";
import Sidebar from "./components/Sidebar.vue";
import { useSessionUnlock } from "./composables/useSessionUnlock";

const route = useRoute();
const bareLayout = computed(() => route.meta.bare === true);

const { tryUnlockOnStartup } = useSessionUnlock();

onMounted(() => {
  // Best-effort: attempt to restore the previous session via the credential blob.
  // The result (unlocked / needs-relogin) is reflected in Go-side IsLoggedIn state
  // which individual views query via App.isLoggedIn().
  tryUnlockOnStartup().catch(() => undefined);
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
