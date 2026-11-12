<template>
  <div class="legal-document-view">
    <div class="legal-document-view__content">
      <h1 class="page-title">{{ pageTitle }}</h1>
      <LegalDocumentPanel :kind="kind" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import LegalDocumentPanel from "../components/LegalDocumentPanel.vue";
import type { LegalDocKind } from "@/data/legalDocuments";

const props = defineProps<{
  kind: LegalDocKind;
}>();

const { t } = useI18n();

const pageTitle = computed(() =>
  props.kind === "privacy" ? t("legal.privacyPolicy") : t("legal.termsOfUse"),
);
</script>

<style scoped>
.legal-document-view__content {
  max-width: 52rem;
}
</style>
