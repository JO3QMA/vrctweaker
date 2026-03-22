<template>
  <div class="encounter-history-view">
    <h1 class="page-title">{{ pageTitle }}</h1>
    <p v-if="idLine" class="id-line">{{ idLine }}</p>

    <div v-if="invalidQuery" class="message message--warn">
      表示できません。URL の kind / vrcUserId / worldId を確認してください。
    </div>
    <div v-else-if="loading" class="message">読み込み中…</div>
    <div v-else-if="error" class="message message--warn">{{ error }}</div>
    <div v-else-if="rows.length === 0" class="message">
      該当する遭遇ログがありません。
    </div>
    <div v-else class="table-wrap">
      <table class="history-table">
        <thead>
          <tr>
            <th>時刻</th>
            <th>アクション</th>
            <th>表示名</th>
            <th>ワールド名</th>
            <th class="col-instance">インスタンス</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.id">
            <td>{{ formatEncounteredAt(row.encounteredAt) }}</td>
            <td>
              <span :class="['action', row.action]">{{
                actionLabel(row.action)
              }}</span>
            </td>
            <td>{{ row.displayName }}</td>
            <td :title="row.worldId || ''">
              {{ row.worldDisplayName || row.worldId || "—" }}
            </td>
            <td class="col-instance mono">{{ row.instanceId || "—" }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useRoute } from "vue-router";
import { App, type UserEncounterDTO } from "../wails/app";

const route = useRoute();

const loading = ref(false);
const error = ref<string | null>(null);
const rows = ref<UserEncounterDTO[]>([]);

function firstQueryString(v: unknown): string {
  if (v == null) return "";
  if (typeof v === "string") return v;
  if (Array.isArray(v)) {
    for (const x of v) {
      if (typeof x === "string") return x;
    }
  }
  return "";
}

const kind = computed(() => firstQueryString(route.query.kind));

const vrcUserId = computed(() => firstQueryString(route.query.vrcUserId));

const worldId = computed(() => firstQueryString(route.query.worldId));

const invalidQuery = computed(() => {
  if (kind.value !== "user" && kind.value !== "world") {
    return true;
  }
  if (kind.value === "user" && !vrcUserId.value.trim()) {
    return true;
  }
  if (kind.value === "world" && !worldId.value.trim()) {
    return true;
  }
  return false;
});

const pageTitle = computed(() => {
  if (kind.value === "user") return "ユーザー別 遭遇履歴";
  if (kind.value === "world") return "ワールド別 遭遇履歴";
  return "遭遇履歴";
});

const idLine = computed(() => {
  if (invalidQuery.value) return "";
  if (kind.value === "user") return vrcUserId.value;
  if (kind.value === "world") return worldId.value;
  return "";
});

function formatEncounteredAt(iso: string): string {
  try {
    return new Date(iso).toLocaleString("ja-JP");
  } catch {
    return iso;
  }
}

function actionLabel(action: string): string {
  if (action === "join") return "参加";
  if (action === "leave") return "退出";
  return action;
}

async function load(): Promise<void> {
  if (invalidQuery.value) {
    rows.value = [];
    error.value = null;
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    if (kind.value === "user") {
      rows.value = await App.encountersByVRCUserID(vrcUserId.value);
    } else {
      rows.value = await App.encountersByWorldID(worldId.value);
    }
  } catch (e) {
    rows.value = [];
    error.value =
      e instanceof Error ? e.message : "データの取得に失敗しました。";
  } finally {
    loading.value = false;
  }
}

watch(
  () => ({
    k: kind.value,
    u: vrcUserId.value,
    w: worldId.value,
  }),
  () => {
    void load();
  },
);

onMounted(() => {
  void load();
});
</script>

<style scoped>
.encounter-history-view {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 0;
}

.page-title {
  margin: 0;
  font-size: 1.25rem;
}

.id-line {
  margin: 0;
  font-size: 0.8rem;
  color: var(--text-secondary);
  word-break: break-all;
}

.message {
  padding: 1.5rem;
  text-align: center;
  color: var(--text-secondary);
}

.message--warn {
  color: var(--text-primary);
}

.table-wrap {
  overflow-x: auto;
  border: 1px solid var(--border);
  border-radius: var(--radius);
}

.history-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.history-table th,
.history-table td {
  padding: 0.45rem 0.6rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

.history-table th {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-weight: 600;
}

.history-table tbody tr:last-child td {
  border-bottom: none;
}

.action.join {
  color: var(--success);
}

.action.leave {
  color: var(--text-secondary);
}

.col-instance {
  max-width: 12rem;
}

.mono {
  font-family: monospace;
  font-size: 0.78rem;
  word-break: break-all;
}
</style>
