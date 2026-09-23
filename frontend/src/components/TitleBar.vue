<template>
  <div class="title-bar" style="--wails-draggable: drag">
    <div class="title-bar-brand">
      <img
        v-if="showAppIcon"
        class="title-bar-icon"
        :src="appIconUrl"
        alt=""
        aria-hidden="true"
        width="16"
        height="16"
        draggable="false"
        data-testid="title-bar-app-icon"
        @error="showAppIcon = false"
      />
      <span class="title-bar-text">{{ t("app.name") }}</span>
    </div>
    <div class="title-bar-actions" style="--wails-draggable: no-drag">
      <button class="title-bar-btn vt-focus-ring--inset" @click="minimize">
        −
      </button>
      <button class="title-bar-btn vt-focus-ring--inset" @click="maximize">
        □
      </button>
      <button class="title-bar-btn close vt-focus-ring--inset" @click="close">
        ×
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { App } from "../wails/app";
import { getRuntime } from "../wails/runtime";
import { formatError } from "../utils/formatError";
import { showToast } from "../utils/showToast";

const { t } = useI18n();
const showAppIcon = ref(true);
// Copied from build/appicon.png via `pnpm run sync-appicon` (predev / build).
const appIconUrl = `${import.meta.env.BASE_URL}appicon.png`;

function minimize() {
  getRuntime()?.WindowMinimise?.();
}

function maximize() {
  getRuntime()?.WindowToggleMaximise?.();
}

async function close() {
  try {
    await App.requestClose();
  } catch (e) {
    showToast.error(formatError(e, t("app.errClose")));
  }
}
</script>

<style scoped>
.title-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 36px;
  padding: 0 var(--space-action-group);
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
  user-select: none;
  flex-shrink: 0;
}

.title-bar-brand {
  display: flex;
  align-items: center;
  gap: var(--space-inline-tight);
  min-width: 0;
}

.title-bar-icon {
  width: var(--icon-size-16);
  height: var(--icon-size-16);
  flex-shrink: 0;
  object-fit: contain;
}

.title-bar-text {
  font-size: var(--font-size-12);
  font-weight: var(--font-weight-500);
  color: var(--color-text-secondary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.title-bar-actions {
  display: flex;
  gap: 0;
}

.title-bar-btn {
  width: 40px;
  height: 36px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: var(--font-size-16);
  line-height: 1;
  cursor: pointer;
  transition:
    background 0.15s,
    color 0.15s;
}

.title-bar-btn:hover {
  background: var(--color-bg-muted);
  color: var(--color-text-primary);
}

.title-bar-btn.close:hover {
  background: var(--color-danger);
  color: var(--color-text-inverse);
}
</style>
