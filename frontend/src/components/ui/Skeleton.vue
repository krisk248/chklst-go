<template>
  <div class="space-y-2 animate-pulse" aria-hidden="true">
    <div
      v-for="i in rows"
      :key="i"
      class="bg-surface-light rounded"
      :class="height"
    ></div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  rows?: number
  height?: string
}>(), {
  rows: 4,
  height: 'h-10',
})
</script>
