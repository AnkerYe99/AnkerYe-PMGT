<template>
  <div class="p-6 max-w-6xl mx-auto">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <button @click="router.back()" class="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors">
          <ArrowLeft class="w-5 h-5" />
        </button>
        <div>
          <h1 class="text-xl font-bold text-slate-900">{{ isNew ? '新建项目' : '编辑项目' }}</h1>
          <p class="text-slate-500 text-sm">{{ isNew ? '创建一个新的项目' : '修改项目信息和文档' }}</p>
        </div>
      </div>
      <div class="flex gap-3">
        <button @click="router.back()" class="btn-secondary">取消</button>
        <button @click="saveProject" :disabled="saving" class="btn-primary">
          <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
          <Save v-else class="w-4 h-4" />
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 bg-slate-100 p-1 rounded-xl mb-6 w-fit">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :disabled="tab.id === 'docs' && isNew"
        :class="[
          'px-4 py-2 rounded-lg text-sm font-medium transition-all',
          activeTab === tab.id ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed'
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Tab: Basic Info -->
    <div v-if="activeTab === 'info'" class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main form -->
      <div class="lg:col-span-2 space-y-5">
        <div class="card p-6 space-y-5">
          <div>
            <label class="label">项目标题 <span class="text-red-500">*</span></label>
            <input v-model="form.title" type="text" placeholder="输入项目标题" class="input" />
          </div>
          <div>
            <label class="label">项目描述</label>
            <textarea v-model="form.summary" rows="3" placeholder="简短描述项目目标和内容..." class="input resize-none"></textarea>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label">项目类型</label>
              <select v-model="form.type" class="input">
                <option value="general">通用</option>
                <option value="dev">开发</option>
                <option value="design">设计</option>
                <option value="marketing">营销</option>
                <option value="ops">运营</option>
                <option value="research">研究</option>
              </select>
            </div>
            <div>
              <label class="label">项目状态</label>
              <select v-model="form.status" class="input">
                <option value="draft">草稿</option>
                <option value="active">进行中</option>
                <option value="done">已完成</option>
                <option value="archived">已归档</option>
              </select>
            </div>
          </div>
          <div>
            <label class="label">标签</label>
            <div class="flex flex-wrap gap-2 mb-2">
              <span
                v-for="tag in tags"
                :key="tag"
                class="inline-flex items-center gap-1 bg-indigo-50 text-indigo-700 px-3 py-1 rounded-full text-sm"
              >
                {{ tag }}
                <button @click="removeTag(tag)" class="hover:text-red-500"><X class="w-3 h-3" /></button>
              </span>
            </div>
            <div class="flex gap-2">
              <input
                v-model="newTag"
                @keydown.enter.prevent="addTag"
                type="text"
                placeholder="输入标签后按 Enter"
                class="input flex-1"
                maxlength="20"
              />
              <button @click="addTag" class="btn-secondary">添加</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar: Color -->
      <div class="space-y-5">
        <div class="card p-6">
          <label class="label mb-3">封面颜色</label>
          <!-- Preview -->
          <div class="h-20 rounded-xl mb-4 transition-all duration-200" :style="{ backgroundColor: form.cover_color }"></div>
          <div class="grid grid-cols-4 gap-2">
            <button
              v-for="color in presetColors"
              :key="color"
              @click="form.cover_color = color"
              :class="['w-full aspect-square rounded-lg border-2 transition-all', form.cover_color === color ? 'border-slate-400 scale-95' : 'border-transparent hover:scale-95']"
              :style="{ backgroundColor: color }"
            ></button>
          </div>
          <div class="mt-3">
            <label class="label text-xs">自定义颜色</label>
            <input v-model="form.cover_color" type="color" class="w-full h-10 rounded-lg cursor-pointer border border-slate-200" />
          </div>
        </div>

        <div class="card p-6">
          <label class="label mb-3">项目预览</label>
          <div class="rounded-xl border border-slate-200 overflow-hidden">
            <div class="h-1.5" :style="{ backgroundColor: form.cover_color }"></div>
            <div class="p-4">
              <div class="font-semibold text-slate-900 mb-1 truncate">{{ form.title || '项目标题' }}</div>
              <div class="text-xs text-slate-400 line-clamp-2">{{ form.summary || '项目描述' }}</div>
              <div class="flex gap-2 mt-3">
                <span class="badge badge-slate text-xs">{{ typeLabel(form.type) }}</span>
                <span :class="statusBadge(form.status)">{{ statusLabel(form.status) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab: Docs -->
    <div v-if="activeTab === 'docs'" class="flex gap-6 h-[calc(100vh-240px)]">
      <!-- Doc list -->
      <div class="w-56 card flex flex-col overflow-hidden">
        <div class="px-4 py-3 border-b border-slate-100 flex items-center justify-between">
          <span class="text-sm font-semibold text-slate-700">文档</span>
          <button @click="createNewDoc" class="p-1 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded">
            <Plus class="w-4 h-4" />
          </button>
        </div>
        <div class="flex-1 overflow-y-auto py-1">
          <div v-if="!docs.length" class="text-center text-slate-400 text-sm py-8">暂无文档</div>
          <div
            v-for="doc in docs"
            :key="doc.id"
            @click="selectDoc(doc)"
            :class="['flex items-center gap-2 px-4 py-2.5 cursor-pointer transition-colors group', activeDoc?.id === doc.id ? 'bg-indigo-50 text-indigo-700' : 'text-slate-600 hover:bg-slate-50']"
          >
            <FileText class="w-3.5 h-3.5 flex-shrink-0 opacity-60" />
            <span class="text-sm truncate flex-1">{{ doc.title }}</span>
            <button @click.stop="deleteDoc(doc)" class="text-slate-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-all">
              <Trash2 class="w-3 h-3" />
            </button>
          </div>
        </div>
      </div>

      <!-- Editor -->
      <div class="flex-1 card flex flex-col overflow-hidden">
        <div v-if="!activeDoc" class="flex-1 flex items-center justify-center text-slate-400">
          <div class="text-center">
            <FileText class="w-12 h-12 text-slate-200 mx-auto mb-3" />
            <p>请选择或创建文档</p>
          </div>
        </div>
        <template v-else>
          <div class="px-4 py-3 border-b border-slate-100 flex items-center gap-3">
            <input v-model="activeDoc.title" @change="saveDoc" type="text" class="flex-1 text-base font-semibold text-slate-900 bg-transparent focus:outline-none" placeholder="文档标题" />
            <button @click="saveDoc" :disabled="docSaving" class="btn-primary text-xs py-1.5">
              <Loader2 v-if="docSaving" class="w-3.5 h-3.5 animate-spin" />
              <Save v-else class="w-3.5 h-3.5" />
              {{ docSaving ? '保存中' : '保存文档' }}
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-4">
            <Editor v-model="activeDoc.content" />
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { projectApi, docApi } from '@/api/index.js'
import Editor from '@/components/Editor.vue'
import { ArrowLeft, Save, Plus, FileText, Trash2, X, Loader2 } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id || route.params.id === 'new')
const activeTab = ref('info')
const saving = ref(false)
const docSaving = ref(false)

const tabs = [
  { id: 'info', label: '基本信息' },
  { id: 'docs', label: '文档编辑' }
]

const presetColors = [
  '#6366F1', '#8B5CF6', '#EC4899', '#EF4444',
  '#F59E0B', '#10B981', '#06B6D4', '#3B82F6'
]

const form = ref({
  title: '',
  summary: '',
  type: 'general',
  status: 'draft',
  cover_color: '#6366F1',
  tags: '[]'
})

const tags = ref([])
const newTag = ref('')
const docs = ref([])
const activeDoc = ref(null)

function typeLabel(t) {
  const map = { general: '通用', dev: '开发', design: '设计', marketing: '营销', ops: '运营', research: '研究' }
  return map[t] || t
}
function statusBadge(s) {
  const map = { draft: 'badge badge-slate', active: 'badge badge-blue', in_progress: 'badge badge-amber', done: 'badge badge-green', archived: 'badge badge-slate' }
  return map[s] || 'badge badge-slate'
}
function statusLabel(s) {
  const map = { draft: '草稿', active: '进行中', in_progress: '进行中', done: '已完成', archived: '已归档' }
  return map[s] || s
}

function addTag() {
  const t = newTag.value.trim()
  if (t && !tags.value.includes(t)) tags.value.push(t)
  newTag.value = ''
}
function removeTag(t) {
  tags.value = tags.value.filter(x => x !== t)
}

async function saveProject() {
  if (!form.value.title.trim()) {
    alert('项目标题不能为空')
    return
  }
  saving.value = true
  try {
    const payload = { ...form.value, tags: JSON.stringify(tags.value) }
    if (isNew.value) {
      const res = await projectApi.create(payload)
      const newId = res.data.data.id
      router.replace('/projects/' + newId + '/edit')
    } else {
      await projectApi.update(route.params.id, payload)
    }
  } catch (e) {
    alert(e.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function createNewDoc() {
  const title = prompt('文档标题')
  if (!title?.trim()) return
  try {
    const res = await docApi.create(route.params.id, { title: title.trim(), content: '', sort_order: docs.value.length })
    await loadDocs()
    const nd = docs.value.find(d => d.id === res.data.data.id)
    if (nd) activeDoc.value = { ...nd }
  } catch {}
}

async function selectDoc(doc) {
  if (activeDoc.value && activeDoc.value.id !== doc.id) {
    await saveDoc()
  }
  const res = await docApi.get(route.params.id, doc.id)
  activeDoc.value = { ...res.data.data }
}

async function saveDoc() {
  if (!activeDoc.value) return
  docSaving.value = true
  try {
    await docApi.update(route.params.id, activeDoc.value.id, {
      title: activeDoc.value.title,
      content: activeDoc.value.content
    })
    const idx = docs.value.findIndex(d => d.id === activeDoc.value.id)
    if (idx !== -1) docs.value[idx].title = activeDoc.value.title
  } catch {} finally {
    docSaving.value = false
  }
}

async function deleteDoc(doc) {
  if (!confirm(`确定删除文档「${doc.title}」吗？`)) return
  await docApi.delete(route.params.id, doc.id)
  if (activeDoc.value?.id === doc.id) activeDoc.value = null
  await loadDocs()
}

async function loadDocs() {
  if (isNew.value) return
  const res = await docApi.list(route.params.id)
  docs.value = res.data.data || []
}

onMounted(async () => {
  if (!isNew.value) {
    try {
      const res = await projectApi.get(route.params.id)
      const p = res.data.data
      form.value = {
        title: p.title,
        summary: p.summary || '',
        type: p.type,
        status: p.status,
        cover_color: p.cover_color,
        tags: p.tags
      }
      try { tags.value = JSON.parse(p.tags) || [] } catch { tags.value = [] }
      await loadDocs()
    } catch {}
  }
})
</script>
