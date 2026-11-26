<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, shallowRef } from "vue";
import { syncLicensesTableScrollFade } from "./licensesTableScrollFade";

const props = defineProps<{
  scrollTestId: string;
  fadeTestId: string;
}>();

const scrollEl = shallowRef<HTMLElement | null>(null);
const fadeVisible = ref(false);
let resizeObserver: ResizeObserver | undefined;

function updateFade() {
  fadeVisible.value = syncLicensesTableScrollFade(scrollEl.value);
}

onMounted(async () => {
  await nextTick();
  updateFade();
  const el = scrollEl.value;
  if (!el) {
    return;
  }
  if (typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(() => updateFade());
    resizeObserver.observe(el);
    const table = el.querySelector("table");
    if (table) {
      resizeObserver.observe(table);
    }
  }
  window.addEventListener("resize", updateFade);
});

onUnmounted(() => {
  resizeObserver?.disconnect();
  resizeObserver = undefined;
  window.removeEventListener("resize", updateFade);
});
</script>

<template>
  <div class="licenses-table-scroll-host">
    <div
      ref="scrollEl"
      class="licenses-table-scroll"
      :data-testid="props.scrollTestId"
      @scroll="updateFade"
    >
      <slot />
    </div>
    <div
      class="licenses-table-scroll-fade"
      :class="{ 'licenses-table-scroll-fade--visible': fadeVisible }"
      :data-testid="props.fadeTestId"
      aria-hidden="true"
    />
  </div>
</template>

<style scoped>
.licenses-table-scroll-host {
  position: relative;
  width: 100%;
  max-width: 100%;
}

.licenses-table-scroll {
  width: 100%;
  max-width: 100%;
  overflow-x: auto;
}

.licenses-table-scroll-fade {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: var(--space-32);
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.15s ease;
  background: linear-gradient(
    to right,
    transparent 0%,
    var(--color-bg-base) 100%
  );
}

.licenses-table-scroll-fade--visible {
  opacity: 1;
}
</style>
