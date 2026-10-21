<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
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
const root = ref<HTMLElement | null>(null);
const compact = ref(false);
let observer: ResizeObserver | undefined;

const eventLabelKeys: Record<string, string> = {
  friend_joined: "automation.triggerFriendJoined",
  friend_left: "automation.triggerFriendLeft",
  "schedule.tick": "automation.triggerSchedule",
  "vrchat.process": "automation.triggerVrchatProcess",
  afk_detected: "automation.triggerAfk",
};

function eventLabel(eventType: string): string {
  if (Object.prototype.hasOwnProperty.call(eventLabelKeys, eventType)) {
    return t(eventLabelKeys[eventType] ?? eventType);
  }
  return eventType;
}

function resultTooltip(row: AutomationRunLogEntryDTO): string {
  const progress = `${t("automation.runLogActions")} ${row.actionsCompleted}/${row.actionsTotal}`;
  const context = row.contextLabel
    ? `${t("automation.runLogContext")} ${row.contextLabel}`
    : "";
  return context ? `${progress} / ${context}` : progress;
}

onMounted(() => {
  observer = new ResizeObserver(() => {
    const width = root.value?.offsetWidth ?? 0;
    compact.value = width > 0 && width < 720;
  });
  if (root.value) observer.observe(root.value);
});

onBeforeUnmount(() => {
  observer?.disconnect();
});
</script>

<template>
  <div ref="root" class="automation-log" data-testid="automation-run-log">
    <div class="automation-log__toolbar">
      <VtButton
        variant="tertiary"
        size="small"
        data-testid="refresh-run-log"
        @click="emit('refresh')"
      >
        {{ t("common.refresh") }}
      </VtButton>
    </div>
    <div class="automation-log__scroll">
      <el-table
        :data="entries"
        size="small"
        stripe
        empty-text="—"
        class="automation-log__table"
        data-testid="automation-run-log-table"
        table-layout="fixed"
      >
        <el-table-column
          prop="at"
          :label="t('automation.runLogAt')"
          :width="compact ? 120 : 160"
          show-overflow-tooltip
        />
        <el-table-column
          prop="itemName"
          :label="t('automation.itemName')"
          show-overflow-tooltip
        />
        <el-table-column
          :label="t('automation.runLogEvent')"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ eventLabel(row.eventType) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('automation.runLogResult')" width="96">
          <template #default="{ row }">
            <el-tooltip
              v-if="compact"
              :content="resultTooltip(row)"
              placement="top"
            >
              <VtTag :variant="row.success ? 'success' : 'danger'" size="small">
                {{
                  row.success
                    ? t("automation.runLogSuccess")
                    : t("automation.runLogFailure")
                }}
              </VtTag>
            </el-tooltip>
            <VtTag
              v-else
              :variant="row.success ? 'success' : 'danger'"
              size="small"
            >
              {{
                row.success
                  ? t("automation.runLogSuccess")
                  : t("automation.runLogFailure")
              }}
            </VtTag>
          </template>
        </el-table-column>
        <el-table-column
          v-if="!compact"
          :label="t('automation.runLogActions')"
          width="72"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.actionsCompleted }}/{{ row.actionsTotal }}
          </template>
        </el-table-column>
        <el-table-column
          v-if="!compact"
          prop="contextLabel"
          :label="t('automation.runLogContext')"
          show-overflow-tooltip
        />
      </el-table>
    </div>
  </div>
</template>

<style scoped>
.automation-log {
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-bg-elevated);
}

.automation-log__toolbar {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  height: 48px;
  box-sizing: border-box;
  padding: 0 var(--space-block);
  border-bottom: 1px solid var(--color-border);
}

.automation-log__scroll {
  flex: 1 1 auto;
  min-height: 0;
  min-width: 0;
  overflow: auto;
  scrollbar-gutter: stable;
}

.automation-log__table {
  width: 100%;
}

.automation-log :deep(.el-table) {
  width: 100%;
  --el-table-header-bg-color: var(--color-bg-muted);
}

.automation-log :deep(.el-table__header),
.automation-log :deep(.el-table__body) {
  width: 100%;
  table-layout: fixed;
}

.automation-log :deep(.cell) {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: var(--line-height-normal);
}
</style>
