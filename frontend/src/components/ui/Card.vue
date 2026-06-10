<template>
  <div
    :class="[
      'rounded-lg border border-surface-border bg-surface-deeper overflow-hidden',
      padding,
    ]"
  >
    <div v-if="title || $slots['header-right']" class="flex items-center justify-between border-b border-surface-border pb-4 mb-4">
      <h2 v-if="title" class="text-lg font-bold text-white">{{ title }}</h2>
      <div v-else></div>
      <slot name="header-right" />
    </div>
    <slot />
  </div>
</template>

<script setup lang="ts">
interface Props {
  title?: string
  padding?: string
}

withDefaults(defineProps<Props>(), {
  padding: 'p-6',
})
</script>
