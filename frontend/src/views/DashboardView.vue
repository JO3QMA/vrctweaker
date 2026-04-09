<template>
  <div class="dashboard">
    <h1 class="page-title">{{ t("routes.dashboard") }}</h1>
    <div class="quick-actions">
      <el-button
        type="primary"
        size="large"
        class="launch-btn"
        :disabled="!defaultProfile"
        @click="launch"
      >
        {{
          defaultProfile
            ? t("dashboard.launchWithProfile", { name: defaultProfile.name })
            : t("dashboard.launch")
        }}
      </el-button>
      <el-card class="status-panel" shadow="never">
        <template #header>
          <span class="status-label">{{ t("dashboard.quickStatus") }}</span>
        </template>
        <div
          class="status-buttons"
          role="group"
          :aria-label="t('dashboard.quickStatus')"
        >
          <el-button
            v-for="s in statusOptions"
            :key="s.value"
            :data-testid="s.testId"
            :class="['status-btn', s.colorClass]"
            @click="setStatusOnly(s.value)"
          >
            {{ s.label }}
          </el-button>
        </div>
      </el-card>
      <el-card class="custom-status-panel" shadow="never">
        <template #header>
          <span class="status-label">{{ t("dashboard.customStatus") }}</span>
        </template>
        <div class="custom-status-row">
          <el-input
            v-model="customDescription"
            :placeholder="t('dashboard.customStatusPlaceholder')"
            maxlength="64"
            show-word-limit
            clearable
            class="custom-status-input"
            @keyup.enter="applyCustomDescription"
          />
          <el-button
            type="primary"
            class="apply-btn"
            @click="applyCustomDescription"
          >
            {{ t("dashboard.applyDescription") }}
          </el-button>
        </div>
      </el-card>
      <el-card class="templates-panel" shadow="never">
        <template #header>
          <span class="status-label">{{ t("dashboard.templatesTitle") }}</span>
        </template>
        <div class="template-buttons">
          <el-button
            v-for="(tpl, idx) in templateDefs"
            :key="idx"
            :class="['status-btn', tpl.colorClass]"
            @click="applyTemplate(tpl)"
          >
            {{ tpl.label }}
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage } from "element-plus";
import { App, type LaunchProfileDTO } from "../wails/app";

const { t } = useI18n();

const defaultProfile = ref<LaunchProfileDTO | null>(null);
const customDescription = ref("");

const statusOptions = computed(() => [
  {
    label: t("dashboard.statusJoinMe"),
    value: "join me",
    colorClass: "status-btn--join-me",
    testId: "dashboard-quick-status-join-me",
  },
  {
    label: t("dashboard.statusActive"),
    value: "active",
    colorClass: "status-btn--active",
    testId: "dashboard-quick-status-active",
  },
  {
    label: t("dashboard.statusAskMe"),
    value: "ask me",
    colorClass: "status-btn--ask-me",
    testId: "dashboard-quick-status-ask-me",
  },
  {
    label: t("dashboard.statusBusy"),
    value: "busy",
    colorClass: "status-btn--busy",
    testId: "dashboard-quick-status-busy",
  },
]);

type TemplateDef = {
  status: string;
  messageKey: string;
  colorClass: string;
};

const templateDefs = computed(() => {
  const defs: TemplateDef[] = [
    {
      status: "busy",
      messageKey: "dashboard.templateBusyWorking",
      colorClass: "status-btn--busy",
    },
    {
      status: "join me",
      messageKey: "dashboard.templateJoinOpen",
      colorClass: "status-btn--join-me",
    },
    {
      status: "ask me",
      messageKey: "dashboard.templateAskWelcome",
      colorClass: "status-btn--ask-me",
    },
  ];
  return defs.map((d) => ({
    ...d,
    label: t(d.messageKey),
  }));
});

function formatError(e: unknown, fallback: string): string {
  if (e instanceof Error && e.message) {
    return e.message;
  }
  return fallback;
function formatError(e: unknown, fallback: string): string {
  if (e instanceof Error && e.message) {
    return e.message;
  }
  if (typeof e === "string" && e) return e;
  if (e && typeof e === "object" && "message" in e) {
    const m = (e as { message: unknown }).message;
    if (typeof m === "string" && m) return m;
  }
  return fallback;
}
});

async function launch() {
  if (!defaultProfile.value) return;
  await App.launchVRChat(defaultProfile.value.id);
}

async function setStatusOnly(status: string) {
  try {
    await App.setStatus(status);
    ElMessage.success(t("dashboard.statusUpdateSuccess"));
  } catch (e) {
    ElMessage.error(formatError(e, t("dashboard.statusUpdateError")));
  }
}

async function applyCustomDescription() {
  try {
    await App.setStatusDescription(customDescription.value.trim());
    ElMessage.success(t("dashboard.statusUpdateSuccess"));
  } catch (e) {
    ElMessage.error(formatError(e, t("dashboard.statusUpdateError")));
  }
}

async function applyTemplate(tpl: { status: string; label: string }) {
  try {
    await App.setStatusAndDescription(tpl.status, tpl.label);
    ElMessage.success(t("dashboard.statusUpdateSuccess"));
  } catch (e) {
    ElMessage.error(formatError(e, t("dashboard.statusUpdateError")));
  }
}
</script>

<style scoped>
.dashboard {
  max-width: 600px;
  margin: 0 auto;
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.launch-btn {
  width: 100%;
  font-size: 1.1rem !important;
  height: 52px !important;
}

.status-panel,
.custom-status-panel,
.templates-panel {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.status-label {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.status-buttons,
.template-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.status-btn {
  border: 1px solid transparent;
  color: #fff !important;
}

.status-btn:hover,
.status-btn:focus {
  filter: brightness(1.08);
  color: #fff !important;
}

.status-btn--join-me {
  background: #2b7fd9 !important;
  border-color: #256bb8 !important;
}

.status-btn--active {
  background: #2e9f4a !important;
  border-color: #267d3c !important;
}

.status-btn--ask-me {
  background: #e8943c !important;
  border-color: #c97d2e !important;
}

.status-btn--busy {
  background: #d94a4a !important;
  border-color: #b83c3c !important;
}

.custom-status-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: flex-start;
}

.custom-status-input {
  flex: 1 1 200px;
  min-width: 0;
}

.apply-btn {
  flex-shrink: 0;
}
</style>
