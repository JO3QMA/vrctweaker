<script setup lang="ts">
import {
  isTwoFactorCodeComplete,
  sanitizeTwoFactorCodeInput,
  sanitizeTwoFactorCodePaste,
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
  if (isTwoFactorCodeComplete(next)) {
    emit("complete");
  }
}

function onPaste(event: ClipboardEvent): void {
  event.preventDefault();
  const text = event.clipboardData?.getData("text") ?? "";
  commitValue(sanitizeTwoFactorCodePaste(text));
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
