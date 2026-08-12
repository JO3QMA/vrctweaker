<template>
  <div class="title-bar" style="--wails-draggable: drag">
    <span class="title-bar-text">{{ t("app.name") }}</span>
    <div class="title-bar-actions" style="--wails-draggable: no-drag">
      <button class="title-bar-btn" @click="minimize">−</button>
      <button class="title-bar-btn" @click="maximize">□</button>
      <button class="title-bar-btn close" @click="close">×</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { getRuntime } from "../wails/runtime";

const { t } = useI18n();

function minimize() {
  getRuntime()?.WindowMinimise?.();
}

function maximize() {
  getRuntime()?.WindowToggleMaximise?.();
}

function close() {
  getRuntime()?.Quit?.();
}
</script>

<style scoped>
.title-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 36px;
  padding: 0 0.5rem;
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
  user-select: none;
  flex-shrink: 0;
}

.title-bar-text {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text-secondary);
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
  font-size: 16px;
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
