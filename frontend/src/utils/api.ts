import { useAuthStore } from '@/stores/auth'

// 检测是否在Wails环境中
export const isWailsEnvironment = () => {
  return !!(window.go && window.go.main && window.go.main.App && typeof window.go.main.App.GetConfig === 'function')
}

// 获取Wails HTTP服务端口
let wailsServerPort: number | null = null

export const getWailsServerPort = async (): Promise<number> => {
  if (isWailsEnvironment()) {
    // 每次都重新获取端口，避免缓存过期
    wailsServerPort = await window.go.main.App.GetServerPort()
    return wailsServerPort
  }
  
  throw new Error('不在Wails环境中')
}

// 获取API基础URL
export const getApiBaseUrl = async (): Promise<string> => {
  if (isWailsEnvironment()) {
    const port = await getWailsServerPort()
    return `http://localhost:${port}`
  } else {
    return '' // 相对路径，使用当前域名
  }
}

// 通用请求函数，根据是否有token决定是否添加Authorization头
export const makeRequest = async (url: string, options: RequestInit = {}) => {
  const authStore = useAuthStore()
  const token = authStore.token
  
  // 如果有token，添加Authorization头
  if (token) {
    options.headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    }
  }
  
  const baseUrl = await getApiBaseUrl()
  const fullUrl = `${baseUrl}${url}`
  
  const response = await fetch(fullUrl, options)
  return response
}

// 配置相关API
export const configAPI = {
  getConfig: async () => {
    if (isWailsEnvironment()) {
      return await window.go.main.App.GetConfig()
    } else {
      const response = await fetch('/api/config')
      const result = await response.json()
      if (result.code === 0) {
        return result.data
      } else {
        throw new Error(result.message || '获取配置失败')
      }
    }
  },
  
  userLogin: async (username: string, password: string) => {
    const response = await makeRequest('/api/user/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    })
    const result = await response.json()
    return result
  }
}

// 视频源相关API
export const videoSourceAPI = {
  getList: async () => {
    const response = await makeRequest('/api/video-source/list')
    const result = await response.json()
    return result
  },
  
  getVideoSourceDetail: async (id: string) => {
    const response = await makeRequest(`/api/video-source/detail?id=${encodeURIComponent(id)}`)
    const result = await response.json()
    return result
  },
  
  saveVideoSource: async (data: any) => {
    const response = await makeRequest('/api/video-source/save', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    const result = await response.json()
    return result
  },
  
  exportVideoSources: async () => {
    const response = await makeRequest('/api/video-source/export')
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }
    return response
  },
  
  importVideoSources: async (data: any[]) => {
    const response = await makeRequest('/api/video-source/import', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    const result = await response.json()
    return result
  },
  
  delete: async (id: string) => {
    const response = await makeRequest('/api/video-source/delete', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ id }),
    })
    const result = await response.json()
    return result
  },
  
  setStatus: async (id: string, status: number) => {
    const response = await makeRequest('/api/video-source/set-status', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ id, status }),
    })
    const result = await response.json()
    return result
  },
  
  checkStatus: async (id: string) => {
    const response = await makeRequest(`/api/video-source/check-status?id=${encodeURIComponent(id)}`)
    const result = await response.json()
    return result
  }
}

// 脚本测试相关API
export const scriptAPI = {
  testLua: async (script: string, method: string, params: any) => {
    const response = await makeRequest('/api/lua/test', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ script, method, params }),
    })
    const result = await response.json()
    return result
  },
  
  testJS: async (script: string, method: string, params: any) => {
    const response = await makeRequest('/api/js/test', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ script, method, params }),
    })
    const result = await response.json()
    return result
  }
}
