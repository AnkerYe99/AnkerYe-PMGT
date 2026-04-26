<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-900">用户管理</h1>
        <p class="text-slate-500 text-sm mt-0.5">管理系统所有用户账号</p>
      </div>
      <button @click="openCreateModal" class="btn-primary">
        <UserPlus class="w-4 h-4" />
        新建用户
      </button>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div class="card p-4">
        <div class="text-2xl font-bold text-slate-900">{{ users.length }}</div>
        <div class="text-sm text-slate-500">总用户数</div>
      </div>
      <div class="card p-4">
        <div class="text-2xl font-bold text-emerald-600">{{ activeCount }}</div>
        <div class="text-sm text-slate-500">活跃用户</div>
      </div>
      <div class="card p-4">
        <div class="text-2xl font-bold text-indigo-600">{{ adminCount }}</div>
        <div class="text-sm text-slate-500">管理员</div>
      </div>
      <div class="card p-4">
        <div class="text-2xl font-bold text-amber-600">{{ devCount }}</div>
        <div class="text-sm text-slate-500">开发者</div>
      </div>
    </div>

    <!-- Table -->
    <div class="card overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center gap-4">
        <div class="relative flex-1 max-w-xs">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="searchQ" type="text" placeholder="搜索用户..." class="input pl-9" />
        </div>
        <select v-model="roleFilter" class="input w-32">
          <option value="">所有角色</option>
          <option value="admin">管理员</option>
          <option value="developer">开发者</option>
          <option value="executor">执行者</option>
        </select>
      </div>

      <div v-if="loading" class="p-8 text-center text-slate-400">
        <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2" />
        加载中...
      </div>
      <table v-else class="w-full">
        <thead>
          <tr class="bg-slate-50 border-b border-slate-100">
            <th class="text-left px-6 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">用户</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider hidden md:table-cell">邮箱</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">角色</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">状态</th>
            <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider hidden lg:table-cell">最后登录</th>
            <th class="px-4 py-3 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          <tr v-for="u in filteredUsers" :key="u.id" class="hover:bg-slate-50 transition-colors">
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-full flex items-center justify-center text-white text-sm font-bold flex-shrink-0" :style="{ backgroundColor: avatarColor(u.username) }">
                  {{ (u.display_name || u.username || '?').charAt(0).toUpperCase() }}
                </div>
                <div>
                  <div class="font-medium text-slate-900">{{ u.display_name || u.username }}</div>
                  <div class="text-xs text-slate-400">@{{ u.username }}</div>
                </div>
              </div>
            </td>
            <td class="px-4 py-4 text-sm text-slate-500 hidden md:table-cell">{{ u.email || '-' }}</td>
            <td class="px-4 py-4">
              <select
                :value="u.role"
                @change="updateRole(u, $event.target.value)"
                class="text-xs border border-slate-200 rounded-lg px-2 py-1 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                :disabled="u.id === auth.user?.id"
              >
                <option value="admin">管理员</option>
                <option value="developer">开发者</option>
                <option value="executor">执行者</option>
              </select>
            </td>
            <td class="px-4 py-4">
              <button @click="toggleStatus(u)" :disabled="u.id === auth.user?.id" :class="u.status === 'active' ? 'badge badge-green cursor-pointer hover:bg-emerald-200' : 'badge badge-red cursor-pointer hover:bg-red-200'">
                {{ u.status === 'active' ? '活跃' : u.status === 'deleted' ? '已删除' : '禁用' }}
              </button>
            </td>
            <td class="px-4 py-4 text-sm text-slate-400 hidden lg:table-cell">{{ formatDate(u.last_login_at) }}</td>
            <td class="px-4 py-4 text-right">
              <div class="flex items-center justify-end gap-1">
                <button @click="openEditModal(u)" class="p-1.5 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded transition-colors" title="编辑">
                  <Edit2 class="w-3.5 h-3.5" />
                </button>
                <button v-if="u.id !== auth.user?.id" @click="deleteUser(u)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors" title="删除">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!loading && !filteredUsers.length" class="p-8 text-center text-slate-400">
        暂无匹配用户
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showModal = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-md">
        <div class="flex items-center justify-between mb-5">
          <h3 class="font-semibold text-slate-900">{{ modalMode === 'create' ? '新建用户' : '编辑用户' }}</h3>
          <button @click="showModal = false" class="p-1 text-slate-400 hover:text-slate-600 rounded">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="space-y-4">
          <div v-if="modalMode === 'create'">
            <label class="label">用户名 <span class="text-red-500">*</span></label>
            <input v-model="modalForm.username" type="text" class="input" placeholder="用户名（唯一）" />
          </div>
          <div>
            <label class="label">显示名</label>
            <input v-model="modalForm.display_name" type="text" class="input" placeholder="显示名称" />
          </div>
          <div>
            <label class="label">邮箱</label>
            <input v-model="modalForm.email" type="email" class="input" placeholder="邮箱地址" />
          </div>
          <div>
            <label class="label">角色</label>
            <select v-model="modalForm.role" class="input">
              <option value="executor">执行者</option>
              <option value="developer">开发者</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div>
            <label class="label">{{ modalMode === 'create' ? '密码 *' : '新密码（留空不修改）' }}</label>
            <input v-model="modalForm.password" type="password" class="input" :placeholder="modalMode === 'create' ? '至少6位' : '留空则不修改'" />
          </div>
        </div>
        <div class="flex gap-3 mt-6">
          <button @click="showModal = false" class="flex-1 btn-secondary">取消</button>
          <button @click="submitModal" :disabled="modalSaving" class="flex-1 btn-primary">
            <Loader2 v-if="modalSaving" class="w-4 h-4 animate-spin" />
            {{ modalSaving ? '保存中...' : '确认' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth.js'
import { userApi } from '@/api/index.js'
import { UserPlus, Search, Edit2, Trash2, X, Loader2 } from 'lucide-vue-next'

const auth = useAuthStore()
const loading = ref(true)
const users = ref([])
const searchQ = ref('')
const roleFilter = ref('')
const showModal = ref(false)
const modalMode = ref('create')
const modalSaving = ref(false)
const editTarget = ref(null)

const modalForm = ref({ username: '', display_name: '', email: '', role: 'executor', password: '' })

const activeCount = computed(() => users.value.filter(u => u.status === 'active').length)
const adminCount = computed(() => users.value.filter(u => u.role === 'admin').length)
const devCount = computed(() => users.value.filter(u => u.role === 'developer').length)

const filteredUsers = computed(() => {
  return users.value.filter(u => {
    const q = searchQ.value.toLowerCase()
    const matchQ = !q || u.username.toLowerCase().includes(q) || (u.display_name || '').toLowerCase().includes(q)
    const matchRole = !roleFilter.value || u.role === roleFilter.value
    return matchQ && matchRole
  })
})

function avatarColor(username) {
  const colors = ['#6366F1', '#8B5CF6', '#EC4899', '#EF4444', '#F59E0B', '#10B981', '#06B6D4', '#3B82F6']
  let hash = 0
  for (let c of (username || '')) hash = (hash + c.charCodeAt(0)) % colors.length
  return colors[hash]
}

function formatDate(d) {
  if (!d) return '从未'
  return new Date(d).toLocaleDateString('zh-CN')
}

function openCreateModal() {
  modalMode.value = 'create'
  modalForm.value = { username: '', display_name: '', email: '', role: 'executor', password: '' }
  showModal.value = true
}

function openEditModal(u) {
  modalMode.value = 'edit'
  editTarget.value = u
  modalForm.value = { username: u.username, display_name: u.display_name || '', email: u.email || '', role: u.role, password: '' }
  showModal.value = true
}

async function submitModal() {
  modalSaving.value = true
  try {
    if (modalMode.value === 'create') {
      if (!modalForm.value.username || !modalForm.value.password) {
        alert('用户名和密码不能为空')
        return
      }
      await userApi.create(modalForm.value)
    } else {
      const payload = { ...modalForm.value }
      if (!payload.password) delete payload.password
      await userApi.update(editTarget.value.id, payload)
    }
    showModal.value = false
    await loadUsers()
  } catch (e) {
    alert(e.response?.data?.message || '操作失败')
  } finally {
    modalSaving.value = false
  }
}

async function updateRole(u, role) {
  try {
    await userApi.update(u.id, { role })
    u.role = role
  } catch {}
}

async function toggleStatus(u) {
  if (u.id === auth.user?.id) return
  const newStatus = u.status === 'active' ? 'disabled' : 'active'
  try {
    await userApi.update(u.id, { status: newStatus })
    u.status = newStatus
  } catch {}
}

async function deleteUser(u) {
  if (!confirm(`确定删除用户「${u.display_name || u.username}」吗？`)) return
  try {
    await userApi.delete(u.id)
    users.value = users.value.filter(x => x.id !== u.id)
  } catch {}
}

async function loadUsers() {
  loading.value = true
  try {
    const res = await userApi.list()
    users.value = res.data.data || []
  } finally {
    loading.value = false
  }
}

onMounted(loadUsers)
</script>
