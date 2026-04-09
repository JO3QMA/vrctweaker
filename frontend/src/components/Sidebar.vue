<template>
  <nav class="sidebar" :aria-label="t('sidebar.a11yLabel')">
    <div class="sidebar-main" role="menubar">
      <router-link
        v-for="item in menuItems"
        :key="item.path"
        v-slot="{ href, navigate, isActive }"
        :to="item.path"
        custom
      >
        <a
          :href="href"
          role="menuitem"
          class="sidebar-item"
          :class="{ 'sidebar-item--active': isActive }"
          :aria-current="isActive ? 'page' : undefined"
          @click="(e) => navigate(e)"
        >
          <span class="sidebar-icon" aria-hidden="true">{{ item.icon }}</span>
          <span class="sidebar-label">{{ item.label }}</span>
        </a>
      </router-link>
    </div>
    <div class="sidebar-footer" role="menubar">
      <router-link v-slot="{ href, navigate, isActive }" to="/settings" custom>
        <a
          :href="href"
          role="menuitem"
          class="sidebar-item"
          :class="{ 'sidebar-item--active': isActive }"
          :aria-current="isActive ? 'page' : undefined"
          @click="(e) => navigate(e)"
        >
          <span class="sidebar-icon" aria-hidden="true">⚙️</span>
          <span class="sidebar-label">{{ t("sidebar.settings") }}</span>
        </a>
      </router-link>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const menuItems = computed(() => [
  { path: "/", icon: "🏠", label: t("sidebar.dashboard") },
  { path: "/launcher", icon: "🚀", label: t("sidebar.launcher") },
  { path: "/gallery", icon: "🖼️", label: t("sidebar.gallery") },
  { path: "/activity", icon: "📊", label: t("sidebar.activity") },
  { path: "/friends", icon: "👥", label: t("sidebar.friends") },
  { path: "/automation", icon: "🤖", label: t("sidebar.automation") },
  { path: "/config", icon: "📝", label: t("sidebar.config") },
]);
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  background: var(--bg-secondary);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-main {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.sidebar-item {
  display: flex;
  align-items: center;
  box-sizing: border-box;
  height: 42px;
  padding: 0 1rem 0 calc(1rem - 3px);
  margin: 0;
  color: var(--text-secondary);
  text-decoration: none;
  cursor: pointer;
  border-left: 3px solid transparent;
  line-height: 42px;
}

.sidebar-item:hover,
.sidebar-item--active {
  background: var(--bg-tertiary) !important;
  color: var(--text-primary) !important;
}

.sidebar-item--active {
  border-left-color: var(--accent);
}

.sidebar-footer {
  margin-top: auto;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}

.sidebar-icon {
  margin-right: 0.5rem;
  font-size: 1rem;
  flex-shrink: 0;
}

.sidebar-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
