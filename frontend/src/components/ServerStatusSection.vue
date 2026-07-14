<template>
  <el-card
    class="server-status-panel"
    shadow="never"
    data-testid="server-status-section"
  >
    <template #header>
      <span class="server-status-title">{{
        t("dashboard.serverStatus.title")
      }}</span>
    </template>

    <div
      v-if="loading"
      class="server-status-message"
      data-testid="server-status-loading"
    >
      {{ t("dashboard.serverStatus.loading") }}
    </div>

    <div v-else-if="fetchState === 'unavailable'" class="server-status-message">
      {{ t("dashboard.serverStatus.fetchUnavailable") }}
    </div>

    <template v-else>
      <div
        class="server-status-summary"
        :class="summaryColorClass"
        data-testid="server-status-summary"
      >
        <span class="server-status-dot" aria-hidden="true" />
        <span>{{ summaryLabel }}</span>
      </div>

      <p
        v-if="fetchState === 'partial'"
        class="server-status-message"
        data-testid="server-status-detail-unavailable"
      >
        {{ t("dashboard.serverStatus.detailUnavailable") }}
      </p>

      <ul
        v-else-if="showDetail"
        class="server-status-detail"
        data-testid="server-status-detail"
      >
        <li
          v-for="(row, idx) in componentRows"
          :key="`component-${idx}`"
          class="server-status-detail-row"
          :class="row.colorClass"
        >
          <span class="server-status-dot" aria-hidden="true" />
          <span class="server-status-detail-name">{{ row.name }}</span>
          <span class="server-status-detail-status">{{ row.statusLabel }}</span>
        </li>
        <li
          v-for="(headline, idx) in headlineRows"
          :key="`headline-${idx}`"
          class="server-status-detail-row server-status-detail-headline"
        >
          {{ headline }}
        </li>
      </ul>

      <a
        class="server-status-link"
        :href="statusPageUrl"
        target="_blank"
        rel="noopener noreferrer"
        data-testid="server-status-external-link"
      >
        {{ t("dashboard.serverStatus.linkToStatusPage") }}
      </a>
    </template>
  </el-card>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { App, type ServerStatusDTO } from "../wails/app";
import {
  SERVER_STATUS_PAGE_URL,
  SERVER_STATUS_POLL_MS,
  serverStatusComponentColorClass,
  serverStatusComponentI18nKey,
  serverStatusSummaryColorClass,
  serverStatusSummaryI18nKey,
} from "../utils/serverStatus";

const { t } = useI18n();

const statusPageUrl = SERVER_STATUS_PAGE_URL;
const loading = ref(true);
const fetchState = ref<ServerStatusDTO["fetchState"]>("unavailable");
const summaryIndicator = ref("");
const components = ref<ServerStatusDTO["components"]>([]);
const incidents = ref<ServerStatusDTO["incidents"]>([]);
const maintenances = ref<ServerStatusDTO["maintenances"]>([]);

let pollTimer: ReturnType<typeof setInterval> | null = null;
let inFlight = false;
let generation = 0;

const summaryColorClass = computed(() =>
  serverStatusSummaryColorClass(summaryIndicator.value),
);

const summaryLabel = computed(() =>
  t(serverStatusSummaryI18nKey(summaryIndicator.value)),
);

const showDetail = computed(
  () =>
    fetchState.value === "ok" &&
    (components.value.length > 0 ||
      incidents.value.length > 0 ||
      maintenances.value.length > 0),
);

const componentRows = computed(() =>
  components.value.map((c) => ({
    name: c.name,
    statusLabel: t(serverStatusComponentI18nKey(c.status)),
    colorClass: serverStatusComponentColorClass(c.status),
  })),
);

const headlineRows = computed(() => [
  ...incidents.value.map((h) => h.name).filter(Boolean),
  ...maintenances.value.map((h) => h.name).filter(Boolean),
]);

function applySnapshot(dto: ServerStatusDTO): void {
  fetchState.value = dto.fetchState;
  summaryIndicator.value = dto.summary?.indicator ?? "";
  components.value = dto.components ?? [];
  incidents.value = dto.incidents ?? [];
  maintenances.value = dto.maintenances ?? [];
}

async function refresh(): Promise<void> {
  if (inFlight) return;
  inFlight = true;
  const gen = generation;
  try {
    const dto = await App.getServerStatus();
    if (gen !== generation) return;
    applySnapshot(dto);
  } catch {
    if (gen !== generation) return;
    if (loading.value) {
      fetchState.value = "unavailable";
    }
  } finally {
    inFlight = false;
    if (gen === generation) {
      loading.value = false;
    }
  }
}

onMounted(() => {
  void refresh();
  pollTimer = setInterval(() => {
    void refresh();
  }, SERVER_STATUS_POLL_MS);
});

onUnmounted(() => {
  generation += 1;
  if (pollTimer !== null) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
});
</script>

<style scoped>
.server-status-panel {
  background: var(--bg-secondary) !important;
  border-color: var(--border) !important;
}

.server-status-title {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.server-status-summary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.server-status-message {
  margin: 0 0 0.5rem;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.server-status-detail {
  list-style: none;
  margin: 0 0 0.75rem;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.server-status-detail-row {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  flex-wrap: wrap;
  font-size: 0.9rem;
}

.server-status-detail-name {
  flex: 1 1 auto;
  min-width: 0;
}

.server-status-detail-status {
  color: var(--text-secondary);
  white-space: nowrap;
}

.server-status-detail-headline {
  padding-left: 1.1rem;
  color: var(--text-secondary);
}

.server-status-dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  flex-shrink: 0;
  background: currentColor;
}

.server-status-link {
  font-size: 0.85rem;
  color: var(--el-color-primary);
}

.server-status--operational {
  color: #2e9f4a;
}

.server-status--degraded {
  color: #d4a017;
}

.server-status--partial {
  color: #e8943c;
}

.server-status--major {
  color: #d94a4a;
}

.server-status--maintenance {
  color: #2b7fd9;
}

.server-status--unknown {
  color: var(--text-secondary);
}
</style>
