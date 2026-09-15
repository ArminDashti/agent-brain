<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { fmtNum, fmtPct, listSessions, type SessionSummary } from '@/lib/api'

const sessions = ref<SessionSummary[]>([])
const q = ref('')
const error = ref('')
const loading = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await listSessions(q.value.trim())
    sessions.value = data.sessions
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex flex-wrap items-end gap-3">
      <div>
        <h1 class="text-2xl font-semibold">Long term memory</h1>
        <p class="text-sm text-muted-foreground">Agent sessions: context %, tokens, mode</p>
      </div>
      <form class="ml-auto flex gap-2" @submit.prevent="load">
        <input
          v-model="q"
          placeholder="Search name…"
          class="rounded-md border border-border bg-card px-3 py-2 text-sm"
        />
        <button class="rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground" type="submit">
          Search
        </button>
      </form>
    </div>
    <p v-if="error" class="mt-4 text-sm text-red-400">{{ error }}</p>
    <p v-else-if="loading" class="mt-4 text-sm text-muted-foreground">Loading…</p>
    <div v-else class="mt-6 overflow-x-auto rounded-lg border border-border">
      <table class="min-w-full text-left text-sm">
        <thead class="bg-muted/40 text-muted-foreground">
          <tr>
            <th class="px-3 py-2 font-medium">Name</th>
            <th class="px-3 py-2 font-medium">Context</th>
            <th class="px-3 py-2 font-medium">In / Out tokens</th>
            <th class="px-3 py-2 font-medium">Mode</th>
            <th class="px-3 py-2 font-medium">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in sessions" :key="s.uuid" class="border-t border-border hover:bg-muted/20">
            <td class="px-3 py-2">
              <RouterLink class="text-primary hover:underline" :to="`/long-term-memory/${s.uuid}`">
                {{ s.name || s.uuid }}
              </RouterLink>
              <div class="text-xs text-muted-foreground">{{ s.subtitle }}</div>
            </td>
            <td class="px-3 py-2">{{ fmtPct(s.context_usage_percent) }}</td>
            <td class="px-3 py-2">{{ fmtNum(s.input_tokens) }} / {{ fmtNum(s.output_tokens) }}</td>
            <td class="px-3 py-2">{{ s.unified_mode || '—' }}</td>
            <td class="px-3 py-2">{{ s.status || '—' }}</td>
          </tr>
          <tr v-if="sessions.length === 0">
            <td colspan="5" class="px-3 py-8 text-center text-muted-foreground">
              No sessions yet. Ingest via MCP ingest_session.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
