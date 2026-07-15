<template>
  <el-card
    class="dashboard-launch-block"
    shadow="never"
    data-testid="dashboard-launch-block"
  >
    <div
      v-if="loading"
      class="launch-block-message"
      data-testid="launch-block-loading"
    >
      {{ t("dashboard.launchBlock.loading") }}
    </div>

    <div
      v-else-if="loadError"
      class="launch-block-message"
      data-testid="launch-block-load-error"
    >
      {{ t("dashboard.launchBlock.loadError") }}
    </div>

    <template v-else>
      <p
        v-if="isEmpty"
        class="launch-block-empty"
        data-testid="launch-block-empty-state"
      >
        {{ t("dashboard.launchBlock.emptyState") }}
        <router-link
          to="/launcher"
          class="launch-block-launcher-link"
          data-testid="launch-block-launcher-link"
        >
          {{ t("dashboard.launchBlock.goToLauncher") }}
        </router-link>
      </p>

      <div class="launch-block-controls">
        <el-select
          v-model="selectedProfileId"
          class="launch-block-profile"
          data-testid="launch-block-profile-select"
          :placeholder="t('dashboard.launchBlock.profilePlaceholder')"
          :disabled="isEmpty"
        >
          <el-option
            v-for="p in profiles"
            :key="p.id"
            :label="p.name"
            :value="p.id"
          />
        </el-select>
        <div
          class="launch-block-actions"
          :class="{ 'launch-block-actions--solo': !rejoin }"
        >
          <el-button
            type="primary"
            class="launch-block-quick-btn"
            data-testid="launch-block-quick-btn"
            :disabled="isEmpty || !selectedProfileId || launching"
            :loading="launching"
            @click="launchQuick"
          >
            {{ t("dashboard.launchBlock.quickLaunch") }}
          </el-button>
          <el-button
            v-if="rejoin"
            type="primary"
            class="launch-block-rejoin-btn"
            data-testid="launch-block-rejoin-btn"
            :disabled="
              isEmpty ||
              !selectedProfileId ||
              !rejoin.playSessionId?.trim() ||
              launching
            "
            :loading="launching"
            @click="launchRejoin"
          >
            {{ rejoinButtonLabel }}
          </el-button>
        </div>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage } from "element-plus";
import {
  App,
  type DashboardLaunchBlockDTO,
  type DashboardRejoinDTO,
  type LaunchProfileDTO,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import { formatError } from "../utils/formatError";
import { truncateRejoinWorldName } from "../utils/truncateText";

const ACTIVITY_ENCOUNTERS_CHANGED_DEBOUNCE_MS = 400;

const { t } = useI18n();

const loading = ref(true);
const loadError = ref(false);
const launching = ref(false);
const profiles = ref<LaunchProfileDTO[]>([]);
const selectedProfileId = ref("");
const rejoin = ref<DashboardRejoinDTO | null>(null);

let encountersChangedDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribeEncountersChanged: (() => void) | undefined;
let generation = 0;
let inFlight = false;
let pendingRefresh = false;
let hasLoadedOnce = false;

const isEmpty = computed(() => profiles.value.length === 0);

const rejoinButtonLabel = computed(() => {
  const name = rejoin.value?.worldDisplayName?.trim();
  if (name) {
    return t("dashboard.launchBlock.rejoinWithWorld", {
      name: truncateRejoinWorldName(name),
    });
  }
  return t("dashboard.launchBlock.rejoinGeneric");
});

function applyBlock(dto: DashboardLaunchBlockDTO, isInitial: boolean): void {
  profiles.value = dto.profiles ?? [];
  rejoin.value = dto.rejoin ?? null;
  if (isInitial) {
    selectedProfileId.value = dto.selectedProfileId ?? "";
  }
}

async function load(): Promise<void> {
  if (inFlight) {
    pendingRefresh = true;
    return;
  }
  inFlight = true;
  pendingRefresh = false;
  const gen = generation;
  try {
    const dto = await App.getDashboardLaunchBlock();
    if (gen !== generation) return;
    loadError.value = false;
    applyBlock(dto, !hasLoadedOnce);
    hasLoadedOnce = true;
  } catch (e) {
    if (gen !== generation) return;
    console.error("DashboardLaunchBlock load failed:", e);
    if (!hasLoadedOnce) {
      loadError.value = true;
      profiles.value = [];
      selectedProfileId.value = "";
      rejoin.value = null;
    } else {
      ElMessage.warning(t("dashboard.launchBlock.refreshError"));
    }
  } finally {
    inFlight = false;
    if (gen === generation) {
      loading.value = false;
    }
    if (pendingRefresh && gen === generation) {
      pendingRefresh = false;
      void load();
    }
  }
}

function scheduleRefresh(): void {
  if (encountersChangedDebounceTimer !== null) {
    clearTimeout(encountersChangedDebounceTimer);
  }
  encountersChangedDebounceTimer = setTimeout(() => {
    encountersChangedDebounceTimer = null;
    void load();
  }, ACTIVITY_ENCOUNTERS_CHANGED_DEBOUNCE_MS);
}

async function launchQuick(): Promise<void> {
  if (!selectedProfileId.value || launching.value) return;
  launching.value = true;
  try {
    await App.launchVRChat(selectedProfileId.value);
  } catch (e) {
    ElMessage.error(formatError(e, t("dashboard.launchBlock.launchError")));
  } finally {
    launching.value = false;
  }
}

async function launchRejoin(): Promise<void> {
  const playSessionId = rejoin.value?.playSessionId?.trim();
  if (!selectedProfileId.value || !playSessionId || launching.value) return;
  launching.value = true;
  try {
    await App.instanceRejoin(selectedProfileId.value, playSessionId);
  } catch (e) {
    ElMessage.error(formatError(e, t("dashboard.launchBlock.rejoinError")));
    void load();
  } finally {
    launching.value = false;
  }
}

onMounted(async () => {
  await load();
  const rt = getRuntime();
  const off = rt?.EventsOn?.("activity:encounters-changed", () => {
    scheduleRefresh();
  });
  if (typeof off === "function") {
    unsubscribeEncountersChanged = off;
  }
});

onUnmounted(() => {
  generation += 1;
  if (encountersChangedDebounceTimer !== null) {
    clearTimeout(encountersChangedDebounceTimer);
    encountersChangedDebounceTimer = null;
  }
  unsubscribeEncountersChanged?.();
});
</script>

<style scoped>
.dashboard-launch-block {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.launch-block-message,
.launch-block-empty {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0 0 0.75rem;
}

.launch-block-launcher-link {
  margin-left: 0.35rem;
}

.launch-block-controls {
  --launch-block-gap: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: var(--launch-block-gap);
}

.launch-block-profile {
  width: 100%;
}

.launch-block-actions {
  display: flex;
  gap: var(--launch-block-gap);
  width: 100%;
}

.launch-block-actions :deep(.el-button) {
  margin: 0;
}

.launch-block-quick-btn,
.launch-block-rejoin-btn {
  flex: 1 1 50%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.launch-block-actions--solo .launch-block-quick-btn {
  flex: 1 1 100%;
}
</style>
