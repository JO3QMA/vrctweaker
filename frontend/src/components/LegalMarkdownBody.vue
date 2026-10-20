<!-- eslint-disable vue/no-v-html -- HTML is produced only by renderLegalMarkdown (marked + DOMPurify). -->
<template>
  <div
    ref="rootEl"
    class="legal-markdown-body"
    data-testid="legal-markdown-body"
    v-html="html"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, onUpdated, ref, watch } from "vue";
import {
  EXTERNAL_LINK_CLASS,
  renderLegalMarkdown,
} from "@/utils/renderLegalMarkdown";
import { openExternalHttpUrl } from "@/utils/openExternalHttpUrl";

const props = defineProps<{
  markdown: string;
}>();

const rootEl = ref<HTMLElement | null>(null);

const html = computed(() => renderLegalMarkdown(props.markdown));

function decodeExternalHref(encoded: string): string | null {
  try {
    return decodeURIComponent(encoded);
  } catch {
    return null;
  }
}

async function openFromSpan(span: HTMLElement): Promise<void> {
  const encoded = span.getAttribute("data-external-href");
  if (!encoded) {
    return;
  }
  const href = decodeExternalHref(encoded);
  if (!href) {
    return;
  }
  await openExternalHttpUrl(href);
}

function onRootClick(event: MouseEvent): void {
  const target = event.target;
  if (!(target instanceof HTMLElement)) {
    return;
  }
  const link = target.closest<HTMLElement>(`.${EXTERNAL_LINK_CLASS}`);
  if (!link || !rootEl.value?.contains(link)) {
    return;
  }
  event.preventDefault();
  void openFromSpan(link);
}

function onRootKeydown(event: KeyboardEvent): void {
  if (event.key !== "Enter" && event.key !== " ") {
    return;
  }
  const target = event.target;
  if (!(target instanceof HTMLElement)) {
    return;
  }
  if (!target.classList.contains(EXTERNAL_LINK_CLASS)) {
    return;
  }
  event.preventDefault();
  void openFromSpan(target);
}

function bindListeners(): void {
  const el = rootEl.value;
  if (!el) {
    return;
  }
  el.removeEventListener("click", onRootClick);
  el.removeEventListener("keydown", onRootKeydown);
  el.addEventListener("click", onRootClick);
  el.addEventListener("keydown", onRootKeydown);
}

onMounted(bindListeners);
onUpdated(bindListeners);

watch(
  () => props.markdown,
  () => {
    bindListeners();
  },
);
</script>

<style scoped>
.legal-markdown-body {
  color: var(--color-text-primary);
  line-height: var(--line-height-relaxed);
  font-size: var(--font-size-14);
}

.legal-markdown-body :deep(h2) {
  font-size: var(--font-size-18);
  margin: var(--space-24) 0 var(--space-inline-tight);
}

.legal-markdown-body :deep(h3) {
  font-size: var(--font-size-16);
  margin: var(--space-block) 0 var(--space-inline-tight);
}

.legal-markdown-body :deep(p),
.legal-markdown-body :deep(ul),
.legal-markdown-body :deep(ol) {
  margin: 0 0 var(--space-block);
}

.legal-markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: var(--space-block);
  font-size: var(--font-size-13);
}

.legal-markdown-body :deep(th),
.legal-markdown-body :deep(td) {
  border: 1px solid var(--color-border);
  padding: var(--space-inline-tight);
  text-align: left;
  vertical-align: top;
}

.legal-markdown-body :deep(code) {
  font-family: ui-monospace, monospace;
  font-size: 0.92em;
}

.legal-markdown-body :deep(.legal-md-external-link) {
  color: var(--color-brand);
  cursor: pointer;
  text-decoration: underline;
}

.legal-markdown-body :deep(.legal-md-external-link:hover) {
  color: var(--color-brand-hover);
}
</style>
