<template>
  <aside class="w-64 bg-slate-900 flex flex-col h-full flex-shrink-0">
    <!-- Brand -->
    <div class="px-4 py-5 border-b border-slate-800">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 bg-indigo-600 rounded-xl flex items-center justify-center flex-shrink-0">
          <FolderKanban class="w-5 h-5 text-white" />
        </div>
        <div>
          <div class="text-white font-bold text-sm leading-tight">AnkerYe</div>
          <div class="text-slate-500 text-xs">项目管理系统</div>
        </div>
      </div>
    </div>

    <!-- Nav -->
    <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
      <div class="text-slate-600 text-xs font-semibold uppercase tracking-wider px-3 mb-2">主菜单</div>

      <router-link to="/dashboard" custom v-slot="{ isActive, navigate }">
        <div :class="isActive ? 'sidebar-item-active' : 'sidebar-item-inactive'" @click="navigate">
          <LayoutDashboard class="w-4 h-4 flex-shrink-0" />
          <span>仪表盘</span>
        </div>
      </router-link>

      <router-link to="/projects" custom v-slot="{ isActive, navigate }">
        <div :class="isActive ? 'sidebar-item-active' : 'sidebar-item-inactive'" @click="navigate">
          <FolderKanban class="w-4 h-4 flex-shrink-0" />
          <span>项目</span>
        </div>
      </router-link>

      <router-link to="/search" custom v-slot="{ isActive, navigate }">
        <div :class="isActive ? 'sidebar-item-active' : 'sidebar-item-inactive'" @click="navigate">
          <Search class="w-4 h-4 flex-shrink-0" />
          <span>全文搜索</span>
        </div>
      </router-link>

      <template v-if="auth.isAdmin">
        <div class="text-slate-600 text-xs font-semibold uppercase tracking-wider px-3 mb-2 mt-4">管理</div>

        <router-link to="/users" custom v-slot="{ isActive, navigate }">
          <div :class="isActive ? 'sidebar-item-active' : 'sidebar-item-inactive'" @click="navigate">
            <Users class="w-4 h-4 flex-shrink-0" />
            <span>用户管理</span>
          </div>
        </router-link>

        <router-link to="/settings" custom v-slot="{ isActive, navigate }">
          <div :class="isActive ? 'sidebar-item-active' : 'sidebar-item-inactive'" @click="navigate">
            <Settings class="w-4 h-4 flex-shrink-0" />
            <span>系统设置</span>
          </div>
        </router-link>
      </template>
    </nav>

    <!-- User Info -->
    <div class="px-3 py-4 border-t border-slate-800">
      <div class="flex items-center gap-3 px-2 py-2 rounded-lg hover:bg-slate-800 transition-colors group cursor-pointer" @click="showUserMenu = !showUserMenu">
        <div class="w-8 h-8 bg-indigo-600 rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
          {{ avatarInitial }}
        </div>
        <div class="flex-1 min-w-0">
          <div class="text-slate-200 text-sm font-medium truncate">{{ auth.user?.display_name || auth.user?.username }}</div>
          <div class="text-slate-500 text-xs">{{ roleLabel }}</div>
        </div>
        <ChevronUp v-if="showUserMenu" class="w-4 h-4 text-slate-500" />
        <ChevronDown v-else class="w-4 h-4 text-slate-500" />
      </div>

      <!-- User Menu Dropdown -->
      <Transition name="slide-up">
        <div v-if="showUserMenu" class="mt-2 bg-slate-800 rounded-lg overflow-hidden border border-slate-700">
          <button @click="goProfile" class="w-full flex items-center gap-2 px-4 py-2.5 text-slate-300 hover:bg-slate-700 text-sm transition-colors">
            <UserCircle class="w-4 h-4" />
            个人设置
          </button>
          <button @click="doLogout" class="w-full flex items-center gap-2 px-4 py-2.5 text-red-400 hover:bg-slate-700 text-sm transition-colors">
            <LogOut class="w-4 h-4" />
            退出登录
          </button>
        </div>
      </Transition>
    </div>
  </aside>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
import { LayoutDashboard, FolderKanban, Search, Users, Settings, LogOut, UserCircle, ChevronDown, ChevronUp } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()
const showUserMenu = ref(false)

const avatarInitial = computed(() => {
  const name = auth.user?.display_name || auth.user?.username || '?'
  return name.charAt(0).toUpperCase()
})

const roleLabel = computed(() => {
  const map = { admin: '管理员', developer: '开发者', executor: '执行者' }
  return map[auth.user?.role] || auth.user?.role
})

function goProfile() {
  showUserMenu.value = false
  router.push('/settings')
}

function doLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.slide-up-enter-active, .slide-up-leave-active {
  transition: all 0.15s ease;
}
.slide-up-enter-from, .slide-up-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>
