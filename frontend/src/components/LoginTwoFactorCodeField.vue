<script setup lang="ts">
import {
  isSanitizedTwoFactorCodeComplete,
  sanitizeTwoFactorCodeInput,
  TWO_FACTOR_CODE_LENGTH,
} from "@/utils/twoFactorCodeInput";

const props = defineProps<{
  modelValue: string;
  disabled?: boolean;
  inputId?: string;
  dataTestid?: string;
  placeholder?: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
  complete: [];
}>();

function commitValue(raw: string): void {
  const next = sanitizeTwoFactorCodeInput(raw);
  emit("update:modelValue", next);
  if (isSanitizedTwoFactorCodeComplete(next)) {
    emit("complete");
  }
}

function onPaste(event: ClipboardEvent): void {
  event.preventDefault();
  commitValue(event.clipboardData?.getData("text") ?? "");
}
</script>

<template>
  <el-input
    :id="inputId"
    :data-testid="dataTestid"
    :model-value="props.modelValue"
    :disabled="disabled"
    :placeholder="placeholder"
    inputmode="numeric"
    :maxlength="TWO_FACTOR_CODE_LENGTH"
    autocomplete="one-time-code"
    @update:model-value="commitValue"
    @paste="onPaste"
  />
</template>
