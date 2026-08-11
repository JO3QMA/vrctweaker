<script setup lang="ts">
defineOptions({
  inheritAttrs: false,
});
</script>

<template>
  <el-input v-bind="$attrs">
    <template v-for="(_, slotName) in $slots" #[slotName]="slotProps">
      <slot :name="slotName" v-bind="slotProps || {}" />
    </template>
  </el-input>
</template>
