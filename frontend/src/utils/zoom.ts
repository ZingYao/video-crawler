export function applyPageScale(scale: number): void {
  const clamped = Math.max(0.5, Math.min(1.5, Number(scale) || 1))
  const root = document.documentElement as HTMLElement

  // 优先使用 CSS zoom（桌面浏览器、WebView 通常支持）
  ;(root.style as any).zoom = String(clamped)

  // 兜底：通过 transform 缩放，并设置 transform-origin
  root.style.setProperty('--app-scale', String(clamped))
  root.style.transformOrigin = 'top left'
  root.style.transform = `scale(var(--app-scale))`

  // 调整 body 宽度，避免缩放后出现横向滚动
  const inv = 1 / clamped
  document.body.style.width = `${inv * 100}%`
}

