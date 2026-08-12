<template>
  <nav class="sidebar">
    <el-menu :default-active="route.path" router class="sidebar-nav">
      <el-menu-item
        v-for="item in menuItems"
        :key="item.path"
        :index="item.path"
      >
        <span class="sidebar-icon nav-glyph-size-default">{{ item.icon }}</span>
        <template #title>{{ item.label }}</template>
      </el-menu-item>
    </el-menu>
    <div class="sidebar-footer">
      <el-menu :default-active="route.path" router class="sidebar-nav">
        <el-menu-item index="/settings">
          <span class="sidebar-icon nav-glyph-size-default">⚙️</span>
          <template #title>{{ t("nav.settings") }}</template>
        </el-menu-item>
      </el-menu>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { App } from "../wails/app";

const route = useRoute();
const { t } = useI18n();
const isWindows = ref(false);

onMounted(async () => {
  try {
    isWindows.value = await App.runtimeIsWindows();
  } catch {
    isWindows.value = false;
  }
});

const menuItems = computed(() => {
  const items = [
    { path: "/", icon: "🏠", label: t("nav.dashboard") },
    { path: "/launcher", icon: "🚀", label: t("nav.launcher") },
    { path: "/gallery", icon: "🖼️", label: t("nav.gallery") },
    { path: "/activity", icon: "📊", label: t("nav.activity") },
    { path: "/me", icon: "👤", label: t("nav.me") },
    { path: "/friends", icon: "👥", label: t("nav.friends") },
    { path: "/automation", icon: "🤖", label: t("nav.automation") },
    { path: "/config", icon: "📝", label: t("nav.configOther") },
  ];
  if (isWindows.value) {
    const configIdx = items.findIndex((item) => item.path === "/config");
    const insertAt = configIdx === -1 ? items.length : configIdx;
    return [
      ...items.slice(0, insertAt),
      { path: "/video", icon: "🎬", label: t("nav.video") },
      ...items.slice(insertAt),
    ];
  }
  return items;
});
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  background: var(--color-bg-elevated);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-nav {
  background: transparent;
  border-right: none !important;
}

.sidebar-nav :deep(.el-menu-item) {
  color: var(--color-text-secondary);
  height: 42px;
  line-height: 42px;
}

.sidebar-nav :deep(.el-menu-item:hover),
.sidebar-nav :deep(.el-menu-item.is-active) {
  background: var(--color-bg-muted) !important;
  color: var(--color-text-primary) !important;
}

.sidebar-nav :deep(.el-menu-item.is-active) {
  border-left: 3px solid var(--color-brand);
}

.sidebar-footer {
  margin-top: auto;
  border-top: 1px solid var(--color-border);
}

.sidebar-icon {
  margin-right: var(--space-action-group);
}
</style>
