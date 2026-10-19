<template>
  <div class="legal-document-panel" data-testid="legal-document-panel">
    <p class="github-supplement text-caption">
      <SafeExternalUrlLink
        :url="githubUrl"
        test-id="legal-view-on-github"
        class="github-link"
      >
        {{ t("legal.viewOnGitHub") }}
      </SafeExternalUrlLink>
    </p>
    <LegalMarkdownBody :markdown="markdown" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import LegalMarkdownBody from "./LegalMarkdownBody.vue";
import SafeExternalUrlLink from "./SafeExternalUrlLink.vue";
import {
  getLegalDocumentMarkdown,
  type LegalDocKind,
} from "@/data/legalDocuments";
import { privacyPolicyUrl, termsOfUseUrl } from "@/constants/legalUrls";

const props = defineProps<{
  kind: LegalDocKind;
}>();

const { locale, t } = useI18n();

const markdown = computed(() =>
  getLegalDocumentMarkdown(props.kind, locale.value),
);

const githubUrl = computed(() =>
  props.kind === "privacy"
    ? privacyPolicyUrl(locale.value)
    : termsOfUseUrl(locale.value),
);
</script>

<style scoped>
.github-supplement {
  margin: 0 0 var(--space-block);
  color: var(--color-text-secondary);
}

.github-link {
  font-size: inherit;
}
</style>
