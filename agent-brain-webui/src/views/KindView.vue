<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  KIND_LABELS,
  createKnowledge,
  deleteKnowledge,
  listKnowledge,
  type KnowledgeItem,
  type StoreTarget,
} from '@/lib/api'

const route = useRoute()
const kind = computed(() => String(route.params.kind || 'general_knowledge'))
const title = computed(() => KIND_LABELS[kind.value] || kind.value)

const items = ref<KnowledgeItem[]>([])
const q = ref('')
const error = ref('')
const loading = ref(false)

const formTitle = ref('')
const formBody = ref('')
const formTags = ref('')
const storeTarget = ref<StoreTarget>('both')

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

async function onCreate() {
  error.value = ''
  try {
    await createKnowledge({
      kind: kind.value,
      title: formTitle.value,
      body: formBody.value,
      tags: formTags.value
        .split(',')
        .map((t) => t.trim())
        .filter(Boolean),
      store_target: storeTarget.value,
    })
    formTitle.value = ''
    formBody.value = ''
    formTags.value = ''
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Create failed'
  }
}

async function onDelete(id: string) {
  await deleteKnowledge(id)
  await load()
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
        <p class="text-sm text-muted-foreground">Browse and add entries for this category.</p>
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

    <form class="space-y-3 rounded-lg border border-border bg-card p-4" @submit.prevent="onCreate">
      <h2 class="text-sm font-medium">Add entry</h2>
      <input
        v-model="formTitle"
        required
        class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        placeholder="Title"
      />
      <textarea
        v-model="formBody"
        rows="4"
        class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        placeholder="Body"
      />
      <input
        v-model="formTags"
        class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        placeholder="Tags (comma-separated)"
      />
      <label class="block text-sm text-muted-foreground">
        Store target
        <select v-model="storeTarget" class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-foreground">
          <option value="postgres">postgres</option>
          <option value="qdrant">qdrant</option>
          <option value="both">both</option>
        </select>
      </label>
      <button class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground" type="submit">
        Save
      </button>
    </form>

    <p v-if="error" class="text-sm text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-sm text-muted-foreground">Loading…</p>

    <ul class="space-y-3">
      <li
        v-for="item in items"
        :key="item.id"
        class="rounded-lg border border-border bg-card p-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h3 class="font-medium">{{ item.title }}</h3>
            <p class="mt-1 whitespace-pre-wrap text-sm text-muted-foreground">{{ item.body }}</p>
            <p class="mt-2 text-xs text-muted-foreground">
              pg:{{ item.in_postgres ? 'yes' : 'no' }} · qd:{{ item.in_qdrant ? 'yes' : 'no' }}
              <span v-if="item.tags?.length"> · {{ item.tags.join(', ') }}</span>
            </p>
          </div>
          <button class="text-xs text-red-400" type="button" @click="onDelete(item.id)">Delete</button>
        </div>
      </li>
    </ul>
  </div>
</template>
