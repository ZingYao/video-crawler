import './assets/main.css'
import './styles/index.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import { ConfigProvider } from 'ant-design-vue'

import App from './App.vue'
import router from './router'

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

app.mount('#app')
