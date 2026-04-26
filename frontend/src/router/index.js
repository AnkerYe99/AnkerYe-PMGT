import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    component: () => import('@/components/Layout.vue'),
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue')
      },
      {
        path: 'projects',
        name: 'Projects',
        component: () => import('@/views/Projects.vue')
      },
      {
        path: 'projects/new',
        name: 'ProjectNew',
        component: () => import('@/views/ProjectEdit.vue'),
        meta: { roles: ['admin', 'developer'] }
      },
      {
        path: 'projects/:id',
        name: 'ProjectDetail',
        component: () => import('@/views/ProjectDetail.vue')
      },
      {
        path: 'projects/:id/edit',
        name: 'ProjectEdit',
        component: () => import('@/views/ProjectEdit.vue'),
        meta: { roles: ['admin', 'developer'] }
      },
      {
        path: 'search',
        name: 'Search',
        component: () => import('@/views/Search.vue')
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { roles: ['admin'] }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/Settings.vue'),
        meta: { roles: ['admin'] }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    if (auth.token && to.name === 'Login') {
      next('/dashboard')
    } else {
      next()
    }
    return
  }
  if (!auth.token) {
    next('/login')
    return
  }
  if (to.meta.roles && !to.meta.roles.includes(auth.user?.role)) {
    next('/dashboard')
    return
  }
  next()
})

export default router
