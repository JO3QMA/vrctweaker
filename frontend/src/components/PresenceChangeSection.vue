<template>
  <el-card
    class="presence-change-section"
    shadow="never"
    data-testid="presence-change-section"
  >
    <template #header>
      <span class="presence-change-title">{{
        t("dashboard.presenceChange.title")
      }}</span>
    </template>

    <div
      v-if="loading"
      class="presence-change-message"
      data-testid="presence-change-loading"
    >
      {{ t("dashboard.presenceChange.loading") }}
    </div>

    <div
      v-else-if="loadError"
      class="presence-change-message"
      data-testid="presence-change-load-error"
    >
      <p class="presence-change-error-text">
        {{ t("dashboard.presenceChange.loadError") }}
      </p>
      <el-button
        size="small"
        data-testid="presence-change-retry"
        @click="retryLoad"
      >
        {{ t("dashboard.presenceChange.retry") }}
      </el-button>
    </div>

    <template v-else>
      <p
        v-if="!loggedIn"
        class="presence-change-login-hint"
        data-testid="presence-change-login-required"
      >
        {{ t("dashboard.presenceChange.loginRequired") }}
        <router-link
          to="/settings"
          class="presence-change-settings-link"
          data-testid="presence-change-settings-link"
        >
          {{ t("dashboard.presenceChange.goToSettings") }}
        </router-link>
      </p>

      <div
        class="presence-color-buttons"
        role="group"
        :aria-label="t('dashboard.presenceChange.colorGroupLabel')"
      >
        <el-button
          v-for="option in colorOptions"
          :key="option.value"
          :data-testid="option.testId"
          :class="[
            'presence-color-btn',
            option.colorClass,
            { 'presence-color-btn--selected': draftStatus === option.value },
          ]"
          :disabled="!loggedIn || applying"
          @click="selectColor(option.value)"
        >
          {{ option.label }}
        </el-button>
      </div>

      <div class="presence-description-row">
        <el-autocomplete
          v-model="draftDescription"
          :fetch-suggestions="queryHistory"
          :placeholder="t('dashboard.presenceChange.descriptionPlaceholder')"
          :maxlength="32"
          clearable
          class="presence-description-input"
          data-testid="presence-change-description"
          :disabled="!loggedIn || applying"
          @select="onDescriptionSelect"
        />
        <el-button
          type="primary"
          class="presence-apply-btn"
          data-testid="presence-change-apply"
          :disabled="!loggedIn || !isDirty || applying"
          :loading="applying"
          @click="applyPresenceChange"
        >
          {{ t("dashboard.presenceChange.apply") }}
        </el-button>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage } from "element-plus";
import { App, type PresenceChangeSectionDTO } from "../wails/app";
import { getRuntime } from "../wails/runtime";
import { formatError } from "../utils/formatError";
import { useSessionUnlock } from "../composables/useSessionUnlock";

const SELF_CACHE_CHANGED_DEBOUNCE_MS = 300;

type PresenceStatus = "join me" | "active" | "ask me" | "busy";

const { t } = useI18n();
const { state: unlockState } = useSessionUnlock();

const loading = ref(true);
const loadError = ref(false);
const applying = ref(false);
const loggedIn = ref(false);
const history = ref<string[]>([]);
const draftStatus = ref<PresenceStatus>("active");
const draftDescription = ref("");
const snapshotStatus = ref<PresenceStatus>("active");
const snapshotDescription = ref("");

let selfCacheDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribeSelfCacheChanged: (() => void) | undefined;
// Load mutex + pending queue: inFlight blocks overlap; pendingRefresh replays one
// deferred load in finally. pendingRefreshOnlyIfNotDirty is OR-merged so any queued
// onlyIfNotDirty request makes the replay conservative. generation drops stale responses.
let generation = 0;
let inFlight = false;
let pendingRefresh = false;
let pendingRefreshOnlyIfNotDirty = false;
let hasLoadedOnce = false;

const colorOptions = computed(() => [
  {
    value: "join me" as const,
    label: t("dashboard.presenceChange.colorJoinMe"),
    colorClass: "presence-color-btn--join-me",
    testId: "presence-change-color-join-me",
  },
  {
    value: "active" as const,
    label: t("dashboard.presenceChange.colorActive"),
    colorClass: "presence-color-btn--active",
    testId: "presence-change-color-active",
  },
  {
    value: "ask me" as const,
    label: t("dashboard.presenceChange.colorAskMe"),
    colorClass: "presence-color-btn--ask-me",
    testId: "presence-change-color-ask-me",
  },
  {
    value: "busy" as const,
    label: t("dashboard.presenceChange.colorBusy"),
    colorClass: "presence-color-btn--busy",
    testId: "presence-change-color-busy",
  },
]);

const isDirty = computed(
  () =>
    draftStatus.value !== snapshotStatus.value ||
    draftDescription.value !== snapshotDescription.value,
);

function applySection(
  dto: PresenceChangeSectionDTO,
  options: { isInitial?: boolean; skipSnapshotUpdate?: boolean } = {},
): void {
  history.value = dto.history ?? [];
  loggedIn.value = dto.loggedIn;
  if (!dto.loggedIn || options.skipSnapshotUpdate) {
    return;
  }
  const status = normalizeStatus(dto.status);
  const description = dto.statusDescription ?? "";
  if (options.isInitial || !isDirty.value) {
    draftStatus.value = status;
    draftDescription.value = description;
    snapshotStatus.value = status;
    snapshotDescription.value = description;
  }
}

function normalizeStatus(status: string): PresenceStatus {
  switch (status) {
    case "join me":
    case "active":
    case "ask me":
    case "busy":
      return status;
    default:
      console.warn(
        "PresenceChangeSection: unexpected status, falling back to active:",
        status,
      );
      return "active";
  }
}

type LoadOptions = {
  onlyIfNotDirty?: boolean;
  skipSnapshotUpdate?: boolean;
};

async function load(options?: LoadOptions): Promise<void> {
  if (options?.onlyIfNotDirty && isDirty.value) {
    return;
  }
  if (inFlight) {
    pendingRefresh = true;
    pendingRefreshOnlyIfNotDirty =
      pendingRefreshOnlyIfNotDirty || Boolean(options?.onlyIfNotDirty);
    return;
  }
  inFlight = true;
  pendingRefresh = false;
  const onlyIfNotDirty = Boolean(options?.onlyIfNotDirty);
  const skipSnapshotUpdate = Boolean(options?.skipSnapshotUpdate);
  const gen = generation;
  try {
    const dto = await App.getPresenceChangeSection();
    if (gen !== generation) return;
    if (onlyIfNotDirty && isDirty.value) return;
    loadError.value = false;
    applySection(dto, {
      isInitial: !hasLoadedOnce,
      skipSnapshotUpdate,
    });
    hasLoadedOnce = true;
  } catch (e) {
    if (gen !== generation) return;
    console.error("PresenceChangeSection load failed:", e);
    if (!hasLoadedOnce) {
      loadError.value = true;
      loggedIn.value = false;
      history.value = [];
    }
  } finally {
    inFlight = false;
    if (gen === generation) {
      loading.value = false;
    }
    if (pendingRefresh && gen === generation) {
      const nextOnlyIfNotDirty = pendingRefreshOnlyIfNotDirty;
      pendingRefresh = false;
      pendingRefreshOnlyIfNotDirty = false;
      void load({ onlyIfNotDirty: nextOnlyIfNotDirty });
    }
  }
}

async function retryLoad(): Promise<void> {
  loading.value = true;
  loadError.value = false;
  await load();
}

function scheduleSelfCacheRefresh(): void {
  if (selfCacheDebounceTimer !== null) {
    clearTimeout(selfCacheDebounceTimer);
  }
  selfCacheDebounceTimer = setTimeout(() => {
    selfCacheDebounceTimer = null;
    void load({ onlyIfNotDirty: true });
  }, SELF_CACHE_CHANGED_DEBOUNCE_MS);
}

function selectColor(status: PresenceStatus): void {
  draftStatus.value = status;
}

function queryHistory(
  queryString: string,
  cb: (results: Array<{ value: string }>) => void,
): void {
  const q = queryString.trim().toLowerCase();
  const items = history.value
    .filter((item) => !q || item.toLowerCase().includes(q))
    .map((value) => ({ value }));
  cb(items);
}

function onDescriptionSelect(item: { value: string }): void {
  draftDescription.value = item.value;
}

function syncSnapshotFromApply(status: string, description: string): void {
  const normalized = normalizeStatus(status);
  draftStatus.value = normalized;
  draftDescription.value = description;
  snapshotStatus.value = normalized;
  snapshotDescription.value = description;
}

async function applyPresenceChange(): Promise<void> {
  if (!loggedIn.value || !isDirty.value || applying.value) return;
  applying.value = true;
  try {
    const result = await App.applyPresenceChange(
      draftStatus.value,
      draftDescription.value,
    );
    syncSnapshotFromApply(result.status, result.statusDescription);
    ElMessage.success(t("dashboard.presenceChange.applySuccess"));
    void load({ skipSnapshotUpdate: true });
  } catch (e) {
    ElMessage.error(formatError(e, t("dashboard.presenceChange.applyError")));
  } finally {
    applying.value = false;
  }
}

onMounted(async () => {
  await load();
  const rt = getRuntime();
  const off = rt?.EventsOn?.("identity:self-cache-changed", () => {
    scheduleSelfCacheRefresh();
  });
  if (typeof off === "function") {
    unsubscribeSelfCacheChanged = off;
  }
});

watch(unlockState, (state) => {
  if (state === "unlocked" && loadError.value) {
    void retryLoad();
  }
});

onUnmounted(() => {
  generation += 1;
  if (selfCacheDebounceTimer !== null) {
    clearTimeout(selfCacheDebounceTimer);
    selfCacheDebounceTimer = null;
  }
  unsubscribeSelfCacheChanged?.();
});
</script>

<style scoped>
.presence-change-section {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.presence-change-title {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.presence-change-message,
.presence-change-login-hint {
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin: 0 0 0.75rem;
}

.presence-change-error-text {
  margin: 0 0 0.5rem;
}

.presence-change-settings-link {
  margin-left: 0.35rem;
}

.presence-color-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
}

.presence-color-buttons :deep(.el-button + .el-button) {
  margin-left: 0;
}

.presence-color-btn {
  border: 1px solid transparent;
  color: #fff !important;
  flex: 0 0 auto;
}

.presence-color-btn:hover,
.presence-color-btn:focus {
  filter: brightness(1.08);
  color: #fff !important;
}

.presence-color-btn--selected {
  outline: 2px solid var(--text-primary);
  outline-offset: 2px;
}

.presence-color-btn--join-me {
  background: #2b7fd9 !important;
  border-color: #256bb8 !important;
}

.presence-color-btn--active {
  background: #2e9f4a !important;
  border-color: #267d3c !important;
}

.presence-color-btn--ask-me {
  background: #e8943c !important;
  border-color: #c97d2e !important;
}

.presence-color-btn--busy {
  background: #d94a4a !important;
  border-color: #b83c3c !important;
}

.presence-description-row {
  display: flex;
  flex-wrap: nowrap;
  gap: 0.5rem;
  align-items: center;
}

.presence-description-row :deep(.el-button + .el-button) {
  margin-left: 0;
}

.presence-description-input {
  flex: 1 1 auto;
  min-width: 0;
  width: 100%;
}

.presence-description-input :deep(.el-input) {
  width: 100%;
}

.presence-apply-btn {
  flex: 0 0 auto;
}
</style>
