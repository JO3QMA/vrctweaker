<template>
  <el-tooltip placement="top" popper-class="vrc-user-tag-chip-tooltip">
    <template #content>
      <div v-for="(line, idx) in tooltipLines" :key="idx">{{ line }}</div>
    </template>
    <el-tag
      size="small"
      :type="tagType"
      class="vrc-user-tag-chip"
      data-testid="user-tag-chip"
      :data-tag-id="tag"
    >
      {{ display.label }}
    </el-tag>
  </el-tooltip>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import {
  resolveUserTagDisplay,
  userTagElementType,
} from "../utils/vrcUserTags";

const props = defineProps<{
  tag: string;
}>();

const { t } = useI18n();

const display = computed(() => resolveUserTagDisplay(props.tag, t));
const tagType = computed(() => userTagElementType(props.tag));
const tooltipLines = computed(() => display.value.tooltip.split("\n"));
</script>

<style scoped>
.vrc-user-tag-chip {
  margin: 0.15rem 0.2rem 0 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
