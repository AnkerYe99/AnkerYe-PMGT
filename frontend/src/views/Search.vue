<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-slate-900">全文搜索</h1>
      <p class="text-slate-500 text-sm mt-0.5">搜索项目和文档内容</p>
    </div>

    <!-- Search Box -->
    <div class="card p-4 mb-6">
      <div class="relative">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
        <input
          v-model="q"
          @input="debouncedSearch"
          type="text"
          placeholder="输入关键词搜索项目、文档..."
          class="w-full pl-12 pr-4 py-3 border border-slate-200 rounded-xl text-base focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
          autofocus
        />
        <button v-if="q" @click="q = ''; results = null" class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600">
          <X class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="searching" class="text-center py-12 text-slate-400">
      <Loader2 class="w-8 h-8 animate-spin mx-auto mb-3 text-indigo-400" />
      搜索中...
    </div>

    <!-- No query -->
    <div v-else-if="!q" class="text-center py-16">
      <Search class="w-16 h-16 text-slate-200 mx-auto mb-4" />
      <p class="text-slate-400 font-medium">请输入关键词开始搜索</p>
      <p class="text-slate-300 text-sm mt-1">支持搜索项目名称、摘要、文档标题和内容</p>
    </div>

    <!-- No results -->
    <div v-else-if="results && !results.results?.length" class="text-center py-16">
      <SearchX class="w-16 h-16 text-slate-200 mx-auto mb-4" />
      <p class="text-slate-400 font-medium">未找到「{{ q }}」相关内容</p>
      <p class="text-slate-300 text-sm mt-1">试试其他关键词</p>
    </div>

    <!-- Results -->
    <div v-else-if="results">
      <div class="flex items-center gap-2 mb-4 text-sm text-slate-500">
        <span>找到 <span class="font-semibold text-slate-900">{{ results.total }}</span> 个结果</span>
        <span>·</span>
        <span>关键词：<span class="text-indigo-600 font-medium">{{ results.query }}</span></span>
      </div>

      <!-- Projects -->
      <div v-if="projectResults.length" class="mb-6">
        <h2 class="text-sm font-semibold text-slate-500 uppercase tracking-wider mb-3 flex items-center gap-2">
          <FolderKanban class="w-4 h-4" />
          项目 ({{ projectResults.length }})
        </h2>
        <div class="space-y-2">
          <div
            v-for="r in projectResults"
            :key="'p-' + r.id"
            @click="router.push('/projects/' + r.id)"
            class="card p-4 cursor-pointer hover:shadow-md transition-all hover:border-indigo-200 group"
          >
            <div class="flex items-start gap-3">
              <div class="w-9 h-9 bg-indigo-100 rounded-xl flex items-center justify-center flex-shrink-0 group-hover:bg-indigo-200 transition-colors">
                <FolderKanban class="w-4 h-4 text-indigo-600" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-slate-900 group-hover:text-indigo-600 transition-colors" v-html="highlight(r.title)"></div>
                <div v-if="r.snippet" class="text-sm text-slate-500 mt-0.5 line-clamp-2" v-html="highlight(r.snippet)"></div>
              </div>
              <ArrowRight class="w-4 h-4 text-slate-300 group-hover:text-indigo-400 transition-colors flex-shrink-0 mt-1" />
            </div>
          </div>
        </div>
      </div>

      <!-- Documents -->
      <div v-if="docResults.length">
        <h2 class="text-sm font-semibold text-slate-500 uppercase tracking-wider mb-3 flex items-center gap-2">
          <FileText class="w-4 h-4" />
          文档 ({{ docResults.length }})
        </h2>
        <div class="space-y-2">
          <div
            v-for="r in docResults"
            :key="'d-' + r.id"
            @click="router.push('/projects/' + r.project_id)"
            class="card p-4 cursor-pointer hover:shadow-md transition-all hover:border-blue-200 group"
          >
            <div class="flex items-start gap-3">
              <div class="w-9 h-9 bg-blue-100 rounded-xl flex items-center justify-center flex-shrink-0 group-hover:bg-blue-200 transition-colors">
                <FileText class="w-4 h-4 text-blue-600" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-slate-900 group-hover:text-blue-600 transition-colors" v-html="highlight(r.title)"></div>
                <div v-if="r.snippet" class="text-sm text-slate-500 mt-0.5 line-clamp-3 search-snippet" v-html="r.snippet"></div>
                <div class="text-xs text-slate-400 mt-1">项目 #{{ r.project_id }}</div>
              </div>
              <ArrowRight class="w-4 h-4 text-slate-300 group-hover:text-blue-400 transition-colors flex-shrink-0 mt-1" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { searchApi } from '@/api/index.js'
import { Search, X, Loader2, FolderKanban, FileText, ArrowRight, SearchX } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const q = ref(route.query.q || '')
const searching = ref(false)
const results = ref(null)
let timer = null

const projectResults = computed(() => results.value?.results?.filter(r => r.type === 'project') || [])
const docResults = computed(() => results.value?.results?.filter(r => r.type === 'document') || [])

function highlight(text) {
  if (!q.value || !text) return text || ''
  const escaped = q.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return text.replace(new RegExp(`(${escaped})`, 'gi'), '<mark class="bg-yellow-200 text-yellow-900 rounded px-0.5">$1</mark>')
}

function debouncedSearch() {
  clearTimeout(timer)
  if (!q.value.trim()) { results.value = null; return }
  timer = setTimeout(doSearch, 400)
}

async function doSearch() {
  if (!q.value.trim()) return
  searching.value = true
  try {
    const res = await searchApi.search(q.value)
    results.value = res.data.data
  } finally {
    searching.value = false
  }
}

onMounted(() => {
  if (q.value) doSearch()
})

watch(() => route.query.q, (val) => {
  if (val && val !== q.value) {
    q.value = val
    doSearch()
  }
})
</script>

<style>
.search-snippet mark {
  background-color: #fef08a;
  color: #713f12;
  border-radius: 2px;
  padding: 0 2px;
}
</style>
