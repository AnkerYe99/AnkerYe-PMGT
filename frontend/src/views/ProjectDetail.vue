<template>
  <div class="flex h-full">
    <!-- Loading -->
    <div v-if="loading" class="flex-1 flex items-center justify-center text-slate-400">
      <Loader2 class="w-6 h-6 animate-spin mr-2" />
      加载中...
    </div>

    <template v-else-if="project">
      <!-- Left: Doc tree -->
      <div class="w-64 border-r border-slate-200 bg-white flex flex-col flex-shrink-0">
        <div class="px-4 py-3 border-b border-slate-100 flex items-center justify-between">
          <span class="text-sm font-semibold text-slate-700">文档列表</span>
          <button v-if="auth.isDev" @click="showNewDoc = true" class="p-1 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded transition-colors" title="新建文档">
            <Plus class="w-4 h-4" />
          </button>
        </div>
        <div class="flex-1 overflow-y-auto py-2">
          <div v-if="!docs.length" class="px-4 py-6 text-center text-slate-400 text-sm">暂无文档</div>
          <div
            v-for="doc in docs"
            :key="doc.id"
            @click="selectDoc(doc)"
            :class="[
              'flex items-center gap-2 px-4 py-2.5 cursor-pointer transition-colors group',
              activeDoc?.id === doc.id ? 'bg-indigo-50 text-indigo-700 border-r-2 border-indigo-600' : 'text-slate-600 hover:bg-slate-50'
            ]"
          >
            <FileText class="w-4 h-4 flex-shrink-0 opacity-60" />
            <span class="text-sm truncate flex-1">{{ doc.title }}</span>
            <button
              v-if="auth.isDev"
              @click.stop="confirmDeleteDoc(doc)"
              class="p-0.5 text-slate-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-all"
            >
              <X class="w-3 h-3" />
            </button>
          </div>
        </div>
      </div>

      <!-- Right: Content -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- Project Header -->
        <div class="bg-white border-b border-slate-200 px-6 py-4">
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <!-- Breadcrumb -->
              <nav class="flex items-center gap-2 text-sm text-slate-400 mb-2">
                <router-link to="/projects" class="hover:text-slate-600">项目</router-link>
                <ChevronRight class="w-3.5 h-3.5" />
                <span class="text-slate-700 font-medium">{{ project.title }}</span>
                <template v-if="activeDoc">
                  <ChevronRight class="w-3.5 h-3.5" />
                  <span class="text-slate-500 truncate">{{ activeDoc.title }}</span>
                </template>
              </nav>
              <div class="flex items-center gap-3 flex-wrap">
                <h1 class="text-xl font-bold text-slate-900">{{ project.title }}</h1>
                <span :class="statusBadge(project.status)">{{ statusLabel(project.status) }}</span>
                <span v-if="project.is_private" class="badge badge-slate">
                  <Lock class="w-2.5 h-2.5 mr-1" />私有
                </span>
              </div>
              <!-- Tags -->
              <div v-if="parseTags(project.tags).length" class="flex flex-wrap gap-1 mt-2">
                <span v-for="tag in parseTags(project.tags)" :key="tag" class="text-xs bg-indigo-50 text-indigo-600 px-2 py-0.5 rounded-full">{{ tag }}</span>
              </div>
            </div>

            <div class="flex items-center gap-2 flex-shrink-0 ml-4">
              <!-- Member avatars -->
              <div class="flex -space-x-2">
                <div v-for="m in members.slice(0, 4)" :key="m.id" class="w-7 h-7 bg-indigo-500 rounded-full border-2 border-white flex items-center justify-center text-white text-xs font-bold" :title="m.username">
                  {{ (m.display_name || m.username || '?').charAt(0).toUpperCase() }}
                </div>
                <div v-if="members.length > 4" class="w-7 h-7 bg-slate-300 rounded-full border-2 border-white flex items-center justify-center text-white text-xs">+{{ members.length - 4 }}</div>
              </div>

              <button v-if="auth.isDev" @click="showMemberModal = true" class="btn-secondary text-xs py-1.5">
                <Users class="w-3.5 h-3.5" />
                成员
              </button>
              <router-link v-if="auth.isDev" :to="'/projects/' + project.id + '/edit'">
                <button class="btn-primary text-xs py-1.5">
                  <Edit2 class="w-3.5 h-3.5" />
                  编辑
                </button>
              </router-link>
            </div>
          </div>
        </div>

        <!-- Doc Content -->
        <div class="flex-1 overflow-y-auto">
          <div v-if="!activeDoc" class="flex flex-col items-center justify-center h-full text-slate-400">
            <FileText class="w-16 h-16 text-slate-200 mb-4" />
            <p class="font-medium">请从左侧选择文档</p>
            <p class="text-sm mt-1">或者创建新文档开始编写</p>
          </div>
          <div v-else class="max-w-4xl mx-auto px-8 py-8">
            <h2 class="text-2xl font-bold text-slate-900 mb-2">{{ activeDoc.title }}</h2>
            <div class="flex items-center gap-4 text-sm text-slate-400 mb-6 pb-4 border-b border-slate-100">
              <span>作者：{{ activeDoc.author_name || '未知' }}</span>
              <span>更新：{{ formatDate(activeDoc.updated_at) }}</span>
            </div>
            <div class="prose prose-slate max-w-none" v-html="activeDoc.content || '<p class=&quot;text-slate-400&quot;>文档内容为空</p>'"></div>
          </div>
        </div>
      </div>
    </template>

    <!-- Not Found -->
    <div v-else class="flex-1 flex items-center justify-center text-slate-400">
      <div class="text-center">
        <FolderKanban class="w-16 h-16 text-slate-200 mx-auto mb-4" />
        <p>项目不存在或无权访问</p>
        <router-link to="/projects" class="text-indigo-600 text-sm mt-2 inline-block hover:text-indigo-700">返回项目列表</router-link>
      </div>
    </div>

    <!-- New Doc Modal -->
    <div v-if="showNewDoc" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showNewDoc = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm">
        <h3 class="font-semibold text-slate-900 mb-4">新建文档</h3>
        <input v-model="newDocTitle" type="text" placeholder="文档标题" class="input mb-4" @keydown.enter="createDoc" />
        <div class="flex gap-3">
          <button @click="showNewDoc = false" class="flex-1 btn-secondary">取消</button>
          <button @click="createDoc" :disabled="!newDocTitle.trim()" class="flex-1 btn-primary">创建</button>
        </div>
      </div>
    </div>

    <!-- Member Modal -->
    <div v-if="showMemberModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showMemberModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-md">
        <div class="flex items-center justify-between mb-5">
          <h3 class="font-semibold text-slate-900">成员管理</h3>
          <button @click="showMemberModal = false" class="p-1 text-slate-400 hover:text-slate-600 rounded">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Add member -->
        <div class="flex gap-2 mb-4">
          <select v-model="addMemberUID" class="input flex-1">
            <option value="">选择用户</option>
            <option v-for="u in availableUsers" :key="u.id" :value="u.id">{{ u.display_name || u.username }}</option>
          </select>
          <button @click="addMember" :disabled="!addMemberUID" class="btn-primary">添加</button>
        </div>

        <!-- Members list -->
        <div class="space-y-2">
          <div v-for="m in members" :key="m.id" class="flex items-center gap-3 p-3 bg-slate-50 rounded-lg">
            <div class="w-8 h-8 bg-indigo-500 rounded-full flex items-center justify-center text-white text-xs font-bold">
              {{ (m.display_name || m.username || '?').charAt(0).toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium text-slate-900">{{ m.display_name || m.username }}</div>
              <div class="text-xs text-slate-400">{{ m.member_role }}</div>
            </div>
            <button @click="removeMember(m.user_id)" class="p-1 text-slate-400 hover:text-red-500 transition-colors">
              <UserMinus class="w-4 h-4" />
            </button>
          </div>
          <div v-if="!members.length" class="text-center text-slate-400 text-sm py-4">暂无成员</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
import { projectApi, docApi, userApi } from '@/api/index.js'
import { Plus, FileText, ChevronRight, Loader2, Lock, Users, Edit2, X, UserMinus, FolderKanban } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const loading = ref(true)
const project = ref(null)
const docs = ref([])
const members = ref([])
const activeDoc = ref(null)
const showNewDoc = ref(false)
const newDocTitle = ref('')
const showMemberModal = ref(false)
const allUsers = ref([])
const addMemberUID = ref('')

const availableUsers = computed(() => {
  const memberIds = new Set(members.value.map(m => m.user_id))
  return allUsers.value.filter(u => !memberIds.has(u.id) && u.status === 'active')
})

import { computed } from 'vue'

function statusBadge(s) {
  const map = { draft: 'badge badge-slate', active: 'badge badge-blue', in_progress: 'badge badge-amber', done: 'badge badge-green', completed: 'badge badge-green', archived: 'badge badge-slate' }
  return map[s] || 'badge badge-slate'
}
function statusLabel(s) {
  const map = { draft: '草稿', active: '进行中', in_progress: '进行中', done: '已完成', completed: '已完成', archived: '已归档' }
  return map[s] || s
}
function parseTags(tags) {
  try { return JSON.parse(tags) || [] } catch { return [] }
}
function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

async function selectDoc(doc) {
  activeDoc.value = doc
  if (!doc.content) {
    try {
      const res = await docApi.get(route.params.id, doc.id)
      activeDoc.value = res.data.data
      const idx = docs.value.findIndex(d => d.id === doc.id)
      if (idx !== -1) docs.value[idx] = res.data.data
    } catch {}
  }
}

async function createDoc() {
  if (!newDocTitle.value.trim()) return
  try {
    const res = await docApi.create(route.params.id, { title: newDocTitle.value.trim(), content: '', sort_order: docs.value.length })
    await loadDocs()
    const newDoc = docs.value.find(d => d.id === res.data.data.id)
    if (newDoc) activeDoc.value = newDoc
    showNewDoc.value = false
    newDocTitle.value = ''
  } catch {}
}

async function confirmDeleteDoc(doc) {
  if (!confirm(`确定删除文档「${doc.title}」吗？`)) return
  await docApi.delete(route.params.id, doc.id)
  if (activeDoc.value?.id === doc.id) activeDoc.value = null
  await loadDocs()
}

async function addMember() {
  if (!addMemberUID.value) return
  try {
    await projectApi.addMember(route.params.id, { user_id: parseInt(addMemberUID.value), member_role: 'executor' })
    await loadMembers()
    addMemberUID.value = ''
  } catch {}
}

async function removeMember(uid) {
  try {
    await projectApi.removeMember(route.params.id, uid)
    await loadMembers()
  } catch {}
}

async function loadDocs() {
  const res = await docApi.list(route.params.id)
  docs.value = res.data.data || []
}

async function loadMembers() {
  const res = await projectApi.listMembers(route.params.id)
  members.value = res.data.data || []
}

onMounted(async () => {
  try {
    const [pRes, dRes, mRes] = await Promise.all([
      projectApi.get(route.params.id),
      docApi.list(route.params.id),
      projectApi.listMembers(route.params.id)
    ])
    project.value = pRes.data.data
    docs.value = dRes.data.data || []
    members.value = mRes.data.data || []
    if (docs.value.length) selectDoc(docs.value[0])
    if (auth.isDev) {
      const uRes = await userApi.list().catch(() => ({ data: { data: [] } }))
      allUsers.value = uRes.data.data || []
    }
  } catch (e) {
    project.value = null
  } finally {
    loading.value = false
  }
})
</script>
