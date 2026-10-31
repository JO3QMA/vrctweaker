<template>
  <el-card
    class="section-card presence-change-section"
    shadow="never"
    data-testid="presence-change-section"
  >
    <template #header>
      <span class="presence-change-title text-body-sm">{{
        t("dashboard.presenceChange.title")
      }}</span>
    </template>

    <div
      v-if="loading"
      class="presence-change-message text-body-sm"
      data-testid="presence-change-loading"
    >
      {{ t("dashboard.presenceChange.loading") }}
    </div>

    <div
      v-else-if="loadError"
      class="presence-change-message text-body-sm"
      data-testid="presence-change-load-error"
    >
      <p class="presence-change-error-text">
        {{ t("dashboard.presenceChange.loadError") }}
      </p>
      <VtButton
        variant="secondary"
        size="small"
        data-testid="presence-change-retry"
        @click="retryLoad"
      >
        {{ t("dashboard.presenceChange.retry") }}
      </VtButton>
    </div>

    <template v-else>
      <p
        v-if="!loggedIn"
        class="presence-change-login-hint text-body-sm"
        data-testid="presence-change-login-required"
      >
        {{ t("dashboard.presenceChange.loginRequired") }}
        <router-link
          to="/settings/account"
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
        <VtButton
          variant="primary"
          class="presence-apply-btn"
          data-testid="presence-change-apply"
          :disabled="!loggedIn || !isDirty || applying"
          :loading="applying"
          @click="applyPresenceChange"
        >
          {{ t("dashboard.presenceChange.apply") }}
        </VtButton>
      </div>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  App,
  type PresenceChangeSectionDTO,
  type PresenceStatus,
} from "../wails/app";
import { getRuntime } from "../wails/runtime";
import { formatError } from "../utils/formatError";
import { showToast } from "../utils/showToast";
import { useSessionUnlock } from "../composables/useSessionUnlock";
import { useAsyncSectionLoad } from "../composables/useAsyncSectionLoad";
import VtButton from "./VtButton.vue";

const SELF_CACHE_CHANGED_DEBOUNCE_MS = 300;

const { t } = useI18n();
const { state: unlockState } = useSessionUnlock();

const loadError = ref(false);
const applying = ref(false);
const loggedIn = ref(false);
const history = ref<string[]>([]);
const draftStatus = ref<PresenceStatus>("active");
const draftDescription = ref("");
const snapshotStatus = ref<PresenceStatus>("active");
const snapshotDescription = ref("");

type LoadOptions = {
  onlyIfNotDirty?: boolean;
  skipSnapshotUpdate?: boolean;
};

let selfCacheDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribeSelfCacheChanged: (() => void) | undefined;
// Load mutex + pending queue + generation は useAsyncSectionLoad に集約。
// pendingLoadOptions は onlyIfNotDirty / skipSnapshotUpdate を OR マージする。
function mergePendingLoadOptions(
  queued: LoadOptions | undefined,
  incoming: LoadOptions | undefined,
): LoadOptions {
  return {
    onlyIfNotDirty:
      Boolean(queued?.onlyIfNotDirty) || Boolean(incoming?.onlyIfNotDirty),
    skipSnapshotUpdate:
      Boolean(queued?.skipSnapshotUpdate) ||
      Boolean(incoming?.skipSnapshotUpdate),
  };
}

const { loading, hasLoadedOnce, begin, end, isStale, markLoaded } =
  useAsyncSectionLoad<LoadOptions>({
    mergeOptions: mergePendingLoadOptions,
  });

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
  // Go normalizes the self status to one of the four values. Keep this single
  // guard only for unexpected external data; apply results are trusted below.
  let status: PresenceStatus;
  if (isPresenceStatus(dto.status)) {
    status = dto.status;
  } else {
    console.warn(
      "PresenceChangeSection: unexpected status, falling back to active:",
      dto.status,
    );
    status = "active";
  }
  const description = dto.statusDescription ?? "";
  if (options.isInitial || !isDirty.value) {
    draftStatus.value = status;
    draftDescription.value = description;
    snapshotStatus.value = status;
    snapshotDescription.value = description;
  }
}

function isPresenceStatus(status: string): status is PresenceStatus {
  switch (status) {
    case "join me":
    case "active":
    case "ask me":
    case "busy":
      return true;
    default:
      return false;
  }
}

async function load(options?: LoadOptions): Promise<void> {
  if (options?.onlyIfNotDirty && isDirty.value) {
    return;
  }
  const gen = begin(options);
  if (gen === null) {
    return;
  }
  const onlyIfNotDirty = Boolean(options?.onlyIfNotDirty);
  const skipSnapshotUpdate = Boolean(options?.skipSnapshotUpdate);
  try {
    const dto = await App.getPresenceChangeSection();
    if (isStale(gen)) return;
    if (onlyIfNotDirty && isDirty.value) return;
    loadError.value = false;
    applySection(dto, {
      isInitial: !hasLoadedOnce.value,
      skipSnapshotUpdate,
    });
    markLoaded();
  } catch (e) {
    if (isStale(gen)) return;
    console.error("PresenceChangeSection load failed:", e);
    if (!hasLoadedOnce.value) {
      loadError.value = true;
      loggedIn.value = false;
      history.value = [];
    } else {
      showToast.warning(t("dashboard.presenceChange.refreshError"));
    }
  } finally {
    const next = end(gen);
    if (next.replay) {
      void load(next.options);
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

function syncSnapshotFromApply(
  status: PresenceStatus,
  description: string,
): void {
  draftStatus.value = status;
  draftDescription.value = description;
  snapshotStatus.value = status;
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
    showToast.success(t("dashboard.presenceChange.applySuccess"));
    void load({ skipSnapshotUpdate: true });
  } catch (e) {
    showToast.error(formatError(e, t("dashboard.presenceChange.applyError")));
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
  if (selfCacheDebounceTimer !== null) {
    clearTimeout(selfCacheDebounceTimer);
    selfCacheDebounceTimer = null;
  }
  unsubscribeSelfCacheChanged?.();
});
</script>

<style scoped>
.presence-change-title {
  color: var(--color-text-secondary);
}

.presence-change-message,
.presence-change-login-hint {
  color: var(--color-text-secondary);
  margin: 0 0 var(--space-form-field);
}

.presence-change-error-text {
  margin: 0 0 var(--space-action-group);
}

.presence-change-settings-link {
  margin-left: var(--space-inline-tight);
}

.presence-color-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-inline-tight);
  margin-bottom: var(--space-action-group);
}

.presence-color-buttons :deep(.el-button + .el-button) {
  margin-left: 0;
}

.presence-color-btn {
  border: 1px solid transparent;
  color: var(--color-text-inverse) !important;
  flex: 0 0 auto;
}

.presence-color-btn:hover,
.presence-color-btn:focus {
  filter: brightness(1.08);
  color: var(--color-text-inverse) !important;
}

.presence-color-btn--selected {
  outline: 2px solid var(--color-text-primary);
  outline-offset: 2px;
}

.presence-color-btn--join-me {
  background: var(--color-presence-join-me) !important;
  border-color: var(--color-presence-join-me-border) !important;
}

.presence-color-btn--active {
  background: var(--color-presence-active) !important;
  border-color: var(--color-presence-active-border) !important;
}

.presence-color-btn--ask-me {
  background: var(--color-presence-ask-me) !important;
  border-color: var(--color-presence-ask-me-border) !important;
}

.presence-color-btn--busy {
  background: var(--color-presence-busy) !important;
  border-color: var(--color-presence-busy-border) !important;
}

.presence-description-row {
  display: flex;
  flex-wrap: nowrap;
  gap: var(--space-action-group);
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
