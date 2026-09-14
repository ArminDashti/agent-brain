<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '@/lib/api'

const router = useRouter()
const route = useRoute()
const username = ref('armin')
const password = ref('dopadopa123')
const error = ref('')
const loading = ref(false)

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await login(username.value, password.value)
    const redirect = (route.query.redirect as string) || '/knowledge/general_knowledge'
    router.replace(redirect)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto mt-16 max-w-md rounded-lg border border-border bg-card p-6">
    <h1 class="text-xl font-semibold">Sign in</h1>
    <p class="mt-1 text-sm text-muted-foreground">Agent Brain knowledge store</p>
    <form class="mt-6 space-y-4" @submit.prevent="onSubmit">
      <label class="block text-sm">
        <span class="text-muted-foreground">Username</span>
        <input v-model="username" class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2" />
      </label>
      <label class="block text-sm">
        <span class="text-muted-foreground">Password</span>
        <input v-model="password" type="password" class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2" />
      </label>
      <p v-if="error" class="text-sm text-red-400">{{ error }}</p>
      <button
        class="w-full rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
        :disabled="loading"
        type="submit"
      >
        {{ loading ? 'Signing in…' : 'Sign in' }}
      </button>
    </form>
  </div>
</template>
