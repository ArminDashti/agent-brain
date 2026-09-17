import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/lib/api'
import LoginView from '@/views/LoginView.vue'
import StatsView from '@/views/StatsView.vue'
import KindView from '@/views/KindView.vue'
import SearchView from '@/views/SearchView.vue'
import SessionsView from '@/views/ltm/SessionsView.vue'
import SessionOverviewView from '@/views/ltm/SessionOverviewView.vue'
import SessionThinkingView from '@/views/ltm/SessionThinkingView.vue'
import SessionTurnsView from '@/views/ltm/SessionTurnsView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/stats' },
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    { path: '/stats', name: 'stats', component: StatsView },
    { path: '/knowledge/:kind', name: 'kind', component: KindView },
    { path: '/search', name: 'search', component: SearchView },
    { path: '/sessions', name: 'sessions', component: SessionsView },
    { path: '/sessions/:uuid', name: 'session', component: SessionOverviewView },
    { path: '/sessions/:uuid/thinking', name: 'session-thinking', component: SessionThinkingView },
    { path: '/sessions/:uuid/turns', name: 'session-turns', component: SessionTurnsView },
    { path: '/long-term-memory', redirect: '/sessions' },
    { path: '/long-term-memory/:uuid', redirect: (to) => `/sessions/${to.params.uuid}` },
    {
      path: '/long-term-memory/:uuid/thinking',
      redirect: (to) => `/sessions/${to.params.uuid}/thinking`,
    },
    {
      path: '/long-term-memory/:uuid/turns',
      redirect: (to) => `/sessions/${to.params.uuid}/turns`,
    },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) return true
  if (!getToken()) return { name: 'login', query: { redirect: to.fullPath } }
  return true
})

export default router
