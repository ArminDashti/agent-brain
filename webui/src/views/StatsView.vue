<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { KIND_LABELS, getStats, type StatsResponse } from '@/lib/api'

const stats = ref<StatsResponse | null>(null)
const error = ref('')
const loading = ref(false)

const kindOrder = ['general_knowledge', 'expertise', 'resolved_issues', 'task_guide'] as const

async function load() {
  loading.value = true
  error.value = ''
  try {
    stats.value = await getStats()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-2xl font-semibold">Stats</h1>
      <p class="text-sm text-muted-foreground">Overview of knowledge and long-term memory.</p>
    </div>

    <p v-if="error" class="text-sm text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-sm text-muted-foreground">Loading…</p>

    <template v-if="stats">
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div class="rounded-lg border border-border bg-card p-4">
          <p class="text-xs uppercase tracking-wide text-muted-foreground">Total knowledge</p>
          <p class="mt-2 text-3xl font-semibold text-primary">{{ stats.total_knowledge }}</p>
        </div>
        <div class="rounded-lg border border-border bg-card p-4">
          <p class="text-xs uppercase tracking-wide text-muted-foreground">LTM sessions</p>
          <p class="mt-2 text-3xl font-semibold text-primary">{{ stats.session_count }}</p>
        </div>
      </div>

      <div>
        <h2 class="mb-3 text-sm font-medium uppercase tracking-wide text-muted-foreground">By kind</h2>
        <div class="grid gap-3 sm:grid-cols-2">
          <RouterLink
            v-for="k in kindOrder"
            :key="k"
            :to="`/knowledge/${k}`"
            class="flex items-center justify-between rounded-lg border border-border bg-card px-4 py-3 hover:border-primary"
          >
            <span class="text-sm">{{ KIND_LABELS[k] }}</span>
            <span class="text-lg font-semibold">{{ stats.by_kind[k] ?? 0 }}</span>
          </RouterLink>
        </div>
      </div>

      <div>
        <h2 class="mb-3 text-sm font-medium uppercase tracking-wide text-muted-foreground">Recent knowledge</h2>
        <ul class="space-y-2">
          <li
            v-for="item in stats.recent_knowledge"
            :key="item.id"
            class="rounded-lg border border-border bg-card px-4 py-3"
          >
            <div class="flex flex-wrap items-baseline gap-2">
              <span class="text-xs text-primary">{{ KIND_LABELS[item.kind] || item.kind }}</span>
              <span class="font-medium">{{ item.title }}</span>
            </div>
            <p class="mt-1 line-clamp-2 text-sm text-muted-foreground">{{ item.body }}</p>
          </li>
        </ul>
        <p v-if="!stats.recent_knowledge?.length" class="text-sm text-muted-foreground">No knowledge yet.</p>
      </div>
    </template>
  </div>
</template>
