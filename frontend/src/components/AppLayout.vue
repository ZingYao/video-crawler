<template>
  <div class="app-layout">
    <!-- 左侧菜单 -->
    <aside class="sidebar" :class="sidebarVisible ? '' : 'collapsed'" :data-visible="sidebarVisible">
      <div class="sidebar-header">
        <div class="logo">
          <span class="logo-icon">🎬</span>
          <span class="logo-text" v-show="sidebarVisible">视频爬虫</span>
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
          <li v-for="item in filteredMenuItems" :key="item.id" class="nav-item">
            <button
              @click="handleMenuClick(item)"
              class="nav-link"
              :class="{ active: activeMenu === item.id }"
            >
              <span class="nav-icon">{{ item.icon }}</span>
              <span class="nav-text" v-show="sidebarVisible">{{ item.label }}</span>
            </button>
          </li>
        </ul>
      </nav>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-content" :class="sidebarVisible ? '' : 'sidebar-collapsed'" :data-visible="sidebarVisible" @click="handleMainClick">
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useConfigStore } from '@/stores/config'
import { isAppEnvironment } from '@/utils/api'
import { DownOutlined, UserOutlined, LogoutOutlined } from '@ant-design/icons-vue'

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
const sidebarVisible = ref(false) // 默认收起，让用户手动控制
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
  {
    id: 'home',
    icon: '🏠',
    label: '首页',
    description: '系统概览和快速操作',
    requiresAdmin: false,
    route: '/'
  },
  {
    id: 'movie',
    icon: '🎭',
    label: '观影',
    description: '搜索和观看视频',
    requiresAdmin: false,
    route: '/movie'
  },
  {
    id: 'watch-history',
    icon: '📺',
    label: '观看历史',
    description: '查看您的视频观看历史',
    requiresAdmin: false,
    route: '/history/watch'
  },
  {
    id: 'user-management',
    icon: '👥',
    label: '用户管理',
    description: '管理系统用户账户',
    requiresAdmin: true,
    route: '/user-management'
  },
  {
    id: 'video-source-management',
    icon: '🎬',
    label: '视频资源管理',
    description: '管理视频资源站点',
    requiresAdmin: false,
    requiresSiteAdmin: true,
    route: '/video-source-management'
  }
  ,{
    id: 'api-docs',
    icon: '🧭',
    label: '接口文档',
    description: 'API 文档与示例',
    requiresAdmin: false,
    route: '/api-docs'
  }
]

// 计算属性
const filteredMenuItems = computed(() => {
  return menuItems.filter(item => {
    // 接口文档仅在应用环境（Wails/Android WebView）展示
    if (item.id === 'api-docs' && !isAppEnvironment()) {
      return false
    }
    // 如果不需要登录，显示所有菜单（除了用户管理）
    if (!configStore.needsLogin()) {
      // 在无需登录模式下，只隐藏用户管理菜单，其他都显示
      if (item.id === 'user-management') {
        return false
      }
      // 其他菜单都显示，包括观看历史
      return true
    }
    
    // 需要登录模式下的权限检查
    if (item.requiresAdmin) {
      // 仅管理员
      return authStore.user?.isAdmin === true
    }
    if (item.requiresSiteAdmin) {
      // 管理员或资源站点管理员
      return authStore.user?.isAdmin === true || authStore.user?.isSiteAdmin === true
    }
    return true
  })
})

const showCloseButton = computed(() => {
  return sidebarVisible.value && window.innerWidth <= 1024
})

// 方法
const toggleSidebar = () => {
  console.log('toggleSidebar called, current sidebarVisible:', sidebarVisible.value)
  sidebarVisible.value = !sidebarVisible.value
  console.log('sidebarVisible after toggle:', sidebarVisible.value)
  console.log('DOM should update now...')
}

const handleMainClick = () => {
  // 暂时禁用移动端自动收起功能，让用户手动控制
  // if (window.innerWidth <= 1024 && sidebarVisible.value) {
  //   console.log('handleMainClick: closing menu due to mobile click')
  //   sidebarVisible.value = false
  // }
}

const handleMenuClick = (item: MenuItem) => {

  
  activeMenu.value = item.id
  router.push(item.route)
}

const goToProfile = () => {
  router.push('/profile')
}

const goToWatchHistory = () => {
  // 跳到观看历史，默认查看自己
  router.push('/history/watch')
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const updateActiveMenu = () => {
  const currentPath = route.path
  const menuItem = menuItems.find(item => item.route === currentPath)
  if (menuItem) {
    activeMenu.value = menuItem.id
  }
}

// 生命周期
onMounted(() => {
  // 移动端默认收起菜单，但允许用户手动展开
  updateActiveMenu()
})
</script>

<style scoped>
@import './AppLayout.css';
</style>
