<template>
  <a
    v-if="safe"
    href="#"
    class="safe-external-url-link"
    :class="linkClass"
    :title="title"
    :aria-label="ariaLabel"
    :data-testid="testId"
    :data-external-url="safeHref"
    @click.prevent="onOpen"
  >
    <slot>{{ display }}</slot>
  </a>
  <span
    v-else
    :class="textClass"
    :data-testid="testId ? `${testId}-text` : undefined"
  >
    <slot>{{ display }}</slot>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { openExternalHttpUrl } from "@/utils/openExternalHttpUrl";
import { formatError } from "@/utils/formatError";
import { showToast } from "@/utils/showToast";
import { parseSafeHttpUrl } from "@/utils/safeHttpUrl";

const { t } = useI18n();

const props = withDefaults(
  defineProps<{
    url: string;
    /** Visible label; defaults to url. */
    label?: string;
    title?: string;
    ariaLabel?: string;
    linkClass?: string;
    textClass?: string;
    testId?: string;
  }>(),
  {
    label: undefined,
    title: undefined,
    ariaLabel: undefined,
    linkClass: "",
    textClass: "wrap",
    testId: undefined,
  },
);

const display = computed(() => props.label ?? props.url);
const parsed = computed(() => parseSafeHttpUrl(props.url));
const safe = computed(() => parsed.value.ok);
const safeHref = computed(() =>
  parsed.value.ok ? parsed.value.href : undefined,
);

async function onOpen() {
  try {
    const opened = await openExternalHttpUrl(props.url);
    if (!opened) {
      showToast.warning(t("userDetail.openExternalUrlError"));
    }
  } catch (e: unknown) {
    showToast.error(formatError(e, t("userDetail.openExternalUrlError")));
  }
}
</script>

<style scoped>
.safe-external-url-link {
  color: var(--color-brand);
  text-decoration: underline;
  cursor: pointer;
}
</style>
