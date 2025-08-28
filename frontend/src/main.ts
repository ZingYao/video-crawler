import './assets/main.css'
import './styles/index.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import { ConfigProvider } from 'ant-design-vue'

import App from './App.vue'
import router from './router'
import { isWails } from './utils/wails'
import { useSettingsStore } from '@/stores/settings'

const app = createApp(App)

// 配置Ant Design主题
const theme = {
  token: {
    colorPrimary: '#10b981',
    colorPrimaryHover: '#34d399',
    colorPrimaryActive: '#059669',
    colorSuccess: '#10b981',
    colorInfo: '#10b981',
    colorWarning: '#f59e0b',
    colorError: '#ef4444',
  },
}

app.use(createPinia())
app.use(router)
app.use(Antd)

// 全局配置主题
app.use(ConfigProvider, theme)

// 检查是否在Wails环境中，如果是则重定向到启动页面
const checkWailsEnvironment = () => {
  const isWailsEnv = isWails()

  // 调试信息（生产环境构建时会自动移除）
  console.log('=== main.ts 环境检测 ===')
  console.log('isWails:', isWailsEnv)
  console.log('window.location.pathname:', window.location.pathname)

  if (isWailsEnv && window.location.pathname === '/') {
    console.log('检测到Wails环境，重定向到启动页面')
    // 延迟重定向，确保路由已经初始化
    setTimeout(() => {
      router.push('/startup')
    }, 100)
  } else {
    console.log('非Wails环境或不在根路径，不重定向')
  }
}

// 只在页面初始加载时检查一次，避免重复检查导致循环跳转
let hasCheckedWailsEnvironment = false

const performWailsEnvironmentCheck = () => {
  if (hasCheckedWailsEnvironment) {
    return
  }
  
  const isWailsEnv = isWails()
  if (isWailsEnv && window.location.pathname === '/') {
    hasCheckedWailsEnvironment = true
    console.log('检测到Wails环境，重定向到启动页面')
    setTimeout(() => {
      router.push('/startup')
    }, 100)
  }
}

// 立即检查一次
performWailsEnvironmentCheck()

// 延迟检查，以防 window.go 在页面加载后才被注入
setTimeout(() => {
  console.log('=== 延迟环境检测 ===')
  performWailsEnvironmentCheck()
}, 1000)

// 再次延迟检查
setTimeout(() => {
  console.log('=== 再次延迟环境检测 ===')
  performWailsEnvironmentCheck()
}, 3000)

app.mount('#app')

// 按设置开关加载虚拟鼠标，并在首次使用时弹出说明
console.log('[MAIN] 开始检查虚拟光标启动条件...')
import('./utils/virtualMouse').then(async ({ initVirtualMouse }) => {
  try {
    console.log('[MAIN] 虚拟鼠标模块加载成功')
    const settingsStore = useSettingsStore()
    console.log('[MAIN] 开始加载设置...')
    await settingsStore.loadSettings()
    
    const settings = settingsStore.settings
    console.log('[MAIN] 设置加载完成，检查虚拟光标配置:', {
      virtualCursorEnabled: settings.virtualCursorEnabled,
      virtualCursorTipsShown: settings.virtualCursorTipsShown,
      allSettings: settings,
      timestamp: Date.now()
    })
    
    if (settings.virtualCursorEnabled) {
      console.log('[MAIN] ✅ 虚拟光标已启用，开始初始化...')
      const initOptions = { baseSpeed: 120, maxSpeed: 600, accelerateIntervalMs: 240, accelerateFactor: 1.35, cursorSize: 22 }
      console.log('[MAIN] 虚拟光标初始化参数:', initOptions)
      
      initVirtualMouse(initOptions)
      console.log('[MAIN] 虚拟光标初始化完成')
      
      // 首次使用提示
      const onFirstUse = async () => {
        console.log('[MAIN] 虚拟光标首次使用事件触发')
        if (!settingsStore.settings.virtualCursorTipsShown) {
          console.log('[MAIN] 显示虚拟光标使用说明弹窗')
          const { Modal } = await import('ant-design-vue')
          const { h } = await import('vue')
          Modal.info({
            title: '虚拟光标使用说明',
            width: 520,
            content: h('div', { style: 'line-height:1.75;color:#374151' }, [
              h('p', { style: 'margin:4px 0 10px;color:#6b7280' }, '用遥控器/键盘即可操控页面元素：'),
              h('ul', { style: 'padding-left:18px;margin:0' }, [
                h('li', '上下左右：移动光标（指数加速，越界限制）'),
                h('li', '回车 / Enter / OK：在光标位置点击'),
                h('li', '双击 上/下：页面上下翻动'),
                h('li', '双击 左/右：页面左右翻动（优先滚动横向容器）'),
                h('li', 'Esc / 返回：退出输入框焦点并恢复光标'),
                h('li', '输入框聚焦：光标隐藏，方向键仅移动文本光标'),
                h('li', '无操作10秒自动隐藏（1秒淡出），操作后0.25秒淡入'),
              ]),
            ]),
            okText: '已了解',
            maskClosable: true,
            centered: true,
          })
          await settingsStore.markVirtualCursorTipsShown()
          console.log('[MAIN] 虚拟光标使用说明已标记为已显示')
        } else {
          console.log('[MAIN] 虚拟光标使用说明已显示过，跳过弹窗')
        }
        window.removeEventListener('virtual-mouse-first-use', onFirstUse as any)
      }
      window.addEventListener('virtual-mouse-first-use', onFirstUse as any, { once: true })
      console.log('[MAIN] 虚拟光标首次使用事件监听器已注册')
    } else {
      console.log('[MAIN] ❌ 虚拟光标未启用，跳过初始化')
    }
  } catch (error) {
    console.log('[MAIN] 虚拟光标初始化失败:', error)
  }
}).catch((error) => {
  console.log('[MAIN] 虚拟鼠标模块加载失败:', error)
})
