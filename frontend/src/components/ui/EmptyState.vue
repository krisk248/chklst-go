<template>
  <div class="text-center py-10">
    <div class="p-4 bg-surface-light rounded-full w-16 h-16 mx-auto mb-4 flex items-center justify-center">
      <component :is="icon" class="w-8 h-8 text-gray-500" />
    </div>
    <p class="text-gray-400 mb-1">{{ title }}</p>
    <p v-if="description" class="text-gray-500 text-sm mb-4">{{ description }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { Inbox } from 'lucide-vue-next'

withDefaults(defineProps<{
  title: string
  description?: string
  icon?: Component
}>(), {
  icon: () => Inbox,
})
</script>
