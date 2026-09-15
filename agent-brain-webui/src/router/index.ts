import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/lib/api'
import LoginView from '@/views/LoginView.vue'
import KindView from '@/views/KindView.vue'
import SearchView from '@/views/SearchView.vue'
import SessionsView from '@/views/ltm/SessionsView.vue'
import SessionOverviewView from '@/views/ltm/SessionOverviewView.vue'
import SessionThinkingView from '@/views/ltm/SessionThinkingView.vue'
import SessionTurnsView from '@/views/ltm/SessionTurnsView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/knowledge/general_knowledge' },
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    { path: '/knowledge/:kind', name: 'kind', component: KindView },
    { path: '/search', name: 'search', component: SearchView },
    { path: '/long-term-memory', name: 'ltm', component: SessionsView },
    { path: '/long-term-memory/:uuid', name: 'ltm-session', component: SessionOverviewView },
    {
      path: '/long-term-memory/:uuid/thinking',
      name: 'ltm-thinking',
      component: SessionThinkingView,
    },
    { path: '/long-term-memory/:uuid/turns', name: 'ltm-turns', component: SessionTurnsView },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) return true
  if (!getToken()) return { name: 'login', query: { redirect: to.fullPath } }
  return true
})

export default router
