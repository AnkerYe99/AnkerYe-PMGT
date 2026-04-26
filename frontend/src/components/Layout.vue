<template>
  <div class="flex h-screen bg-slate-50 overflow-hidden">
    <Sidebar />
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Top bar -->
      <header class="bg-white border-b border-slate-200 px-6 py-3 flex items-center gap-4 flex-shrink-0">
        <!-- Breadcrumb -->
        <nav class="flex items-center gap-2 text-sm flex-1 min-w-0">
          <span class="text-slate-400">{{ currentSection }}</span>
          <ChevronRight v-if="subPage" class="w-4 h-4 text-slate-300 flex-shrink-0" />
          <span v-if="subPage" class="text-slate-700 font-medium truncate">{{ subPage }}</span>
        </nav>

        <!-- Quick Search -->
        <div class="relative">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            type="text"
            v-model="searchQ"
            @keydown.enter="doSearch"
            placeholder="搜索项目、文档..."
            class="pl-9 pr-4 py-1.5 text-sm border border-slate-200 rounded-lg bg-slate-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 w-56 transition-all focus:w-72"
          />
        </div>

        <!-- Notification / User -->
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-indigo-600 rounded-full flex items-center justify-center text-white text-xs font-bold cursor-pointer" :title="auth.user?.display_name">
            {{ avatarInitial }}
          </div>
        </div>
      </header>

      <!-- Main Content -->
      <main class="flex-1 overflow-y-auto">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
import Sidebar from './Sidebar.vue'
import { ChevronRight, Search } from 'lucide-vue-next'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const searchQ = ref('')

const avatarInitial = computed(() => {
  const name = auth.user?.display_name || auth.user?.username || '?'
  return name.charAt(0).toUpperCase()
})

const currentSection = computed(() => {
  const map = {
    '/dashboard': '仪表盘',
    '/projects': '项目',
    '/search': '搜索',
    '/users': '用户管理',
    '/settings': '系统设置'
  }
  for (const [key, val] of Object.entries(map)) {
    if (route.path.startsWith(key)) return val
  }
  return '首页'
})

const subPage = computed(() => {
  if (route.name === 'ProjectDetail' || route.name === 'ProjectEdit') {
    return route.params.id ? `项目 #${route.params.id}` : '新建项目'
  }
  return null
})

function doSearch() {
  if (searchQ.value.trim()) {
    router.push({ name: 'Search', query: { q: searchQ.value.trim() } })
    searchQ.value = ''
  }
}
</script>
