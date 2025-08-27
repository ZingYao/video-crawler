/**
 * 统一的 Wails 环境检测方法
 * 通过检查 window.go 对象和 domain 协议来判断是否为 Wails 环境
 * @returns {boolean} 是否为 Wails 环境
 */
export const isWails = (): boolean => {
  // 方法1: 检查 window.go 对象是否存在
  const hasGoObject = !!(window as any).go
  
  // 方法2: 检查当前域名协议是否为 wails://
  const hasWailsProtocol = window.location.protocol === 'wails:'
  
  // 方法3: 检查当前域名是否包含 wails.localhost
  const hasWailsDomain = window.location.origin.includes('wails.localhost')
  
  // 方法4: 检查当前域名是否包含 wails://
  const hasWailsOrigin = window.location.origin.includes('wails://')
  
  // 任一条件满足即为 Wails 环境
  const result = hasGoObject || hasWailsProtocol || hasWailsDomain || hasWailsOrigin
  
  // 调试信息（生产环境构建时会自动移除）
  console.error('=== isWails 环境检测 ===')
  console.error('hasGoObject:', hasGoObject)
  console.error('hasWailsProtocol:', hasWailsProtocol)
  console.error('hasWailsDomain:', hasWailsDomain)
  console.error('hasWailsOrigin:', hasWailsOrigin)
  console.error('window.location.protocol:', window.location.protocol)
  console.error('window.location.origin:', window.location.origin)
  console.error('window.go:', (window as any).go)
  console.error('最终结果:', result)
  
  return result
}

/**
 * 获取 Wails 环境详细信息
 * @returns {object} Wails 环境详细信息
 */
export const getWailsInfo = () => {
  return {
    isWails: isWails(),
    hasGoObject: !!(window as any).go,
    hasWailsProtocol: window.location.protocol === 'wails:',
    hasWailsDomain: window.location.origin.includes('wails.localhost'),
    hasWailsOrigin: window.location.origin.includes('wails://'),
    protocol: window.location.protocol,
    origin: window.location.origin,
    href: window.location.href,
    goObject: (window as any).go
  }
}

/**
 * 检测是否为客户端环境（Wails 或 Android）
 * @returns {boolean} 是否为客户端环境
 */
export const isClientEnvironment = (): boolean => {
  // 检测 Wails 环境
  const isWailsEnv = isWails()
  
  // 检测 Android 环境
  const isAndroidEnv = !!(window as any).AndroidKV || 
                      window.location.hostname.includes('android') ||
                      /Android/.test(navigator.userAgent)
  
  const result = isWailsEnv || isAndroidEnv
  
  // 调试信息
  console.error('=== isClientEnvironment 环境检测 ===')
  console.error('isWailsEnv:', isWailsEnv)
  console.error('isAndroidEnv:', isAndroidEnv)
  console.error('hasAndroidKV:', !!(window as any).AndroidKV)
  console.error('hostname:', window.location.hostname)
  console.error('userAgent:', navigator.userAgent)
  console.error('最终结果:', result)
  
  return result
}
