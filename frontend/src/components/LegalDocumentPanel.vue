<template>
  <div class="legal-document-panel" data-testid="legal-document-panel">
    <p
      v-if="lastUpdatedLabel"
      class="last-updated text-caption"
      data-testid="legal-last-updated"
    >
      {{ lastUpdatedLabel }}
    </p>
    <LegalMarkdownBody :markdown="bodyMarkdown" />
    <p class="latest-github text-caption">
      <SafeExternalUrlLink
        :url="githubUrl"
        test-id="legal-latest-on-github"
        class="github-link"
      >
        {{ t("legal.latestOnGitHub") }}
      </SafeExternalUrlLink>
    </p>
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
import {
  legalDocLocale,
  privacyPolicyUrl,
  termsOfUseUrl,
} from "@/constants/legalUrls";
import {
  formatLegalLastUpdated,
  parseLegalDocument,
} from "@/utils/parseLegalDocument";

const props = defineProps<{
  kind: LegalDocKind;
}>();

const { locale, t } = useI18n();

const rawMarkdown = computed(() =>
  getLegalDocumentMarkdown(props.kind, locale.value),
);

const parsed = computed(() => parseLegalDocument(rawMarkdown.value));

const bodyMarkdown = computed(() => parsed.value.bodyMarkdown);

const lastUpdatedLabel = computed(() => {
  const iso = parsed.value.lastUpdated;
  if (!iso) {
    return "";
  }
  const docLocale = legalDocLocale(locale.value);
  const date = formatLegalLastUpdated(iso, docLocale);
  return t("legal.lastUpdated", { date }, { locale: docLocale });
});

const githubUrl = computed(() =>
  props.kind === "privacy"
    ? privacyPolicyUrl(locale.value)
    : termsOfUseUrl(locale.value),
);
</script>

<style scoped>
.last-updated {
  margin: 0 0 var(--space-block);
  color: var(--color-text-secondary);
}

.latest-github {
  margin: var(--space-section) 0 0;
  color: var(--color-text-secondary);
}

.github-link {
  font-size: inherit;
}
</style>
