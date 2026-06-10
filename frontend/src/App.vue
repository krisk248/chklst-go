<template>
  <!-- Auth gate: show login when auth is enabled and the session isn't valid -->
  <LoginGate v-if="authStore.checked && authStore.authEnabled && !authStore.authenticated" />
  <MainLayout v-else>
    <RouterView />
    <Notification v-if="appStore.hasNotification" />
  </MainLayout>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from './stores'
import { useAuthStore } from './stores/auth'
import MainLayout from './components/layout/MainLayout.vue'
import LoginGate from './components/LoginGate.vue'
import Notification from './components/ui/Notification.vue'
import { RouterView } from 'vue-router'

const appStore = useAppStore()
const authStore = useAuthStore()

onMounted(() => authStore.checkStatus())

// WebSocket disabled - single user local app doesn't need real-time updates
</script>
