<template>
  <nav class="sidebar">
    <el-menu :default-active="route.path" router class="sidebar-nav">
      <el-menu-item
        v-for="item in menuItems"
        :key="item.path"
        :index="item.path"
      >
        <span class="sidebar-icon">{{ item.icon }}</span>
        <template #title>{{ item.label }}</template>
      </el-menu-item>
    </el-menu>
    <div class="sidebar-footer">
      <el-menu :default-active="route.path" router class="sidebar-nav">
        <el-menu-item index="/settings">
          <span class="sidebar-icon">⚙️</span>
          <template #title>{{ t("nav.settings") }}</template>
        </el-menu-item>
      </el-menu>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";

const route = useRoute();
const { t } = useI18n();

const menuItems = computed(() => [
  { path: "/", icon: "🏠", label: t("nav.dashboard") },
  { path: "/launcher", icon: "🚀", label: t("nav.launcher") },
  { path: "/gallery", icon: "🖼️", label: t("nav.gallery") },
  { path: "/activity", icon: "📊", label: t("nav.activity") },
  { path: "/friends", icon: "👥", label: t("nav.friends") },
  { path: "/automation", icon: "🤖", label: t("nav.automation") },
  { path: "/config", icon: "📝", label: t("nav.configOther") },
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

.sidebar-nav {
  background: transparent;
  border-right: none !important;
}

.sidebar-nav :deep(.el-menu-item) {
  color: var(--text-secondary);
  height: 42px;
  line-height: 42px;
}

.sidebar-nav :deep(.el-menu-item:hover),
.sidebar-nav :deep(.el-menu-item.is-active) {
  background: var(--bg-tertiary) !important;
  color: var(--text-primary) !important;
}

.sidebar-nav :deep(.el-menu-item.is-active) {
  border-left: 3px solid var(--accent);
}

.sidebar-footer {
  margin-top: auto;
  border-top: 1px solid var(--border);
}

.sidebar-icon {
  margin-right: 0.5rem;
  font-size: 1rem;
}
</style>
