<template>
  <nav class="sidebar">
    <el-menu :default-active="route.path" router class="sidebar-nav">
      <el-menu-item
        v-for="item in menuItems"
        :key="item.path"
        :index="item.path"
        class="vt-focus-ring--inset"
      >
        <VtIcon size="default" class="sidebar-icon">
          <component :is="item.icon" />
        </VtIcon>
        <template #title>{{ item.label }}</template>
      </el-menu-item>
    </el-menu>
    <div class="sidebar-footer">
      <el-menu :default-active="route.path" router class="sidebar-nav">
        <el-menu-item :index="settingsItem.path" class="vt-focus-ring--inset">
          <VtIcon size="default" class="sidebar-icon">
            <component :is="settingsItem.icon" />
          </VtIcon>
          <template #title>{{ settingsItem.label }}</template>
        </el-menu-item>
      </el-menu>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import {
  NAV_MAIN_MENU_ITEMS,
  NAV_SETTINGS_MENU_ITEM,
} from "../navigation/navMenuItems";
import VtIcon from "./VtIcon.vue";
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

const menuItems = computed(() =>
  NAV_MAIN_MENU_ITEMS.filter(
    (item) => !item.windowsOnly || isWindows.value,
  ).map((item) => ({
    ...item,
    label: t(item.labelKey),
  })),
);

const settingsItem = computed(() => ({
  ...NAV_SETTINGS_MENU_ITEM,
  label: t(NAV_SETTINGS_MENU_ITEM.labelKey),
}));
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
  border-left: 3px solid transparent;
}

.sidebar-nav :deep(.el-menu-item:hover) {
  background: var(--color-bg-muted) !important;
  color: var(--color-text-primary) !important;
}

.sidebar-nav :deep(.el-menu-item.is-active) {
  background: var(--color-bg-muted) !important;
  color: var(--color-brand) !important;
  font-weight: var(--font-weight-600);
  border-left-color: var(--color-brand);
}

.sidebar-footer {
  margin-top: auto;
  border-top: 1px solid var(--color-border);
}

.sidebar-icon {
  margin-right: var(--space-action-group);
}
</style>
