<template>
  <div class="min-h-screen flex items-center justify-center bg-background p-4">
    <div class="w-full max-w-sm bg-surface-deeper border border-surface-border rounded-lg p-8 shadow-xl">
      <div class="text-center mb-6">
        <h1 class="text-3xl font-bold text-accent">chklst</h1>
        <p class="text-sm text-gray-400 mt-1">Sign in to continue</p>
      </div>
      <form @submit.prevent="submit" class="space-y-4">
        <div>
          <label class="block text-sm text-gray-300 mb-1">Password</label>
          <input
            ref="pwInput"
            v-model="password"
            type="password"
            autofocus
            class="w-full px-3 py-2 text-sm bg-surface-light border border-surface-border rounded-lg text-white focus:outline-none focus:border-accent"
            placeholder="••••••••"
          />
        </div>
        <p v-if="errorMsg" class="text-red-400 text-sm">{{ errorMsg }}</p>
        <button
          type="submit"
          :disabled="busy || !password"
          class="w-full px-4 py-2 rounded-lg bg-accent text-white font-medium disabled:opacity-60"
        >
          {{ busy ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const password = ref('')
const busy = ref(false)
const errorMsg = ref('')

const submit = async () => {
  busy.value = true
  errorMsg.value = ''
  const res = await auth.login(password.value)
  busy.value = false
  if (!res.ok) {
    errorMsg.value = res.error || 'Login failed'
    password.value = ''
  }
}
</script>
