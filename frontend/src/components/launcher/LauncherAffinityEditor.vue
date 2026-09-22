<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import VtAlert from "../VtAlert.vue";
import VtButton from "../VtButton.vue";
import VtCheckbox from "../VtCheckbox.vue";
import VtInput from "../VtInput.vue";
import {
  coreStatesFromMask,
  DEFAULT_LOGICAL_PROCESSOR_COUNT,
  formatAffinityHex,
  hasHiddenBitsAbove,
  hiddenMaskAbove,
  mergeVisibleWithHidden,
  parseAffinityHex,
  affinityVisibleSelectionIssue,
} from "../../utils/affinityMask";
import { resolveLogicalProcessorCount } from "../../utils/affinityProcessorCount";

const props = defineProps<{
  modelValue: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const { t } = useI18n();

const coreCount = ref(DEFAULT_LOGICAL_PROCESSOR_COUNT);
const coreCountReady = ref(false);
const coreStates = ref<boolean[]>(
  coreStatesFromMask(0n, DEFAULT_LOGICAL_PROCESSOR_COUNT),
);
const hiddenMask = ref(0n);
const parseFailed = ref(false);
const overflowDismissed = ref(false);
const hexExpertOpen = ref(false);
const hexDraft = ref("");
const hexFieldError = ref("");

const fullMask = computed(() =>
  mergeVisibleWithHidden(coreStates.value, hiddenMask.value),
);

const showOverflowAlert = computed(
  () =>
    !overflowDismissed.value &&
    hasHiddenBitsAbove(fullMask.value, coreCount.value),
);

const validationError = computed(() => {
  if (parseFailed.value) {
    return t("launcher.affinityErrInvalid");
  }
  const issue = affinityVisibleSelectionIssue(fullMask.value, coreCount.value);
  if (issue === "hiddenOnly") {
    return t("launcher.affinityHiddenOnlyWarning");
  }
  if (issue === "empty") {
    return t("launcher.affinityErrEmpty");
  }
  return "";
});

function syncFromModelValue(raw: string) {
  parseFailed.value = false;
  hexFieldError.value = "";
  const parsed = parseAffinityHex(raw);
  if (!parsed.ok) {
    parseFailed.value = true;
    hexDraft.value = raw;
    coreStates.value = coreStatesFromMask(0n, coreCount.value);
    hiddenMask.value = 0n;
    return;
  }
  const prevHidden = hiddenMask.value;
  const nextHidden = hiddenMaskAbove(parsed.mask, coreCount.value);
  if (nextHidden !== prevHidden) {
    overflowDismissed.value = false;
  }
  hiddenMask.value = nextHidden;
  coreStates.value = coreStatesFromMask(parsed.mask, coreCount.value);
  hexDraft.value = formatAffinityHex(parsed.mask);
}

function emitMask(mask: bigint) {
  emit("update:modelValue", formatAffinityHex(mask));
}

function applyCoreStates(states: boolean[]) {
  coreStates.value = states;
  emitMask(mergeVisibleWithHidden(states, hiddenMask.value));
}

function setCoreAllowed(index: number, allowed: boolean) {
  const next = [...coreStates.value];
  next[index] = allowed;
  applyCoreStates(next);
}

function allowAllCores() {
  applyCoreStates(Array.from({ length: coreCount.value }, () => true));
}

function suppressAllVisible() {
  applyCoreStates(Array.from({ length: coreCount.value }, () => false));
}

function commitHexDraft() {
  hexFieldError.value = "";
  const parsed = parseAffinityHex(hexDraft.value);
  if (!parsed.ok) {
    hexFieldError.value = t("launcher.affinityErrInvalid");
    return;
  }
  parseFailed.value = false;
  const nextHidden = hiddenMaskAbove(parsed.mask, coreCount.value);
  if (nextHidden !== hiddenMask.value) {
    overflowDismissed.value = false;
  }
  hiddenMask.value = nextHidden;
  coreStates.value = coreStatesFromMask(parsed.mask, coreCount.value);
  hexDraft.value = formatAffinityHex(parsed.mask);
  emitMask(parsed.mask);
}

function onHexDraftKeydown(e: KeyboardEvent) {
  if (e.key === "Enter") {
    commitHexDraft();
  }
}

async function loadCoreCount() {
  coreCount.value = await resolveLogicalProcessorCount();
  coreStates.value = coreStatesFromMask(0n, coreCount.value);
}

watch(
  () => props.modelValue,
  (v) => {
    if (!coreCountReady.value) return;
    syncFromModelValue(v);
  },
);

onMounted(async () => {
  await loadCoreCount();
  coreCountReady.value = true;
  syncFromModelValue(props.modelValue);
});
</script>

<template>
  <div class="launcher-affinity-editor" data-testid="affinity-editor">
    <p class="text-body-sm affinity-help">
      {{ t("launcher.affinityHelp") }}
    </p>

    <VtAlert
      v-if="parseFailed"
      variant="warning"
      :title="t('launcher.affinityErrInvalid')"
      class="affinity-alert"
      data-testid="affinity-parse-alert"
    />

    <VtAlert
      v-if="showOverflowAlert"
      variant="info"
      :title="t('launcher.affinityHiddenBits')"
      class="affinity-alert"
      data-testid="affinity-overflow-alert"
    >
      <VtButton
        variant="tertiary"
        size="small"
        data-testid="affinity-overflow-dismiss"
        @click="overflowDismissed = true"
      >
        {{ t("launcher.affinityOverflowDismiss") }}
      </VtButton>
    </VtAlert>

    <div class="affinity-bulk">
      <VtButton
        variant="secondary"
        size="small"
        data-testid="affinity-allow-all"
        @click="allowAllCores"
      >
        {{ t("launcher.affinityAllowAll") }}
      </VtButton>
      <VtButton
        variant="secondary"
        size="small"
        data-testid="affinity-suppress-all"
        @click="suppressAllVisible"
      >
        {{ t("launcher.affinitySuppressAll") }}
      </VtButton>
    </div>

    <div
      class="affinity-core-grid"
      role="group"
      :aria-label="t('launcher.affinityGridAria')"
    >
      <template v-for="index in coreCount" :key="`affinity-core-${index - 1}`">
        <div
          v-if="index === 17"
          class="affinity-ccd-divider"
          role="separator"
          :aria-label="t('launcher.affinityCcdDivider')"
        />
        <VtCheckbox
          class="affinity-core-checkbox"
          :model-value="coreStates[index - 1]"
          size="small"
          :data-testid="`affinity-core-${index - 1}`"
          :aria-label="t('launcher.affinityCoreAria', { n: index - 1 })"
          @update:model-value="setCoreAllowed(index - 1, $event)"
        >
          C{{ index - 1 }}
        </VtCheckbox>
      </template>
    </div>

    <p
      v-if="validationError && !parseFailed"
      class="text-body-sm affinity-validation"
      data-testid="affinity-validation"
    >
      {{ validationError }}
    </p>

    <button
      type="button"
      class="affinity-hex-toggle text-body-sm"
      data-testid="affinity-hex-toggle"
      @click="hexExpertOpen = !hexExpertOpen"
    >
      {{ t("launcher.affinityHexEdit") }}
    </button>

    <div v-if="hexExpertOpen" class="affinity-hex-panel">
      <VtInput
        v-model="hexDraft"
        :placeholder="t('launcher.affinityPh')"
        data-testid="affinity-hex-input"
        size="small"
        @blur="commitHexDraft"
        @keydown="onHexDraftKeydown"
      />
      <p v-if="hexFieldError" class="text-body-sm affinity-validation">
        {{ hexFieldError }}
      </p>
    </div>
  </div>
</template>

<style scoped>
.affinity-help {
  margin: 0 0 var(--space-block);
  color: var(--color-text-secondary);
}

.affinity-alert {
  margin-bottom: var(--space-block);
}

.affinity-bulk {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-action-group);
  margin-bottom: var(--space-block);
}

.affinity-core-grid {
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: var(--space-inline-tight);
  max-width: 520px;
}

.affinity-ccd-divider {
  grid-column: 1 / -1;
  height: 0;
  border-top: 1px dashed var(--color-border);
  margin: var(--space-inline-tight) 0;
}

.affinity-core-checkbox {
  margin: 0;
  font-size: var(--font-size-12);
}

.affinity-validation {
  margin: var(--space-block) 0 0;
  color: var(--color-danger);
}

.affinity-hex-toggle {
  margin-top: var(--space-block);
  padding: 0;
  border: none;
  background: none;
  color: var(--color-brand);
  cursor: pointer;
  text-decoration: underline;
}

.affinity-hex-panel {
  margin-top: var(--space-form-field);
  max-width: 240px;
}
</style>
