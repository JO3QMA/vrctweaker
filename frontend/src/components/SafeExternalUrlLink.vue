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
import { openExternalHttpUrl } from "@/utils/openExternalHttpUrl";
import { parseSafeHttpUrl } from "@/utils/safeHttpUrl";

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
  await openExternalHttpUrl(props.url);
}
</script>

<style scoped>
.safe-external-url-link {
  color: var(--color-brand);
  text-decoration: underline;
  cursor: pointer;
}
</style>
