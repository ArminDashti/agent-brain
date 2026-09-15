<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fmtNum, getSessionTurns } from '@/lib/api'

const route = useRoute()
const uuid = computed(() => route.params.uuid as string)
const items = ref<Array<Record<string, unknown>>>([])
const error = ref('')

onMounted(async () => {
  try {
    const data = await getSessionTurns(uuid.value)
    items.value = data.turns
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed'
  }
})
</script>

<template>
  <div>
    <div class="flex items-center gap-3">
      <h1 class="text-2xl font-semibold">Turns</h1>
      <nav class="ml-auto flex gap-3 text-sm">
        <RouterLink class="text-muted-foreground hover:text-foreground" :to="`/long-term-memory/${uuid}`">
          Overview
        </RouterLink>
        <RouterLink
          class="text-muted-foreground hover:text-foreground"
          :to="`/long-term-memory/${uuid}/thinking`"
        >
          Thinking
        </RouterLink>
        <RouterLink class="text-primary" :to="`/long-term-memory/${uuid}/turns`">Turns</RouterLink>
      </nav>
    </div>
    <p v-if="error" class="mt-4 text-red-400">{{ error }}</p>
    <div v-else class="mt-6 space-y-3">
      <article
        v-for="t in items"
        :key="String(t.bubble_id)"
        class="rounded-lg border border-border bg-card p-4 text-sm"
      >
        <div class="flex flex-wrap gap-3 text-xs text-muted-foreground">
          <span class="rounded bg-muted px-2 py-0.5 text-foreground">{{ t.turn_type }}</span>
          <span v-if="t.tool_name">tool: {{ t.tool_name }}</span>
          <span v-if="t.mcp_name">mcp: {{ t.mcp_name }}</span>
          <span>
            in/out {{ fmtNum(t.input_tokens as number | null) }}/{{
              fmtNum(t.output_tokens as number | null)
            }}
          </span>
          <span v-if="t.duration_ms != null">{{ fmtNum(t.duration_ms as number | null) }} ms</span>
        </div>
        <pre v-if="t.text" class="mt-2 max-h-48 overflow-auto whitespace-pre-wrap">{{ t.text }}</pre>
        <pre
          v-if="t.has_thinking && t.thinking_text"
          class="mt-2 whitespace-pre-wrap border-t border-border pt-2 text-muted-foreground"
          >{{ t.thinking_text }}</pre
        >
      </article>
      <p v-if="items.length === 0" class="text-muted-foreground">No turns.</p>
    </div>
  </div>
</template>
