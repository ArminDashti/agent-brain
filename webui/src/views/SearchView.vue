<script setup lang="ts">
import { ref } from 'vue'
import { searchKnowledge, type KnowledgeItem, type SearchMode } from '@/lib/api'

const query = ref('')
const mode = ref<SearchMode>('hybrid')
const items = ref<KnowledgeItem[]>([])
const error = ref('')
const loading = ref(false)

async function onSearch() {
  error.value = ''
  loading.value = true
  try {
    const res = await searchKnowledge({ query: query.value, mode: mode.value })
    items.value = res.items || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Search failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold">Search</h1>
      <p class="text-sm text-muted-foreground">Structured (SQLite), semantic (Qdrant), or hybrid.</p>
    </div>
    <form class="flex flex-wrap gap-2" @submit.prevent="onSearch">
      <input
        v-model="query"
        required
        class="min-w-[16rem] flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm"
        placeholder="Query…"
      />
      <select v-model="mode" class="rounded-md border border-border bg-background px-3 py-2 text-sm">
        <option value="structured">structured</option>
        <option value="semantic">semantic</option>
        <option value="hybrid">hybrid</option>
      </select>
      <button class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground" type="submit">
        Search
      </button>
    </form>
    <p v-if="error" class="text-sm text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-sm text-muted-foreground">Searching…</p>
    <ul class="space-y-3">
      <li v-for="item in items" :key="item.id" class="rounded-lg border border-border bg-card p-4">
        <h3 class="font-medium">{{ item.title }}</h3>
        <p class="mt-1 text-xs text-muted-foreground">{{ item.kind }} <span v-if="item.score != null">· score {{ item.score.toFixed(3) }}</span></p>
        <p class="mt-2 whitespace-pre-wrap text-sm text-muted-foreground">{{ item.body }}</p>
      </li>
    </ul>
  </div>
</template>
