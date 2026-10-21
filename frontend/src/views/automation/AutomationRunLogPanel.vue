<script setup lang="ts">
import { useI18n } from "vue-i18n";
import VtButton from "../../components/VtButton.vue";
import VtTag from "../../components/VtTag.vue";
import type { AutomationRunLogEntryDTO } from "../../wails/app";

defineProps<{
  entries: AutomationRunLogEntryDTO[];
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const { t } = useI18n();
</script>

<template>
  <el-card
    class="section-card automation-log"
    shadow="never"
    data-testid="automation-run-log"
  >
    <template #header>
      <span class="automation-log__title">{{
        t("automation.runLogTitle")
      }}</span>
      <VtButton
        variant="tertiary"
        size="small"
        data-testid="refresh-run-log"
        @click="emit('refresh')"
      >
        {{ t("common.refresh") }}
      </VtButton>
    </template>
    <el-table
      :data="entries"
      size="small"
      stripe
      empty-text="—"
      class="automation-log__table"
      data-testid="automation-run-log-table"
    >
      <el-table-column prop="at" :label="t('automation.runLogAt')" />
      <el-table-column prop="itemName" :label="t('automation.itemName')" />
      <el-table-column prop="eventType" :label="t('automation.runLogEvent')" />
      <el-table-column :label="t('automation.runLogResult')">
        <template #default="{ row }">
          <VtTag :variant="row.success ? 'success' : 'danger'" size="small">
            {{
              row.success
                ? t("automation.runLogSuccess")
                : t("automation.runLogFailure")
            }}
          </VtTag>
        </template>
      </el-table-column>
      <el-table-column :label="t('automation.runLogActions')">
        <template #default="{ row }">
          {{ row.actionsCompleted }}/{{ row.actionsTotal }}
        </template>
      </el-table-column>
      <el-table-column
        prop="contextLabel"
        :label="t('automation.runLogContext')"
      />
    </el-table>
  </el-card>
</template>

<style scoped>
.automation-log {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  margin-bottom: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.automation-log :deep(.el-card__header) {
  height: 48px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-action-group);
  padding: 0 var(--space-block);
}

.automation-log__title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.automation-log :deep(.el-card__body) {
  flex: 1 1 auto;
  min-height: 0;
  min-width: 0;
  overflow: auto;
  scrollbar-gutter: stable;
  padding: 0;
}

.automation-log__table {
  width: 100%;
}

.automation-log :deep(.el-table) {
  width: 100%;
  --el-table-header-bg-color: var(--color-bg-muted);
}

.automation-log :deep(.cell) {
  white-space: normal;
  overflow-wrap: anywhere;
  word-break: break-word;
  line-height: var(--line-height-normal);
}
</style>
