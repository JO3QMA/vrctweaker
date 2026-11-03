<template>
  <div class="startup-screen" data-testid="startup-screen">
    <TitleBar />
    <main class="startup-screen__main">
      <section
        v-if="loading"
        class="startup-panel startup-panel--loading"
        data-testid="startup-loading"
      >
        <div class="startup-panel__lead-row">
          <VtIcon size="compact" class="startup-panel__spinner is-loading">
            <Loading />
          </VtIcon>
          <p class="text-body startup-panel__lead">
            {{ loadingMessage }}
          </p>
        </div>
      </section>

      <section
        v-else-if="failed"
        class="startup-panel startup-panel--failure"
        data-testid="startup-failure"
      >
        <VtAlert variant="danger" :title="summaryText" />
        <el-collapse class="startup-detail vt-collapse-panel">
          <el-collapse-item :title="t('app.startup.details')" name="detail">
            <pre
              class="startup-detail__body text-body-sm vt-focus-ring"
              data-testid="startup-failure-detail"
              tabindex="0"
              >{{ status.detail }}</pre>
          </el-collapse-item>
        </el-collapse>
        <div class="startup-actions">
          <VtButton
            variant="primary"
            class="vt-focus-ring startup-retry-btn"
            data-testid="startup-retry"
            :loading="retrying"
            @click="onRetry"
          >
            {{ t("app.startup.retry") }}
          </VtButton>
          <VtButton
            v-if="status.canOpenDataFolder"
            variant="secondary"
            class="vt-btn-secondary-dark"
            data-testid="startup-open-data-folder"
            @click="onOpenDataFolder"
          >
            {{ t("app.startup.openDataFolder") }}
          </VtButton>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { Loading } from "@element-plus/icons-vue";
import TitleBar from "./TitleBar.vue";
import VtAlert from "./VtAlert.vue";
import VtButton from "./VtButton.vue";
import VtIcon from "./VtIcon.vue";
import type { StartupInitStatus } from "../composables/useStartupInit";

const props = defineProps<{
  status: StartupInitStatus;
  retrying: boolean;
}>();

const emit = defineEmits<{
  retry: [];
  openDataFolder: [];
}>();

const { t, te } = useI18n();

const loading = computed(
  () =>
    props.status.loading || (!props.status.ready && !props.status.summaryKey),
);
const failed = computed(
  () => !props.status.ready && !props.status.loading && props.status.summaryKey,
);

const loadingMessage = computed(() => {
  if (props.status.phase === "dedupe") {
    return t("app.startup.migrating");
  }
  return t("app.startup.loading");
});

const summaryText = computed(() => {
  const key = props.status.summaryKey;
  if (key && te(key)) {
    return t(key);
  }
  return t("app.startup.failed.databaseOpen");
});

function onRetry(): void {
  emit("retry");
}

function onOpenDataFolder(): void {
  emit("openDataFolder");
}
</script>

<style scoped>
.startup-screen {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-base);
}

.startup-screen__main {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-section);
}

.startup-panel {
  width: min(100%, 560px);
  display: flex;
  flex-direction: column;
  gap: var(--space-block);
}

.startup-panel--loading {
  align-items: center;
}

.startup-panel__lead-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-inline-tight);
}

.startup-panel__spinner {
  color: var(--color-text-secondary);
}

@media (prefers-reduced-motion: reduce) {
  .startup-panel__spinner.is-loading {
    animation: none;
  }
}

.startup-panel__lead {
  margin: 0;
  text-align: center;
  color: var(--color-text-secondary);
}

.startup-detail__body {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--color-text-secondary);
  max-height: 168px;
  overflow: auto;
  padding: var(--space-form-field);
  background: var(--color-bg-base);
  border-radius: var(--radius);
}

.startup-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-action-group);
  justify-content: center;
}

.startup-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.startup-actions :deep(.startup-retry-btn.el-button) {
  min-width: 96px;
}

.startup-actions :deep(.vt-focus-ring.el-button:focus:not(:focus-visible)) {
  outline: none;
  box-shadow: none;
}
</style>
