<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'
import { clearToken, getToken, KIND_LABELS } from '@/lib/api'

const route = useRoute()
const router = useRouter()
const authed = computed(() => !!getToken() && route.name !== 'login')

const nav = [
  { kind: 'general_knowledge', to: '/knowledge/general_knowledge' },
  { kind: 'solution', to: '/knowledge/solution' },
  { kind: 'task_procedure', to: '/knowledge/task_procedure' },
  { kind: 'expertise', to: '/knowledge/expertise' },
]

function logout() {
  clearToken()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen">
    <header v-if="authed" class="border-b border-border bg-card/60 backdrop-blur">
      <div class="mx-auto flex max-w-6xl flex-wrap items-center gap-4 px-4 py-3">
        <RouterLink to="/knowledge/general_knowledge" class="text-sm font-semibold tracking-wide text-primary">
          Agent Brain
        </RouterLink>
        <nav class="flex flex-wrap gap-3 text-sm text-muted-foreground">
          <RouterLink
            v-for="item in nav"
            :key="item.kind"
            class="hover:text-foreground"
            :to="item.to"
          >
            {{ KIND_LABELS[item.kind] }}
          </RouterLink>
          <RouterLink class="hover:text-foreground" to="/search">Search</RouterLink>
        </nav>
        <button class="ml-auto text-sm text-muted-foreground hover:text-foreground" @click="logout">
          Sign out
        </button>
      </div>
    </header>
    <main class="mx-auto max-w-6xl px-4 py-6">
      <RouterView />
    </main>
  </div>
</template>
