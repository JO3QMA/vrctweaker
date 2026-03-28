<template>
  <el-tag size="small" :type="tagType" class="vrc-status-tag">
    {{ displayLabel }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { vrcStatusElementTagType } from "../utils/vrcStatus";

const props = withDefaults(
  defineProps<{
    /** VRChat API の status（例: join me, active, offline） */
    status?: string | null;
  }>(),
  { status: "" },
);

const tagType = computed(() => vrcStatusElementTagType(props.status));

const displayLabel = computed(() => {
  const s = props.status?.trim();
  return s ? s : "—";
});
</script>

<style scoped>
.vrc-status-tag {
  flex-shrink: 0;
  box-sizing: border-box;
  width: 9ch;
  max-width: 100%;
  justify-content: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
