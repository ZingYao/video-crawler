<template>
  <div class="app-layout">
    <!-- 左侧菜单 -->
    <aside :class="getSidebarClass()" :data-visible="isSidebarOpen" ref="sidebarRef" data-focus-scope>
      <div class="sidebar-header">
        <div class="logo">
          <span class="logo-icon">🎬</span>
          <span class="logo-text" v-show="sidebarVisible">视频聚合</span>
        </div>
        <!-- 移动端收起菜单按钮 -->
        <button 
          v-show="showCloseButton" 
          @click="toggleSidebar" 
          class="close-menu-btn"
          title="收起菜单"
        >
          <span>✕</span>
        </button>
      </div>

      <nav class="sidebar-nav">
        <ul class="nav-list">
          <li v-for="(item, idx) in filteredMenuItems" :key="item.id" class="nav-item">
            <button
              @click="handleMenuClick(item)"
              class="nav-link"
              :class="{ active: activeMenu === item.id }"
              :data-focus-order="idx + 1"
              tabindex="0"
            >
              <span class="nav-icon">{{ item.icon }}</span>
              <span class="nav-text" v-show="sidebarVisible">{{ item.label }}</span>
            </button>
          </li>
        </ul>
      </nav>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-content" :class="isSidebarOpen ? '' : 'sidebar-collapsed'" :data-visible="isSidebarOpen" @click="handleMainClick">
      <!-- 顶部导航栏 -->
      <header class="top-header">
        <div class="header-left">
          <button @click="toggleSidebar" class="menu-toggle">
            <span>☰</span>
          </button>
          <div class="header-title">
            <slot name="header-title">
              <h1>{{ pageTitle }}</h1>
            </slot>
          </div>
        </div>
        
        <div class="header-right">
          <div class="header-actions">
            <slot name="header-actions"></slot>
          </div>
          
          <!-- 用户信息区域 -->
          <div v-if="configStore.needsLogin() && authStore.isAuthenticated" class="user-info-section">
            <a-dropdown :trigger="['click']" placement="bottomRight">
              <div class="user-info-card">
                <div class="user-avatar">
                  <span>{{ authStore.user?.nickname?.charAt(0) || authStore.user?.username?.charAt(0) || 'U' }}</span>
                </div>
                <div class="user-details">
                  <div class="user-name">{{ authStore.user?.nickname || authStore.user?.username || '用户' }}</div>
                  <div class="user-role">{{ authStore.user?.isAdmin ? '管理员' : (authStore.user?.isSiteAdmin ? '资源站点管理员' : '普通用户') }}</div>
                </div>
                <DownOutlined class="dropdown-arrow" />
              </div>
              
              <template #overlay>
                <a-menu class="user-dropdown-menu">
                  <a-menu-item key="profile" @click="goToProfile">
                    <UserOutlined />
                    <span>个人中心</span>
                  </a-menu-item>
                  <a-menu-item key="watch-history" @click="goToWatchHistory">
                    <span>📺</span>
                    <span>观看历史</span>
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item key="logout" @click="handleLogout">
                    <LogoutOutlined />
                    <span>退出登录</span>
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </div>
      </header>

      <!-- 页面内容 -->
      <div class="page-content">
        <slot></slot>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useConfigStore } from '@/stores/config'
import { DownOutlined, UserOutlined, LogoutOutlined } from '@ant-design/icons-vue'
import { isClientEnvironment } from '@/utils/wails'

const getSidebarClass = () => {
  let sClass = 'sidebar'
  if (autoCollapsed.value && !sidebarVisible.value) {
    sClass += ' collapsed'
  }
  console.log('sidebar debug',sClass,autoCollapsed.value,sidebarVisible.value,autoCollapsed.value && !sidebarVisible.value,autoCollapsed.value || sidebarVisible.value,isSidebarOpen.value)
  return sClass
}

// 检测是否在应用环境中
const isAppEnvironment = () => {
  // 简单的检测逻辑，可以根据需要扩展
  return !!((window as any).go || /Android|iPhone|iPad/.test(navigator.userAgent))
}

// Props
interface Props {
  pageTitle?: string
}

const props = withDefaults(defineProps<Props>(), {
  pageTitle: '首页'
})

// 响应式数据
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const configStore = useConfigStore()
const sidebarVisible = ref(false) // 用户手动开关的显式状态
// 展开状态：大屏始终展开；小屏根据用户开关
const isSidebarOpen = computed(() => autoCollapsed.value || sidebarVisible.value)
const autoCollapsed = ref(false) // 按缩放与窗口宽度自动折叠
const DESIGN_BREAKPOINT = 1024 // 设计断点（未缩放）

const getScale = () => {
  const s = (window as any).__APP_SCALE
  const ds = Number(document.documentElement.getAttribute('data-app-scale') || '')
  return Number(s || ds || 1) || 1
}

const recomputeAutoCollapsed = () => {
  const scale = getScale()
  const effectiveWidth = window.innerWidth / scale
  autoCollapsed.value = effectiveWidth > DESIGN_BREAKPOINT
  console.log('autoCollapsed', autoCollapsed.value)
}
const sidebarRef = ref<HTMLElement | null>(null)
let outsideClickHandler: ((e: MouseEvent) => void) | null = null
let outsideClickTimer: number | null = null
const activeMenu = ref('')

// 菜单项类型定义
interface MenuItem {
  id: string
  icon: string
  label: string
  description: string
  requiresAdmin: boolean
  route: string
  requiresSiteAdmin?: boolean
  requiresWails?: boolean
}

// 菜单项配置
const menuItems: MenuItem[] = [
  { id: 'home', icon: '🏠', label: '首页', description: '系统概览和快速操作', requiresAdmin: false, route: '/' },
  { id: 'movie', icon: '🎭', label: '观影', description: '搜索和观看视频', requiresAdmin: false, route: '/movie' },
  { id: 'watch-history', icon: '📺', label: '观看历史', description: '查看您的视频观看历史', requiresAdmin: false, route: '/history/watch' },
  { id: 'user-management', icon: '👥', label: '用户管理', description: '管理系统用户账户', requiresAdmin: true, route: '/user-management' },
  { id: 'video-source-management', icon: '🎬', label: '视频资源管理', description: '管理视频资源站点', requiresAdmin: false, requiresSiteAdmin: true, route: '/video-source-management' },
  { id: 'api-docs', icon: '🧭', label: '接口文档', description: 'API 文档与示例', requiresAdmin: false, requiresWails: true, route: '/api-docs' },
  { id: 'settings', icon: '⚙️', label: '设置', description: '播放控制与搜索网站设置', requiresAdmin: false, route: '/settings' }
]

// 计算属性
const filteredMenuItems = computed(() => {
  return menuItems.filter(item => {
    // 如果系统配置为不需要登录，只隐藏用户管理相关菜单，允许访问站点管理
    if (!configStore.needsLogin()) {
      if (item.id === 'user-management') {
        return false
      }
      // 在无需登录模式下，允许访问站点管理页面
      if (item.id === 'video-source-management') {
        return true
      }
    }
    
    // 检查管理员权限
    if (item.requiresAdmin && !authStore.user?.isAdmin) {
      return false
    }
    
    // 检查站点管理员权限（管理员也拥有站点管理员权限）
    if (item.requiresSiteAdmin && !authStore.user?.isSiteAdmin && !authStore.user?.isAdmin) {
      return false
    }
    
    // 检查客户端环境要求（Wails 或 Android）
    if (item.requiresWails && !isClientEnvironment()) {
      return false
    }
    
    return true
  })
})

// 检查当前路由是否匹配菜单项
const isActiveMenu = (menuId: string) => {
  return route.path === menuItems.find(item => item.id === menuId)?.route
}

const showCloseButton = computed(() => isSidebarOpen.value && (window.innerWidth / (getScale() || 1)) <= DESIGN_BREAKPOINT)

// 方法
const removeOutsideListener = () => {
  if (outsideClickHandler) {
    document.removeEventListener('click', outsideClickHandler, true)
    outsideClickHandler = null
  }
  if (outsideClickTimer) {
    window.clearTimeout(outsideClickTimer)
    outsideClickTimer = null
  }
}

const scheduleOutsideListener = () => {
  removeOutsideListener()
  // 延迟100ms后开始监听一次性外部点击
  outsideClickTimer = window.setTimeout(() => {
    outsideClickHandler = (e: MouseEvent) => {
      const target = e.target as Node | null
      const inSidebar = !!(sidebarRef.value && target && sidebarRef.value.contains(target))
      if (!inSidebar) {
        console.log('outside click',sidebarVisible.value)
        sidebarVisible.value = false
        removeOutsideListener()
      }
    }
    // 使用捕获阶段，优先于子元素处理，且只处理一次
    document.addEventListener('click', outsideClickHandler, true)
  }, 100)
}

const toggleSidebar = () => {
  const next = !sidebarVisible.value
  sidebarVisible.value = next
  if (next) {
    scheduleOutsideListener()
  } else {
    removeOutsideListener()
  }
}

const handleMainClick = () => {}
const handleMenuClick = (item: MenuItem) => { activeMenu.value = item.id; router.push(item.route) }
const goToProfile = () => router.push('/profile')
const goToWatchHistory = () => router.push('/history/watch')
const handleLogout = () => { authStore.logout(); router.push('/login') }

const updateActiveMenu = () => {
  const currentPath = route.path
  const menuItem = menuItems.find(item => item.route === currentPath)
  if (menuItem) activeMenu.value = menuItem.id
}

// 生命周期
onMounted(() => {
  updateActiveMenu()
  recomputeAutoCollapsed()
  window.addEventListener('resize', recomputeAutoCollapsed)
  window.addEventListener('app-scale-changed', recomputeAutoCollapsed as any)
})


const getFocusable = (): HTMLElement[] => {
  const selector = 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
  const list = Array.from(document.querySelectorAll(selector)) as HTMLElement[]
  return list.filter(el => !el.hasAttribute('disabled') && el.tabIndex !== -1 && el.offsetParent !== null)
}

const focusByOffset = (offset: number) => {
  const focusables = getFocusable()
  if (focusables.length === 0) return
  const active = document.activeElement as HTMLElement | null
  let idx = focusables.findIndex(el => el === active)
  if (idx === -1) idx = 0
  let next = (idx + offset) % focusables.length
  if (next < 0) next = focusables.length - 1
  focusables[next].focus()
}

</script>

<style scoped>
@import './AppLayout.css';
</style>
