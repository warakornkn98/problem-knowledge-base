import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/features/auth/LoginView.vue'),
    meta: { public: true, layout: 'blank' },
  },
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      { path: '', redirect: { name: 'dashboard' } },
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/features/dashboard/DashboardView.vue'),
        meta: { title: 'Dashboard' },
      },
      {
        path: 'problems',
        name: 'problems',
        component: () => import('@/features/problems/ProblemListView.vue'),
        meta: { title: 'Problems' },
      },
      {
        path: 'problems/new',
        name: 'problem-new',
        component: () => import('@/features/problems/ProblemFormView.vue'),
        meta: { title: 'New Problem' },
      },
      {
        path: 'problems/:id',
        name: 'problem-detail',
        component: () => import('@/features/problems/ProblemDetailView.vue'),
        meta: { title: 'Problem' },
        props: true,
      },
      {
        path: 'problems/:id/edit',
        name: 'problem-edit',
        component: () => import('@/features/problems/ProblemFormView.vue'),
        meta: { title: 'Edit Problem' },
        props: true,
      },
      {
        path: 'search',
        name: 'search',
        component: () => import('@/features/problems/SearchView.vue'),
        meta: { title: 'Search' },
      },
      {
        path: 'categories',
        name: 'categories',
        component: () => import('@/features/categories/CategoriesView.vue'),
        meta: { title: 'Categories' },
      },
      {
        path: 'tags',
        name: 'tags',
        component: () => import('@/features/tags/TagsView.vue'),
        meta: { title: 'Tags' },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: { name: 'dashboard' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.ready) {
    await auth.bootstrap()
  }
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'login', query: to.fullPath !== '/' ? { redirect: to.fullPath } : undefined }
  }
  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
  return true
})

router.afterEach((to) => {
  const base = 'Problem Knowledge Base'
  document.title = to.meta.title ? `${to.meta.title as string} · ${base}` : base
})

export default router
