import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/lib/api'
import LoginView from '@/views/LoginView.vue'
import KindView from '@/views/KindView.vue'
import SearchView from '@/views/SearchView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/knowledge/general_knowledge' },
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    { path: '/knowledge/:kind', name: 'kind', component: KindView },
    { path: '/search', name: 'search', component: SearchView },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) return true
  if (!getToken()) return { name: 'login', query: { redirect: to.fullPath } }
  return true
})

export default router
