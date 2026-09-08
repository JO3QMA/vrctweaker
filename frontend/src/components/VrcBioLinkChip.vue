<template>
  <a
    class="bio-link-chip"
    :href="href"
    target="_blank"
    rel="noopener noreferrer"
    :title="url"
    :aria-label="ariaLabel"
    data-testid="bio-link-chip"
    :data-site="site"
  >
    <VtIcon size="compact" decorative>
      <BioLinkSiteIcon :site="site" />
    </VtIcon>
    <span class="bio-link-chip-label">{{ label }}</span>
  </a>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import BioLinkSiteIcon from "./BioLinkSiteIcon.vue";
import VtIcon from "./VtIcon.vue";
import {
  bioLinkHostname,
  detectBioLinkSite,
  parseBioLinkUrl,
} from "../utils/bioLinkHost";

const props = defineProps<{
  url: string;
}>();

const { t } = useI18n();

const parsed = computed(() => parseBioLinkUrl(props.url));
const href = computed(() => parsed.value?.href ?? props.url.trim());
const site = computed(() => detectBioLinkSite(props.url));

const label = computed(() => {
  const siteKey = site.value;
  if (siteKey !== "link") {
    return t(`userDetail.bioLinkSites.${siteKey}`);
  }
  return bioLinkHostname(props.url);
});

const ariaLabel = computed(() =>
  t("userDetail.bioLinkOpen", {
    site: label.value,
    url: props.url.trim(),
  }),
);
</script>

<style scoped>
.bio-link-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-inline-tight);
  max-width: 100%;
  padding: 0.2rem 0.55rem;
  border: 1px solid var(--color-border);
  border-radius: 999px;
  background: var(--color-bg-muted);
  color: var(--el-color-primary);
  font-size: var(--font-size-14);
  line-height: var(--line-height-tight);
  text-decoration: none;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease;
}

.bio-link-chip:hover {
  border-color: color-mix(
    in srgb,
    var(--el-color-primary) 45%,
    var(--color-border)
  );
  background: color-mix(
    in srgb,
    var(--el-color-primary) 8%,
    var(--color-bg-muted)
  );
}

.bio-link-chip-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
