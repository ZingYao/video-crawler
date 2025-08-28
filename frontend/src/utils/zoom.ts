export function applyPageScale(scale: number): void {
  const clamped = Math.max(0.5, Math.min(1.5, Number(scale) || 1))
  const root = document.documentElement as HTMLElement

  const canUseZoom = 'zoom' in (root.style as any)

  if (canUseZoom) {
    // 使用 zoom，不再叠加 transform，避免布局计算异常
    ;(root.style as any).zoom = String(clamped)
    root.style.transform = ''
    // 仍设置 CSS 变量，供布局按比例计算高度（如 sidebar/app-layout）
    root.style.setProperty('--app-scale', String(clamped))
    root.style.transformOrigin = ''
    // 对于支持 zoom 的环境，保持标准高度设置，避免产生额外空白
    root.style.height = '100%'
    document.body.style.width = ''
    document.body.style.height = ''
    document.body.style.minHeight = '100vh'
    // 同步全局状态，供其他模块(如虚拟光标)读取并适配坐标/边界
    ;(window as any).__APP_SCALE = clamped
    root.setAttribute('data-app-scale', String(clamped))
    root.style.setProperty('--app-scale', String(clamped))
    window.dispatchEvent(new CustomEvent('app-scale-changed', { detail: { scale: clamped } }))
    return
  }

  // 兜底：通过 transform 缩放，并设置尺寸以避免底部留白
  root.style.setProperty('--app-scale', String(clamped))
  root.style.transformOrigin = 'top left'
  root.style.transform = `scale(var(--app-scale))`
  // 让布局区域按缩放后的视觉尺寸伸展，避免出现底部空白
  root.style.height = '100%'
  document.body.style.minHeight = '100vh'
  document.body.style.width = '100%'
  // 同步全局状态与事件
  ;(window as any).__APP_SCALE = clamped
  root.setAttribute('data-app-scale', String(clamped))
  window.dispatchEvent(new CustomEvent('app-scale-changed', { detail: { scale: clamped } }))
}

