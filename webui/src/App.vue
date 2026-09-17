<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'
import { clearToken, getToken, KIND_LABELS } from '@/lib/api'

const route = useRoute()
const router = useRouter()
const authed = computed(() => !!getToken() && route.name !== 'login')

const nav = [
  { kind: 'general_knowledge', to: '/knowledge/general_knowledge' },
  { kind: 'expertise', to: '/knowledge/expertise' },
  { kind: 'resolved_issues', to: '/knowledge/resolved_issues' },
  { kind: 'task_guide', to: '/knowledge/task_guide' },
]

function logout() {
  clearToken()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen">
    <header v-if="authed" class="border-b border-border bg-card/80 backdrop-blur">
      <div class="mx-auto flex max-w-6xl flex-wrap items-center gap-4 px-4 py-3">
        <RouterLink to="/stats" class="text-sm font-semibold tracking-wide text-primary">
          Agent Brain
        </RouterLink>
        <nav class="flex flex-wrap gap-3 text-sm text-muted-foreground">
          <RouterLink class="hover:text-foreground" to="/stats">Stats</RouterLink>
          <RouterLink
            v-for="item in nav"
            :key="item.kind"
            class="hover:text-foreground"
            :to="item.to"
          >
            {{ KIND_LABELS[item.kind] }}
          </RouterLink>
          <RouterLink class="hover:text-foreground" to="/search">Search</RouterLink>
          <RouterLink class="hover:text-foreground" to="/sessions">Sessions</RouterLink>
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
