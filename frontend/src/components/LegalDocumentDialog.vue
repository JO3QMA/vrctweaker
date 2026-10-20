<template>
  <el-dialog
    v-model="open"
    :title="dialogTitle"
    width="min(720px, 92vw)"
    class="legal-document-dialog"
    data-testid="legal-document-dialog"
    append-to-body
    destroy-on-close
    @closed="onClosed"
  >
    <LegalDocumentPanel v-if="kind" :kind="kind" />
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import LegalDocumentPanel from "./LegalDocumentPanel.vue";
import type { LegalDocKind } from "@/data/legalDocuments";

const props = defineProps<{
  kind: LegalDocKind | null;
}>();

const open = defineModel<boolean>({ required: true });

const { t } = useI18n();

const dialogTitle = computed(() => {
  if (props.kind === "privacy") {
    return t("legal.privacyPolicy");
  }
  if (props.kind === "terms") {
    return t("legal.termsOfUse");
  }
  return "";
});

function onClosed(): void {
  open.value = false;
}
</script>

<style scoped>
.legal-document-dialog :deep(.el-dialog__body) {
  max-height: min(70vh, 640px);
  overflow-y: auto;
}
</style>
