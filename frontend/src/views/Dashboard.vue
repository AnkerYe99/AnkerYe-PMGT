<template>
  <div class="p-6">
    <!-- Welcome -->
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900">
          你好，{{ auth.user?.display_name || auth.user?.username }}
          <span class="wave">👋</span>
        </h1>
        <p class="text-slate-500 mt-1">欢迎使用 AnkerYe 项目管理系统</p>
      </div>
      <router-link v-if="auth.isDev" to="/projects/new">
        <button class="btn-primary">
          <Plus class="w-4 h-4" />
          新建项目
        </button>
      </router-link>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="card p-5">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 bg-indigo-100 rounded-xl flex items-center justify-center">
            <FolderKanban class="w-5 h-5 text-indigo-600" />
          </div>
          <span class="text-xs text-slate-400 bg-slate-50 px-2 py-1 rounded-full">总计</span>
        </div>
        <div class="text-2xl font-bold text-slate-900">{{ stats.total }}</div>
        <div class="text-sm text-slate-500 mt-0.5">项目总数</div>
      </div>
      <div class="card p-5">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 bg-amber-100 rounded-xl flex items-center justify-center">
            <Activity class="w-5 h-5 text-amber-600" />
          </div>
          <span class="text-xs text-emerald-500 bg-emerald-50 px-2 py-1 rounded-full">进行中</span>
        </div>
        <div class="text-2xl font-bold text-slate-900">{{ stats.active }}</div>
        <div class="text-sm text-slate-500 mt-0.5">进行中项目</div>
      </div>
      <div class="card p-5">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 bg-emerald-100 rounded-xl flex items-center justify-center">
            <CheckCircle2 class="w-5 h-5 text-emerald-600" />
          </div>
          <span class="text-xs text-slate-400 bg-slate-50 px-2 py-1 rounded-full">完成</span>
        </div>
        <div class="text-2xl font-bold text-slate-900">{{ stats.done }}</div>
        <div class="text-sm text-slate-500 mt-0.5">已完成项目</div>
      </div>
      <div class="card p-5">
        <div class="flex items-center justify-between mb-3">
          <div class="w-10 h-10 bg-blue-100 rounded-xl flex items-center justify-center">
            <Users class="w-5 h-5 text-blue-600" />
          </div>
          <span class="text-xs text-slate-400 bg-slate-50 px-2 py-1 rounded-full">成员</span>
        </div>
        <div class="text-2xl font-bold text-slate-900">{{ stats.members }}</div>
        <div class="text-sm text-slate-500 mt-0.5">系统成员</div>
      </div>
    </div>

    <!-- Recent Projects -->
    <div class="card">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
        <h2 class="text-base font-semibold text-slate-900">最近更新的项目</h2>
        <router-link to="/projects" class="text-sm text-indigo-600 hover:text-indigo-700 font-medium">查看全部</router-link>
      </div>
      <div v-if="loading" class="p-8 text-center text-slate-400">
        <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2" />
        加载中...
      </div>
      <div v-else-if="!recentProjects.length" class="p-12 text-center">
        <FolderKanban class="w-12 h-12 text-slate-200 mx-auto mb-3" />
        <p class="text-slate-400">暂无项目</p>
        <router-link v-if="auth.isDev" to="/projects/new" class="mt-3 inline-block text-indigo-600 text-sm font-medium hover:text-indigo-700">立即创建</router-link>
      </div>
      <div v-else>
        <div
          v-for="project in recentProjects"
          :key="project.id"
          class="flex items-center gap-4 px-6 py-4 hover:bg-slate-50 transition-colors border-b border-slate-50 last:border-0 cursor-pointer"
          @click="router.push('/projects/' + project.id)"
        >
          <div class="w-1 h-10 rounded-full flex-shrink-0" :style="{ backgroundColor: project.cover_color }"></div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-medium text-slate-900 truncate">{{ project.title }}</span>
              <span v-if="project.is_private" class="badge badge-slate">
                <Lock class="w-2.5 h-2.5 mr-1" />私有
              </span>
            </div>
            <div class="text-sm text-slate-400 truncate mt-0.5">{{ project.summary || '暂无描述' }}</div>
          </div>
          <div class="flex items-center gap-3 flex-shrink-0">
            <span :class="statusBadge(project.status)">{{ statusLabel(project.status) }}</span>
            <span class="text-xs text-slate-400">{{ formatDate(project.updated_at) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
import { projectApi, userApi } from '@/api/index.js'
import { FolderKanban, Activity, CheckCircle2, Users, Plus, Loader2, Lock } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()
const loading = ref(true)
const projects = ref([])
const userCount = ref(0)

const recentProjects = computed(() => projects.value.slice(0, 10))

const stats = computed(() => ({
  total: projects.value.length,
  active: projects.value.filter(p => p.status === 'active' || p.status === 'in_progress').length,
  done: projects.value.filter(p => p.status === 'done' || p.status === 'completed').length,
  members: userCount.value
}))

function statusBadge(s) {
  const map = {
    draft: 'badge badge-slate',
    active: 'badge badge-blue',
    in_progress: 'badge badge-amber',
    done: 'badge badge-green',
    completed: 'badge badge-green',
    archived: 'badge badge-slate'
  }
  return map[s] || 'badge badge-slate'
}

function statusLabel(s) {
  const map = {
    draft: '草稿',
    active: '进行中',
    in_progress: '进行中',
    done: '已完成',
    completed: '已完成',
    archived: '已归档'
  }
  return map[s] || s
}

function formatDate(d) {
  if (!d) return ''
  const date = new Date(d)
  const now = new Date()
  const diff = now - date
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + ' 分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + ' 小时前'
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

onMounted(async () => {
  try {
    const [pRes, uRes] = await Promise.allSettled([
      projectApi.list(),
      auth.isAdmin ? userApi.list() : Promise.resolve(null)
    ])
    if (pRes.status === 'fulfilled') projects.value = pRes.value.data.data || []
    if (uRes.status === 'fulfilled' && uRes.value) userCount.value = (uRes.value.data.data || []).length
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.wave {
  display: inline-block;
  animation: wave 2s ease-in-out infinite;
}
@keyframes wave {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(20deg); }
  75% { transform: rotate(-10deg); }
}
</style>
