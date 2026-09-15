<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fmtNum, getSessionThinking } from '@/lib/api'

const route = useRoute()
const uuid = computed(() => route.params.uuid as string)
const items = ref<Array<Record<string, unknown>>>([])
const error = ref('')

onMounted(async () => {
  try {
    const data = await getSessionThinking(uuid.value)
    items.value = data.thinking
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed'
  }
})
</script>

<template>
  <div>
    <div class="flex items-center gap-3">
      <h1 class="text-2xl font-semibold">Thinking</h1>
      <nav class="ml-auto flex gap-3 text-sm">
        <RouterLink class="text-muted-foreground hover:text-foreground" :to="`/long-term-memory/${uuid}`">
          Overview
        </RouterLink>
        <RouterLink class="text-primary" :to="`/long-term-memory/${uuid}/thinking`">Thinking</RouterLink>
        <RouterLink
          class="text-muted-foreground hover:text-foreground"
          :to="`/long-term-memory/${uuid}/turns`"
        >
          Turns
        </RouterLink>
      </nav>
    </div>
    <p v-if="error" class="mt-4 text-red-400">{{ error }}</p>
    <ol v-else class="mt-6 space-y-3">
      <li
        v-for="(t, i) in items"
        :key="String(t.bubble_id)"
        class="rounded-lg border border-border bg-card p-4"
      >
        <div class="flex gap-3 text-xs text-muted-foreground">
          <span>#{{ i + 1 }}</span>
          <span>duration {{ fmtNum(t.thinking_duration_ms as number | null) }} ms</span>
        </div>
        <pre class="mt-2 whitespace-pre-wrap text-sm">{{
          (t.thinking_text as string) || '(no thinking text in extract)'
        }}</pre>
      </li>
      <li v-if="items.length === 0" class="text-sm text-muted-foreground">
        No thinking steps for this session.
      </li>
    </ol>
  </div>
</template>
