<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('appUpdate.title')"
    width="520px"
    :close-on-click-modal="false"
    data-testid="app-update-dialog"
    @close="onClose"
  >
    <p class="version-line text-body-sm">
      {{
        t("appUpdate.versionLine", {
          current: status?.currentVersion ?? "",
          latest: status?.latestVersion ?? "",
        })
      }}
    </p>
    <p v-if="status?.releaseNotes" class="notes-label text-body-sm">
      {{ t("appUpdate.releaseNotes") }}
    </p>
    <pre
      v-if="status?.releaseNotes"
      class="release-notes"
      data-testid="app-update-notes"
      >{{ status.releaseNotes }}</pre>
    <template #footer>
      <div class="dialog-actions">
        <VtButton variant="tertiary" @click="onLater">
          {{ t("appUpdate.later") }}
        </VtButton>
        <VtButton
          v-if="!status?.autoApplySupported"
          variant="secondary"
          data-testid="app-update-open-release"
          @click="onOpenRelease"
        >
          {{ t("appUpdate.openRelease") }}
        </VtButton>
        <VtButton
          v-if="status?.autoApplySupported"
          variant="primary"
          :loading="applying"
          data-testid="app-update-apply"
          @click="onApply"
        >
          {{ t("appUpdate.apply") }}
        </VtButton>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import VtButton from "./VtButton.vue";
import { useAppUpdateLifecycle } from "../composables/useAppUpdate";
import { showToast } from "../utils/showToast";

const { t } = useI18n();
const {
  dialogVisible,
  status,
  applying,
  dismissLater,
  applyUpdate,
  openReleasePage,
  closeDialog,
} = useAppUpdateLifecycle();

function onLater() {
  void dismissLater();
}

function onClose() {
  closeDialog();
}

async function onApply() {
  try {
    await applyUpdate();
  } catch (e) {
    const msg =
      e instanceof Error && e.message.includes("app update:")
        ? t("appUpdate.errApply")
        : t("appUpdate.errApply");
    showToast.error(msg);
  }
}

async function onOpenRelease() {
  try {
    await openReleasePage();
  } catch {
    showToast.error(t("appUpdate.errOpenRelease"));
  }
}
</script>

<style scoped>
.version-line {
  margin: 0 0 var(--space-block);
}

.notes-label {
  margin: 0 0 var(--space-inline-tight);
  color: var(--color-text-secondary);
}

.release-notes {
  margin: 0;
  max-height: 240px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: var(--font-family-ui);
  font-size: var(--font-size-12);
  line-height: var(--line-height-normal);
  padding: var(--space-form-field);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}

.dialog-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-action-group);
  justify-content: flex-end;
}
</style>
