<template>
  <div class="dashboard">
    <h1 class="page-title">{{ t("dashboard.title") }}</h1>
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
            ? t("dashboard.launchWithProfile", {
                name: defaultProfile.name,
              })
            : t("dashboard.launchVRChat")
        }}
      </el-button>
      <el-card class="status-panel" shadow="never">
        <template #header>
          <span class="status-label">{{ t("dashboard.quickStatus") }}</span>
        </template>
        <div class="status-buttons">
          <el-button
            v-for="s in statusOptions"
            :key="s.value"
            @click="setStatus(s.value)"
          >
            {{ s.label }}
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { App, type LaunchProfileDTO } from "../wails/app";

const { t } = useI18n();

const defaultProfile = ref<LaunchProfileDTO | null>(null);

const statusOptions = computed(() => [
  { label: t("dashboard.statusActive"), value: "active" },
  { label: t("dashboard.statusJoinMe"), value: "join me" },
  { label: t("dashboard.statusAskMe"), value: "ask me" },
  { label: t("dashboard.statusBusy"), value: "busy" },
]);

onMounted(async () => {
  const profiles = await App.launchProfiles();
  defaultProfile.value =
    profiles.find((p) => p.isDefault) ?? profiles[0] ?? null;
});

async function launch() {
  if (!defaultProfile.value) return;
  await App.launchVRChat(defaultProfile.value.id);
}

async function setStatus(status: string) {
  await App.setStatus(status);
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

.status-panel {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.status-label {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.status-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
</style>
