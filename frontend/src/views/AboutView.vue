<template>
  <div class="space-y-6">
    <div class="page-header">
      <h1 class="page-title">About chklst</h1>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2">
      <button
        v-for="t in tabs"
        :key="t.id"
        @click="activeTab = t.id"
        :class="[
          'px-4 py-2 rounded-lg text-sm font-medium transition-all',
          activeTab === t.id ? 'bg-accent text-white' : 'bg-surface-light text-gray-300 hover:bg-surface-border',
        ]"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- ABOUT TAB -->
    <div v-if="activeTab === 'about'" class="max-w-3xl">
      <Card>
        <div class="text-center mb-8 pb-8 border-b border-surface-border">
          <h1 class="text-5xl font-bold text-accent mb-2">chklst</h1>
          <p class="text-xl text-gray-300">Deployment Tracker with a local AI assistant (Parson)</p>
          <p class="text-sm text-gray-500 mt-2">Version 3.0.0</p>
        </div>

        <div class="space-y-6">
          <div>
            <h2 class="text-2xl font-bold text-white mb-3">About This Application</h2>
            <p class="text-gray-300 leading-relaxed">
              chklst tracks software deployments across projects and environments, and automates the
              daily status report with <strong class="text-white">Parson</strong> — a fully local AI
              assistant (Ollama / Gemma) that runs on this machine, so your data never leaves it.
            </p>
            <p class="text-gray-300 leading-relaxed mt-4">
              It records deployment history, computes deployment-health analytics, and each evening
              drafts and emails the daily activity report to your team automatically.
            </p>
          </div>

          <div>
            <h2 class="text-xl font-bold text-white mb-3">Tech Stack</h2>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-center">
              <div v-for="s in stack" :key="s.k" class="bg-surface-light p-4 rounded-lg border border-surface-border">
                <p class="font-semibold text-accent text-sm">{{ s.k }}</p>
                <p class="text-gray-300">{{ s.v }}</p>
              </div>
            </div>
          </div>

          <div class="bg-surface-light p-6 rounded-lg border border-surface-border">
            <h2 class="text-xl font-bold text-white mb-3">Developer</h2>
            <p class="text-gray-300"><span class="font-semibold">Developer:</span> Kannan</p>
            <p class="text-gray-300"><span class="font-semibold">Company:</span> TTS</p>
          </div>

          <div class="text-center pt-6 border-t border-surface-border">
            <p class="text-gray-500 text-sm">© 2026 chklst. Built for fast, reliable deployment tracking.</p>
          </div>
        </div>
      </Card>
    </div>

    <!-- FEATURES TAB -->
    <div v-else class="space-y-4">
      <p class="text-gray-400 text-sm">Everything chklst can do, grouped by area.</p>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card v-for="group in featureGroups" :key="group.title" :title="group.title">
          <ul class="space-y-3">
            <li v-for="f in group.features" :key="f.name" class="flex items-start gap-3">
              <component :is="f.icon" class="w-4 h-4 text-accent mt-0.5 shrink-0" />
              <div>
                <p class="text-white text-sm font-medium">
                  {{ f.name }}
                  <span v-if="f.tag" class="ml-1 text-[10px] px-1.5 py-0.5 rounded bg-accent/20 text-accent align-middle">{{ f.tag }}</span>
                </p>
                <p class="text-gray-400 text-xs">{{ f.desc }}</p>
              </div>
            </li>
          </ul>
        </Card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Card from '../components/ui/Card.vue'
import {
  Rocket, Folder, BookOpen, Clock, Sparkles, Activity, TrendingUp, BarChart3,
  Download, Mail, CalendarClock, CalendarOff, Brain, ShieldCheck, Server, Keyboard, Settings,
} from 'lucide-vue-next'

const tabs = [
  { id: 'about', label: 'About' },
  { id: 'features', label: 'Features' },
]
const activeTab = ref('features')

const stack = [
  { k: 'Frontend', v: 'Vue 3 + Vite' },
  { k: 'Backend', v: 'Go + Fiber' },
  { k: 'Database', v: 'SQLite' },
  { k: 'AI', v: 'Ollama / Gemma' },
]

const featureGroups = [
  {
    title: 'Deployment',
    features: [
      { name: 'Quick Deploy', icon: Rocket, desc: 'Favorite cards, search, one-click save & copy to Teams format. Override date/developer when logging another dev’s patch.' },
      { name: 'Projects & Components', icon: Folder, desc: 'CRUD, duplicate a project as a new copy, TOML import/export with conflict handling.' },
      { name: 'Library', icon: BookOpen, desc: 'Preset developers, build/deploy servers, environments for fast form filling.' },
      { name: 'History', icon: Clock, desc: 'Search, filter by week/month/project, edit timestamp & developer, view/copy any deployment.' },
    ],
  },
  {
    title: 'Parson — local AI',
    features: [
      { name: 'Daily Summary', icon: Sparkles, desc: 'Auto-pulls the day’s deployments + your notes and writes the activity-report email.', tag: 'AI' },
      { name: 'Model picker & prompt config', icon: Brain, desc: 'Choose the Ollama model, tune temperature, and edit Parson’s system prompt — all in Settings.' },
      { name: 'Scheduler', icon: CalendarClock, desc: 'Generates at 6:30 PM, auto-sends at 8 PM (Sun–Thu). Restart-safe catch-up + send retries.' },
      { name: 'Holidays', icon: CalendarOff, desc: 'Skip generating/sending on dates you mark as holidays.' },
      { name: 'Parson Activity', icon: Activity, desc: 'Per-day log: generation time, duration, reviewed, locked, and sent status.' },
    ],
  },
  {
    title: 'Analytics',
    features: [
      { name: 'Project health scoring', icon: TrendingUp, desc: '0–100 health per project with Healthy / Watch / Dirty badges, worst-first.', tag: 'AI' },
      { name: 'Red-flag detection', icon: ShieldCheck, desc: 'Flags 2+ patches/day, no-JIRA patches, and >3 patches/week automatically.' },
      { name: 'Week / month + frequency', icon: BarChart3, desc: 'This-week & this-month rollups and a per-week deployment frequency graph.' },
      { name: 'Developer / environment / trend', icon: BarChart3, desc: 'Per-developer JIRA discipline, per-environment risk, and 6-month health trend.' },
      { name: 'AI narrative', icon: Brain, desc: 'Parson writes an analysis with recommendations from the computed metrics.', tag: 'AI' },
    ],
  },
  {
    title: 'Reports, Export & System',
    features: [
      { name: 'Reports & Export', icon: Download, desc: 'Monthly reports and Excel / PDF / CSV export of deployment data.' },
      { name: 'Email (SMTP)', icon: Mail, desc: 'Send reports via your mail server, with a dedicated test recipient and a PAi filter tag.' },
      { name: 'Settings', icon: Settings, desc: 'Jira, SMTP, webhooks, Parson AI, schedule and holidays — all configurable in-app.' },
      { name: 'Docker + backups', icon: Server, desc: 'One-command Docker deploy on port 8000, with automatic database backups.' },
      { name: 'Keyboard shortcuts', icon: Keyboard, desc: 'Press ? for help; g+key jumps between sections; Ctrl+S / Ctrl+R in Quick Deploy.' },
    ],
  },
]
</script>
