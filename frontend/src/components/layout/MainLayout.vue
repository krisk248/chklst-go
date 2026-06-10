<template>
  <div class="main-container">
    <Sidebar />
    <div class="flex-1 flex flex-col ml-48">
      <Header />
      <main class="flex-1 overflow-y-auto bg-background p-8">
        <slot />
      </main>
    </div>

    <!-- Keyboard shortcuts help (press ?) -->
    <Teleport to="body">
      <div
        v-if="showShortcuts"
        class="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4"
        @click.self="showShortcuts = false"
      >
        <div class="bg-surface-deeper border border-surface-border rounded-lg max-w-md w-full p-6 shadow-xl">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-bold text-white">Keyboard Shortcuts</h2>
            <button class="text-gray-400 hover:text-white" @click="showShortcuts = false">✕</button>
          </div>
          <div class="space-y-2 text-sm">
            <div v-for="s in shortcuts" :key="s.keys" class="flex items-center justify-between">
              <span class="text-gray-300">{{ s.label }}</span>
              <kbd class="px-2 py-0.5 bg-surface-light border border-surface-border rounded text-xs text-gray-200">{{ s.keys }}</kbd>
            </div>
          </div>
          <p class="text-[11px] text-gray-500 mt-4">Quick Deploy also has Ctrl+S (save), Ctrl+R (repeat last), Ctrl+L (clear).</p>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const router = useRouter()
const showShortcuts = ref(false)

const shortcuts = [
  { keys: '?', label: 'Show this help' },
  { keys: 'g q', label: 'Go to Quick Deploy' },
  { keys: 'g d', label: 'Go to Daily Summary' },
  { keys: 'g p', label: 'Go to Projects' },
  { keys: 'g h', label: 'Go to History' },
  { keys: 'g a', label: 'Go to Analysis' },
  { keys: 'g s', label: 'Go to Settings' },
]

const gotoMap: Record<string, string> = {
  q: '/deployment',
  d: '/daily-summary',
  p: '/projects',
  h: '/history',
  a: '/analysis',
  s: '/settings',
}

// "g then <key>" navigation, gmail-style. Inactive while typing in a field.
let gPressed = false
let gTimer: ReturnType<typeof setTimeout> | null = null

const isTyping = (e: KeyboardEvent) => {
  const t = e.target as HTMLElement | null
  return !!t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable)
}

const onKeydown = (e: KeyboardEvent) => {
  if (isTyping(e) || e.ctrlKey || e.metaKey || e.altKey) return

  if (e.key === '?') {
    showShortcuts.value = !showShortcuts.value
    return
  }
  if (e.key === 'Escape') {
    showShortcuts.value = false
    return
  }
  if (e.key === 'g') {
    gPressed = true
    if (gTimer) clearTimeout(gTimer)
    gTimer = setTimeout(() => (gPressed = false), 1500)
    return
  }
  if (gPressed && gotoMap[e.key]) {
    gPressed = false
    showShortcuts.value = false
    router.push(gotoMap[e.key])
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>
