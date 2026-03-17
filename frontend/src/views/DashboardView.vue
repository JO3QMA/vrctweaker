<template>
  <div class="dashboard">
    <h1 class="page-title">
      ダッシュボード
    </h1>
    <div class="quick-actions">
      <button
        class="launch-btn"
        :disabled="!defaultProfile"
        @click="launch"
      >
        {{
          defaultProfile
            ? `VRChat 起動 (${defaultProfile.name})`
            : "VRChat 起動"
        }}
      </button>
      <div class="status-panel">
        <p class="status-label">
          クイックステータス
        </p>
        <div class="status-buttons">
          <button
            v-for="s in statusOptions"
            :key="s.value"
            class="status-btn"
            @click="setStatus(s.value)"
          >
            {{ s.label }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { App, type LaunchProfileDTO } from "../wails/app";

const defaultProfile = ref<LaunchProfileDTO | null>(null);
const statusOptions = [
  { label: "Join Me", value: "join me" },
  { label: "Ask Me", value: "ask me" },
  { label: "Busy", value: "busy" },
];

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

.page-title {
  margin: 0 0 1.5rem;
  font-size: 1.5rem;
}

.launch-btn {
  width: 100%;
  padding: 1rem 2rem;
  font-size: 1.2rem;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: var(--radius);
  margin-bottom: 1rem;
}

.launch-btn:hover {
  background: var(--accent-hover);
}

.status-panel {
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--radius);
}

.status-label {
  margin: 0 0 0.5rem;
  font-size: 0.9rem;
  color: var(--text-secondary);
}
.status-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.status-btn {
  padding: 0.4rem 0.8rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
}
.status-btn:hover {
  background: var(--accent);
  color: white;
}
</style>
