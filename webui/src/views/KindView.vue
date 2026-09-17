<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { KIND_LABELS, listKnowledge, type KnowledgeItem } from '@/lib/api'

const route = useRoute()
const kind = computed(() => String(route.params.kind || 'general_knowledge'))
const title = computed(() => KIND_LABELS[kind.value] || kind.value)

const items = ref<KnowledgeItem[]>([])
const q = ref('')
const error = ref('')
const loading = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await listKnowledge(kind.value, q.value)
    items.value = res.items || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  } finally {
    loading.value = false
  }
}

watch(kind, () => {
  q.value = ''
  load()
})

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold">{{ title }}</h1>
        <p class="text-sm text-muted-foreground">Browse entries for this category (read-only).</p>
      </div>
      <form class="flex gap-2" @submit.prevent="load">
        <input
          v-model="q"
          class="rounded-md border border-border bg-background px-3 py-2 text-sm"
          placeholder="Filter…"
        />
        <button class="rounded-md border border-border px-3 py-2 text-sm" type="submit">Filter</button>
      </form>
    </div>

    <p v-if="error" class="text-sm text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-sm text-muted-foreground">Loading…</p>

    <ul class="space-y-3">
      <li
        v-for="item in items"
        :key="item.id"
        class="rounded-lg border border-border bg-card p-4"
      >
        <h3 class="font-medium">{{ item.title }}</h3>
        <p class="mt-1 whitespace-pre-wrap text-sm text-muted-foreground">{{ item.body }}</p>
        <p class="mt-2 text-xs text-muted-foreground">
          sqlite:{{ item.in_sqlite ? 'yes' : 'no' }} · qd:{{ item.in_qdrant ? 'yes' : 'no' }}
          <span v-if="item.tags?.length"> · {{ item.tags.join(', ') }}</span>
        </p>
      </li>
    </ul>
    <p v-if="!loading && !items.length" class="text-sm text-muted-foreground">No entries yet.</p>
  </div>
</template>
