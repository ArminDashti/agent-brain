<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fmtNum, fmtPct, getSession, type SessionSummary } from '@/lib/api'

const route = useRoute()
const uuid = computed(() => route.params.uuid as string)
const session = ref<SessionSummary | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    session.value = await getSession(uuid.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed'
  }
})
</script>

<template>
  <div v-if="error" class="text-red-400">{{ error }}</div>
  <div v-else-if="!session" class="text-muted-foreground">Loading…</div>
  <div v-else>
    <div class="flex flex-wrap items-start gap-3">
      <div>
        <h1 class="text-2xl font-semibold">{{ session.name || session.uuid }}</h1>
        <p class="text-sm text-muted-foreground">{{ session.subtitle }}</p>
      </div>
      <nav class="ml-auto flex gap-3 text-sm">
        <RouterLink class="text-primary" :to="`/sessions/${uuid}`">Overview</RouterLink>
        <RouterLink
          class="text-muted-foreground hover:text-foreground"
          :to="`/sessions/${uuid}/thinking`"
        >
          Thinking
        </RouterLink>
        <RouterLink
          class="text-muted-foreground hover:text-foreground"
          :to="`/sessions/${uuid}/turns`"
        >
          Turns
        </RouterLink>
      </nav>
    </div>

    <div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="text-xs uppercase tracking-wide text-muted-foreground">Context</div>
        <div class="mt-2 text-2xl font-semibold">{{ fmtPct(session.context_usage_percent) }}</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="text-xs uppercase tracking-wide text-muted-foreground">Input tokens</div>
        <div class="mt-2 text-2xl font-semibold">{{ fmtNum(session.input_tokens) }}</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="text-xs uppercase tracking-wide text-muted-foreground">Output tokens</div>
        <div class="mt-2 text-2xl font-semibold">{{ fmtNum(session.output_tokens) }}</div>
      </div>
      <div class="rounded-lg border border-border bg-card p-4">
        <div class="text-xs uppercase tracking-wide text-muted-foreground">Cache R/W</div>
        <div class="mt-2 text-2xl font-semibold">
          {{ fmtNum(session.cache_read_tokens) }} / {{ fmtNum(session.cache_write_tokens) }}
        </div>
      </div>
    </div>

    <div class="mt-6 rounded-lg border border-border bg-card p-4 text-sm">
      <h2 class="font-medium">Useful info</h2>
      <dl class="mt-3 grid gap-2 sm:grid-cols-2">
        <div>
          <dt class="text-muted-foreground">Mode</dt>
          <dd>{{ session.unified_mode || '—' }}</dd>
        </div>
        <div>
          <dt class="text-muted-foreground">Status</dt>
          <dd>{{ session.status || '—' }}</dd>
        </div>
        <div>
          <dt class="text-muted-foreground">Workspace</dt>
          <dd class="break-all">{{ session.workspace_path || session.workspace_id || '—' }}</dd>
        </div>
        <div>
          <dt class="text-muted-foreground">Turns / thinking</dt>
          <dd>{{ session.turn_count ?? 0 }} / {{ session.thinking_count ?? 0 }}</dd>
        </div>
        <div>
          <dt class="text-muted-foreground">Lines +/-</dt>
          <dd>+{{ session.total_lines_added }} / -{{ session.total_lines_removed }}</dd>
        </div>
        <div>
          <dt class="text-muted-foreground">Files changed</dt>
          <dd>{{ session.files_changed_count }}</dd>
        </div>
      </dl>
    </div>
  </div>
</template>
