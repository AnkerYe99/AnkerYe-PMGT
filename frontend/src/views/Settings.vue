<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-slate-900">系统设置</h1>
      <p class="text-slate-500 text-sm mt-0.5">管理系统配置、API Keys 和数据备份</p>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 bg-slate-100 p-1 rounded-xl mb-6 w-fit">
      <button v-for="tab in tabs" :key="tab.id" @click="activeTab = tab.id"
        :class="['px-4 py-2 rounded-lg text-sm font-medium transition-all flex items-center gap-2', activeTab === tab.id ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700']">
        <component :is="tab.icon" class="w-4 h-4" />
        {{ tab.label }}
      </button>
    </div>

    <!-- General Settings -->
    <div v-if="activeTab === 'general'" class="max-w-lg space-y-5">
      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-slate-900">基本设置</h2>
        <div>
          <label class="label">系统名称</label>
          <input v-model="settings.site_name" type="text" class="input" placeholder="AnkerYe PMGT" />
        </div>
        <div>
          <label class="label">系统描述</label>
          <input v-model="settings.site_description" type="text" class="input" placeholder="项目管理系统" />
        </div>
        <div>
          <label class="label">管理员邮箱</label>
          <input v-model="settings.admin_email" type="email" class="input" placeholder="admin@example.com" />
        </div>
        <button @click="saveSettings" :disabled="settingSaving" class="btn-primary w-full">
          <Loader2 v-if="settingSaving" class="w-4 h-4 animate-spin" />
          <Save v-else class="w-4 h-4" />
          {{ settingSaving ? '保存中...' : '保存设置' }}
        </button>
      </div>

      <!-- Change Password -->
      <div class="card p-6 space-y-4">
        <h2 class="font-semibold text-slate-900">修改密码</h2>
        <div>
          <label class="label">当前密码</label>
          <input v-model="pwdForm.old_password" type="password" class="input" />
        </div>
        <div>
          <label class="label">新密码</label>
          <input v-model="pwdForm.new_password" type="password" class="input" placeholder="至少6位" />
        </div>
        <button @click="changePassword" :disabled="pwdSaving" class="btn-primary w-full">
          <Loader2 v-if="pwdSaving" class="w-4 h-4 animate-spin" />
          修改密码
        </button>
      </div>
    </div>

    <!-- API Keys -->
    <div v-if="activeTab === 'apikeys'">
      <div class="flex items-center justify-between mb-4">
        <p class="text-sm text-slate-500">API Key 用于第三方系统访问，创建后完整 Key 只显示一次。</p>
        <button @click="showCreateKey = true" class="btn-primary">
          <Plus class="w-4 h-4" />
          新建 API Key
        </button>
      </div>

      <!-- New Key Result -->
      <div v-if="newKeyResult" class="card p-4 mb-4 border-2 border-emerald-200 bg-emerald-50">
        <div class="flex items-start gap-3">
          <div class="w-8 h-8 bg-emerald-100 rounded-lg flex items-center justify-center flex-shrink-0">
            <Key class="w-4 h-4 text-emerald-600" />
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-semibold text-emerald-800 mb-1">API Key 创建成功！请立即复制，此后不再显示。</p>
            <div class="flex items-center gap-2">
              <code class="flex-1 text-sm bg-white border border-emerald-200 rounded px-3 py-2 font-mono break-all text-emerald-800">{{ newKeyResult }}</code>
              <button @click="copyKey" class="btn-secondary text-xs py-1.5 flex-shrink-0">
                <Copy class="w-3.5 h-3.5" />
                复制
              </button>
            </div>
          </div>
        </div>
        <button @click="newKeyResult = null" class="mt-3 text-sm text-emerald-600 hover:text-emerald-700">我已复制，关闭提示</button>
      </div>

      <div class="card overflow-hidden">
        <table class="w-full">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-100">
              <th class="text-left px-6 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">名称</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">前缀</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">权限</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider hidden md:table-cell">最后使用</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">状态</th>
              <th class="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr v-for="k in apiKeys" :key="k.id" class="hover:bg-slate-50">
              <td class="px-6 py-4 font-medium text-slate-900">{{ k.name }}</td>
              <td class="px-4 py-4"><code class="text-xs bg-slate-100 px-2 py-1 rounded font-mono">{{ k.key_prefix }}...</code></td>
              <td class="px-4 py-4"><span class="badge badge-blue">{{ k.permissions }}</span></td>
              <td class="px-4 py-4 text-sm text-slate-400 hidden md:table-cell">{{ k.last_used_at ? formatDate(k.last_used_at) : '从未' }}</td>
              <td class="px-4 py-4"><span :class="k.status === 'active' ? 'badge badge-green' : 'badge badge-red'">{{ k.status === 'active' ? '活跃' : '已撤销' }}</span></td>
              <td class="px-4 py-4">
                <button @click="revokeKey(k.id)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!apiKeys.length" class="p-8 text-center text-slate-400">暂无 API Key</div>
      </div>
    </div>

    <!-- Backup -->
    <div v-if="activeTab === 'backup'" class="max-w-lg space-y-5">
      <div class="card p-6">
        <h2 class="font-semibold text-slate-900 mb-1">数据备份</h2>
        <p class="text-sm text-slate-500 mb-5">将所有数据打包加密下载，备份文件格式为 .bak。</p>
        <a href="/api/v1/settings/backup" download class="btn-primary w-full justify-center">
          <Download class="w-4 h-4" />
          下载备份文件
        </a>
      </div>
      <div class="card p-6">
        <h2 class="font-semibold text-slate-900 mb-1">数据恢复</h2>
        <p class="text-sm text-slate-500 mb-5 bg-amber-50 rounded-lg p-3 border border-amber-100">
          <span class="font-semibold text-amber-700">注意：</span>恢复操作将解析备份文件，当前数据不会自动覆盖，请联系管理员手动操作。
        </p>
        <div class="border-2 border-dashed border-slate-200 rounded-xl p-6 text-center hover:border-indigo-300 transition-colors cursor-pointer" @click="$refs.backupFile.click()">
          <Upload class="w-8 h-8 text-slate-300 mx-auto mb-2" />
          <p class="text-sm text-slate-500">点击选择备份文件 (.bak)</p>
        </div>
        <input ref="backupFile" type="file" accept=".bak" class="hidden" @change="restoreBackup" />
        <div v-if="restoreResult" class="mt-3 p-3 bg-emerald-50 rounded-lg border border-emerald-200 text-sm text-emerald-700">
          {{ restoreResult }}
        </div>
      </div>
    </div>

    <!-- Update -->
    <div v-if="activeTab === 'update'" class="max-w-lg space-y-5">
      <div class="card p-6">
        <h2 class="font-semibold text-slate-900 mb-4">当前版本</h2>
        <div class="flex items-center gap-4 p-4 bg-slate-50 rounded-xl">
          <div class="w-12 h-12 bg-indigo-100 rounded-xl flex items-center justify-center">
            <Package class="w-6 h-6 text-indigo-600" />
          </div>
          <div>
            <div class="font-semibold text-slate-900">AnkerYe PMGT</div>
            <div class="text-sm text-slate-500">版本 {{ currentVersion }}</div>
          </div>
        </div>

        <div class="mt-4 flex gap-3">
          <button @click="checkUpdate" :disabled="checkingUpdate" class="flex-1 btn-secondary">
            <Loader2 v-if="checkingUpdate" class="w-4 h-4 animate-spin" />
            <RefreshCw v-else class="w-4 h-4" />
            {{ checkingUpdate ? '检查中...' : '检查更新' }}
          </button>
          <button v-if="updateInfo?.has_update" @click="applyUpdate" :disabled="applyingUpdate" class="flex-1 btn-primary">
            <Loader2 v-if="applyingUpdate" class="w-4 h-4 animate-spin" />
            <ArrowUpCircle v-else class="w-4 h-4" />
            {{ applyingUpdate ? '升级中...' : '立即升级' }}
          </button>
        </div>

        <div v-if="updateInfo" class="mt-4">
          <div v-if="updateInfo.has_update" class="p-4 bg-indigo-50 rounded-xl border border-indigo-200">
            <div class="font-semibold text-indigo-800 mb-1">发现新版本 {{ updateInfo.latest_version }}</div>
            <pre v-if="updateInfo.changelog" class="text-xs text-indigo-700 whitespace-pre-wrap">{{ updateInfo.changelog }}</pre>
          </div>
          <div v-else class="p-4 bg-emerald-50 rounded-xl border border-emerald-200 flex items-center gap-3">
            <CheckCircle2 class="w-5 h-5 text-emerald-600 flex-shrink-0" />
            <span class="text-emerald-700 text-sm">当前已是最新版本</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Key Modal -->
    <div v-if="showCreateKey" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showCreateKey = false">
      <div class="bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm">
        <h3 class="font-semibold text-slate-900 mb-5">新建 API Key</h3>
        <div class="space-y-4">
          <div>
            <label class="label">名称 <span class="text-red-500">*</span></label>
            <input v-model="keyForm.name" type="text" class="input" placeholder="用途描述" />
          </div>
          <div>
            <label class="label">权限</label>
            <select v-model="keyForm.permissions" class="input">
              <option value="read">只读</option>
              <option value="read,write">读写</option>
              <option value="all">完全访问</option>
            </select>
          </div>
          <div>
            <label class="label">过期时间（留空永不过期）</label>
            <input v-model="keyForm.expires_at" type="date" class="input" />
          </div>
        </div>
        <div class="flex gap-3 mt-6">
          <button @click="showCreateKey = false" class="flex-1 btn-secondary">取消</button>
          <button @click="createAPIKey" :disabled="!keyForm.name || keySaving" class="flex-1 btn-primary">
            <Loader2 v-if="keySaving" class="w-4 h-4 animate-spin" />
            创建
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { settingsApi, apiKeyApi, updateApi } from '@/api/index.js'
import api from '@/api/index.js'
import { Save, Plus, Key, Copy, Trash2, Download, Upload, Package, RefreshCw, ArrowUpCircle, CheckCircle2, Loader2, Settings as SettingsIcon, Database, Shield } from 'lucide-vue-next'
import { authApi } from '@/api/index.js'

const activeTab = ref('general')
const tabs = [
  { id: 'general', label: '常规', icon: SettingsIcon },
  { id: 'apikeys', label: 'API Keys', icon: Shield },
  { id: 'backup', label: '备份恢复', icon: Database },
  { id: 'update', label: '在线升级', icon: Package }
]

// Settings
const settings = ref({ site_name: '', site_description: '', admin_email: '' })
const settingSaving = ref(false)
const pwdForm = ref({ old_password: '', new_password: '' })
const pwdSaving = ref(false)

// API Keys
const apiKeys = ref([])
const showCreateKey = ref(false)
const keyForm = ref({ name: '', permissions: 'read', expires_at: '' })
const keySaving = ref(false)
const newKeyResult = ref(null)

// Backup
const restoreResult = ref(null)

// Update
const currentVersion = ref('1.0.0')
const checkingUpdate = ref(false)
const applyingUpdate = ref(false)
const updateInfo = ref(null)

function formatDate(d) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('zh-CN')
}

async function saveSettings() {
  settingSaving.value = true
  try {
    await settingsApi.update(settings.value)
  } finally {
    settingSaving.value = false
  }
}

async function changePassword() {
  if (!pwdForm.value.old_password || !pwdForm.value.new_password) {
    alert('请填写完整')
    return
  }
  pwdSaving.value = true
  try {
    await authApi.changePassword(pwdForm.value)
    alert('密码修改成功')
    pwdForm.value = { old_password: '', new_password: '' }
  } catch (e) {
    alert(e.response?.data?.message || '修改失败')
  } finally {
    pwdSaving.value = false
  }
}

async function createAPIKey() {
  if (!keyForm.value.name) return
  keySaving.value = true
  try {
    const res = await apiKeyApi.create(keyForm.value)
    newKeyResult.value = res.data.data.key
    showCreateKey.value = false
    keyForm.value = { name: '', permissions: 'read', expires_at: '' }
    await loadAPIKeys()
  } finally {
    keySaving.value = false
  }
}

async function revokeKey(id) {
  if (!confirm('确定撤销该 API Key？')) return
  await apiKeyApi.delete(id)
  await loadAPIKeys()
}

function copyKey() {
  navigator.clipboard?.writeText(newKeyResult.value)
  alert('已复制到剪贴板')
}

async function restoreBackup(e) {
  const file = e.target.files[0]
  if (!file) return
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await api.post('/settings/restore', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
    restoreResult.value = `解析成功：备份时间 ${res.data.data?.timestamp}，版本 ${res.data.data?.version}`
  } catch (err) {
    restoreResult.value = '恢复失败：' + (err.response?.data?.message || '未知错误')
  }
  e.target.value = ''
}

async function checkUpdate() {
  checkingUpdate.value = true
  try {
    const res = await updateApi.check()
    updateInfo.value = res.data.data
  } finally {
    checkingUpdate.value = false
  }
}

async function applyUpdate() {
  if (!confirm('确定要升级系统？服务将短暂重启。')) return
  applyingUpdate.value = true
  try {
    await updateApi.apply()
    alert('升级指令已发送，服务即将重启')
  } catch (e) {
    alert(e.response?.data?.message || '升级失败')
  } finally {
    applyingUpdate.value = false
  }
}

async function loadAPIKeys() {
  const res = await apiKeyApi.list()
  apiKeys.value = res.data.data || []
}

onMounted(async () => {
  const [sRes, vRes] = await Promise.allSettled([
    settingsApi.get(),
    updateApi.version()
  ])
  if (sRes.status === 'fulfilled') settings.value = { ...settings.value, ...sRes.value.data.data }
  if (vRes.status === 'fulfilled') currentVersion.value = vRes.value.data.data?.version || '1.0.0'
  await loadAPIKeys()
})
</script>
