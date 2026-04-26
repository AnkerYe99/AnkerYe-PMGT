<template>
  <div class="p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-900">项目列表</h1>
        <p class="text-slate-500 text-sm mt-0.5">管理所有项目</p>
      </div>
      <router-link v-if="auth.isDev" to="/projects/new">
        <button class="btn-primary">
          <Plus class="w-4 h-4" />
          新建项目
        </button>
      </router-link>
    </div>

    <!-- Filters -->
    <div class="card p-4 mb-6">
      <div class="flex flex-wrap items-center gap-4">
        <!-- Search -->
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="searchQ" @input="debouncedSearch" placeholder="搜索项目名称..." class="input pl-9" />
        </div>

        <!-- Status tabs -->
        <div class="flex gap-1 bg-slate-100 p-1 rounded-lg">
          <button
            v-for="tab in statusTabs"
            :key="tab.value"
            @click="activeStatus = tab.value; loadProjects()"
            :class="[
              'px-3 py-1.5 rounded-md text-sm font-medium transition-all',
              activeStatus === tab.value ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'
            ]"
          >
            {{ tab.label }}
          </button>
        </div>

        <!-- View toggle -->
        <div class="flex gap-1 bg-slate-100 p-1 rounded-lg">
          <button @click="viewMode = 'grid'" :class="['p-1.5 rounded', viewMode === 'grid' ? 'bg-white shadow-sm' : 'text-slate-400 hover:text-slate-600']">
            <LayoutGrid class="w-4 h-4" />
          </button>
          <button @click="viewMode = 'list'" :class="['p-1.5 rounded', viewMode === 'list' ? 'bg-white shadow-sm' : 'text-slate-400 hover:text-slate-600']">
            <List class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-16 text-slate-400">
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      加载中...
    </div>

    <!-- Empty -->
    <div v-else-if="!projects.length" class="card p-16 text-center">
      <FolderKanban class="w-16 h-16 text-slate-200 mx-auto mb-4" />
      <h3 class="text-slate-500 font-medium mb-2">暂无项目</h3>
      <p class="text-slate-400 text-sm mb-4">还没有任何项目，快来创建第一个吧</p>
      <router-link v-if="auth.isDev" to="/projects/new">
        <button class="btn-primary">
          <Plus class="w-4 h-4" />
          创建项目
        </button>
      </router-link>
    </div>

    <!-- Grid View -->
    <div v-else-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <div
        v-for="p in projects"
        :key="p.id"
        class="card hover:shadow-md transition-all duration-200 cursor-pointer group overflow-hidden"
        @click="router.push('/projects/' + p.id)"
      >
        <!-- Color bar -->
        <div class="h-1.5 w-full" :style="{ backgroundColor: p.cover_color }"></div>
        <div class="p-5">
          <div class="flex items-start justify-between mb-3">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <h3 class="font-semibold text-slate-900 truncate group-hover:text-indigo-600 transition-colors">{{ p.title }}</h3>
                <Lock v-if="p.is_private" class="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
              </div>
              <div class="flex items-center gap-2">
                <span class="badge badge-slate text-xs">{{ typeLabel(p.type) }}</span>
                <span :class="statusBadge(p.status)">{{ statusLabel(p.status) }}</span>
              </div>
            </div>
            <!-- Actions -->
            <div class="relative" @click.stop>
              <button @click="toggleMenu(p.id)" class="p-1 text-slate-400 hover:text-slate-600 rounded transition-colors opacity-0 group-hover:opacity-100">
                <MoreHorizontal class="w-4 h-4" />
              </button>
              <div v-if="openMenu === p.id" class="absolute right-0 top-6 bg-white border border-slate-200 rounded-xl shadow-lg z-10 min-w-36 py-1">
                <router-link :to="'/projects/' + p.id + '/edit'" class="block px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 flex items-center gap-2" @click="openMenu = null">
                  <Edit2 class="w-3.5 h-3.5" />编辑
                </router-link>
                <button v-if="canEdit(p)" @click="confirmDelete(p)" class="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2">
                  <Trash2 class="w-3.5 h-3.5" />删除
                </button>
              </div>
            </div>
          </div>

          <p class="text-sm text-slate-500 line-clamp-2 mb-4 min-h-10">{{ p.summary || '暂无描述' }}</p>

          <!-- Tags -->
          <div v-if="parseTags(p.tags).length" class="flex flex-wrap gap-1 mb-3">
            <span v-for="tag in parseTags(p.tags).slice(0, 3)" :key="tag" class="text-xs bg-indigo-50 text-indigo-600 px-2 py-0.5 rounded-full">{{ tag }}</span>
            <span v-if="parseTags(p.tags).length > 3" class="text-xs text-slate-400">+{{ parseTags(p.tags).length - 3 }}</span>
          </div>

          <div class="flex items-center justify-between text-xs text-slate-400 pt-3 border-t border-slate-50">
            <span>{{ p.creator_name || '未知' }}</span>
            <div class="flex items-center gap-3">
              <span class="flex items-center gap-1"><Users class="w-3 h-3" />{{ p.member_count }}</span>
              <span>{{ formatDate(p.updated_at) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div v-else class="card overflow-hidden">
      <table class="w-full">
        <thead>
          <tr class="bg-slate-50 border-b border-slate-100">
            <th class="text-left px-6 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">项目名称</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider hidden md:table-cell">类型</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">状态</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider hidden lg:table-cell">创建者</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider hidden lg:table-cell">更新时间</th>
            <th class="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          <tr v-for="p in projects" :key="p.id" class="hover:bg-slate-50 transition-colors cursor-pointer" @click="router.push('/projects/' + p.id)">
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="w-1 h-8 rounded-full flex-shrink-0" :style="{ backgroundColor: p.cover_color }"></div>
                <div>
                  <div class="flex items-center gap-2">
                    <span class="font-medium text-slate-900 hover:text-indigo-600">{{ p.title }}</span>
                    <Lock v-if="p.is_private" class="w-3.5 h-3.5 text-slate-400" />
                  </div>
                  <div class="text-xs text-slate-400 truncate max-w-48">{{ p.summary || '暂无描述' }}</div>
                </div>
              </div>
            </td>
            <td class="px-4 py-4 hidden md:table-cell">
              <span class="badge badge-slate">{{ typeLabel(p.type) }}</span>
            </td>
            <td class="px-4 py-4">
              <span :class="statusBadge(p.status)">{{ statusLabel(p.status) }}</span>
            </td>
            <td class="px-4 py-4 text-sm text-slate-500 hidden lg:table-cell">{{ p.creator_name || '未知' }}</td>
            <td class="px-4 py-4 text-sm text-slate-400 hidden lg:table-cell">{{ formatDate(p.updated_at) }}</td>
            <td class="px-4 py-4" @click.stop>
              <div class="flex items-center gap-1">
                <router-link v-if="canEdit(p)" :to="'/projects/' + p.id + '/edit'" class="p-1.5 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded transition-colors">
                  <Edit2 class="w-3.5 h-3.5" />
                </router-link>
                <button v-if="canEdit(p)" @click="confirmDelete(p)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Delete Confirm Modal -->
    <div v-if="deleteTarget" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="deleteTarget = null">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 bg-red-100 rounded-xl flex items-center justify-center flex-shrink-0">
            <Trash2 class="w-5 h-5 text-red-600" />
          </div>
          <div>
            <h3 class="font-semibold text-slate-900">删除项目</h3>
            <p class="text-sm text-slate-500">确定要删除「{{ deleteTarget.title }}」吗？</p>
          </div>
        </div>
        <p class="text-sm text-slate-500 mb-5 bg-amber-50 rounded-lg px-3 py-2 border border-amber-100">
          删除后项目将进入回收站，管理员可恢复。
        </p>
        <div class="flex gap-3">
          <button @click="deleteTarget = null" class="flex-1 btn-secondary">取消</button>
          <button @click="doDelete" :disabled="deleting" class="flex-1 btn-danger">
            <Loader2 v-if="deleting" class="w-4 h-4 animate-spin" />
            确认删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
import { projectApi } from '@/api/index.js'
import { Plus, Search, FolderKanban, LayoutGrid, List, MoreHorizontal, Edit2, Trash2, Loader2, Lock, Users } from 'lucide-vue-next'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(true)
const projects = ref([])
const searchQ = ref('')
const activeStatus = ref('')
const viewMode = ref('grid')
const openMenu = ref(null)
const deleteTarget = ref(null)
const deleting = ref(false)

let searchTimer = null

const statusTabs = [
  { value: '', label: '全部' },
  { value: 'draft', label: '草稿' },
  { value: 'active', label: '进行中' },
  { value: 'done', label: '已完成' },
  { value: 'archived', label: '已归档' }
]

function statusBadge(s) {
  const map = { draft: 'badge badge-slate', active: 'badge badge-blue', in_progress: 'badge badge-amber', done: 'badge badge-green', completed: 'badge badge-green', archived: 'badge badge-slate' }
  return map[s] || 'badge badge-slate'
}

function statusLabel(s) {
  const map = { draft: '草稿', active: '进行中', in_progress: '进行中', done: '已完成', completed: '已完成', archived: '已归档' }
  return map[s] || s
}

function typeLabel(t) {
  const map = { general: '通用', dev: '开发', design: '设计', marketing: '营销', ops: '运营', research: '研究' }
  return map[t] || t
}

function parseTags(tags) {
  try { return JSON.parse(tags) || [] } catch { return [] }
}

function formatDate(d) {
  if (!d) return ''
  const date = new Date(d)
  const now = new Date()
  const diff = now - date
  if (diff < 86400000) return Math.floor(diff / 3600000) + 'h前'
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

function canEdit(p) {
  return auth.isAdmin || (auth.isDev && p.created_by === auth.user?.id)
}

function toggleMenu(id) {
  openMenu.value = openMenu.value === id ? null : id
}

function confirmDelete(p) {
  openMenu.value = null
  deleteTarget.value = p
}

async function doDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await projectApi.delete(deleteTarget.value.id)
    projects.value = projects.value.filter(p => p.id !== deleteTarget.value.id)
    deleteTarget.value = null
  } finally {
    deleting.value = false
  }
}

function debouncedSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadProjects(), 400)
}

async function loadProjects() {
  loading.value = true
  try {
    const params = {}
    if (activeStatus.value) params.status = activeStatus.value
    if (searchQ.value) params.q = searchQ.value
    const res = await projectApi.list(params)
    projects.value = res.data.data || []
  } finally {
    loading.value = false
  }
}

// Close menu on outside click
function handleClick(e) {
  if (!e.target.closest('.relative')) openMenu.value = null
}

onMounted(() => {
  loadProjects()
  document.addEventListener('click', handleClick)
})
</script>
