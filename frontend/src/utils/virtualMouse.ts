// 虚拟鼠标：方向键控制移动，持续按压指数级加速，限制在屏幕边缘

type Vec2 = { x: number; y: number }

const clamp = (v: number, min: number, max: number) => Math.max(min, Math.min(max, v))

export interface VirtualMouseOptions {
  // 初始速度（像素/秒）
  baseSpeed?: number
  // 最大速度（像素/秒）
  maxSpeed?: number
  // 加速倍增周期（毫秒）：每隔该时间速度 *= accelerateFactor
  accelerateIntervalMs?: number
  // 每次倍增的倍数（>1）
  accelerateFactor?: number
  // 鼠标大小（像素）
  cursorSize?: number
}

export class VirtualMouse {
  private opts: Required<VirtualMouseOptions>
  private cursor: HTMLDivElement
  private pos: Vec2 = { x: 200, y: 200 }
  private vel: Vec2 = { x: 0, y: 0 }
  private held: Record<string, boolean> = {}
  private speed: number
  private accelTimer: number | null = null
  private rafId: number | null = null
  private enabled = false
  private lastPressTs: Record<string, number> = {}
  private pressCount: Record<string, number> = {}
  private lastKeyUpTs: Record<string, number> = {}
  private keyUpCount: Record<string, number> = {}
  private dblPressThreshold = 500 // 更严格的双击间隔，提高判定准确性
  private hiddenByInputFocus = false
  private hiddenByInactivity = false
  private inactivityTimer: number | null = null
  private inactivityMs = 10000

  constructor(options?: VirtualMouseOptions) {
    this.opts = {
      baseSpeed: options?.baseSpeed ?? 300, // 300 px/s 起步
      maxSpeed: options?.maxSpeed ?? 1600,  // 合理上限
      accelerateIntervalMs: options?.accelerateIntervalMs ?? 250,
      accelerateFactor: options?.accelerateFactor ?? 1.35,
      cursorSize: options?.cursorSize ?? 22,
    }
    this.speed = this.opts.baseSpeed
    this.cursor = this.createCursor()
  }

  private createCursor(): HTMLDivElement {
    const el = document.createElement('div')
    el.setAttribute('data-virtual-mouse', 'true')
    const size = this.opts.cursorSize
    el.style.cssText = [
      'position:fixed',
      `width:${size}px`,
      `height:${size}px`,
      'border-radius:50%\n',
      'pointer-events:none',
      'z-index:2147483647',
      'background:#10b981',
      'border:2px solid #ffffff',
      'box-shadow:0 0 8px rgba(16,185,129,0.6)',
      'transform:translate(-50%, -50%)',
      'opacity:0',
      'transition: background-color 120ms ease, opacity 250ms ease',
      'mix-blend-mode: normal',
    ].join(';')
    document.body.appendChild(el)
    return el
  }

  private setCursorPos(p: Vec2) {
    this.cursor.style.left = `${p.x}px`
    this.cursor.style.top = `${p.y}px`
  }

  private updateVelocity() {
    let x = 0, y = 0
    if (this.held['ArrowLeft']) x -= 1
    if (this.held['ArrowRight']) x += 1
    if (this.held['ArrowUp']) y -= 1
    if (this.held['ArrowDown']) y += 1
    // 归一化
    if (x !== 0 || y !== 0) {
      const len = Math.hypot(x, y)
      x /= len; y /= len
    }
    this.vel.x = x
    this.vel.y = y
  }

  private startAcceleration() {
    this.stopAcceleration()
    this.accelTimer = window.setInterval(() => {
      // 若没有方向键按下，速度回到基础值
      if (!(this.held['ArrowLeft'] || this.held['ArrowRight'] || this.held['ArrowUp'] || this.held['ArrowDown'])) {
        this.speed = this.opts.baseSpeed
        return
      }
      // 指数级加速，至上限
      this.speed = Math.min(this.opts.maxSpeed, this.speed * this.opts.accelerateFactor)
    }, this.opts.accelerateIntervalMs)
  }

  private stopAcceleration() {
    if (this.accelTimer !== null) { window.clearInterval(this.accelTimer); this.accelTimer = null }
  }

  private loop = () => {
    const dt = 1 / 60 // 以 60fps 近似，稳定动画（避免依赖高分辨率时间）
    const dx = this.vel.x * this.speed * dt
    const dy = this.vel.y * this.speed * dt

    const size = this.opts.cursorSize
    const maxX = window.innerWidth - 1
    const maxY = window.innerHeight - 1
    // 更新位置 + 边界限制（中心点不越界）
    this.pos.x = clamp(this.pos.x + dx, size / 2, maxX - size / 2)
    this.pos.y = clamp(this.pos.y + dy, size / 2, maxY - size / 2)
    this.setCursorPos(this.pos)

    this.rafId = window.requestAnimationFrame(this.loop)
  }

  private registerActivity() {
    // 若由输入框隐藏则不显示
    if (!this.hiddenByInputFocus) this.showCursor(250)
    // 重置 10s 无操作隐藏
    if (this.inactivityTimer) { window.clearTimeout(this.inactivityTimer); this.inactivityTimer = null }
    this.inactivityTimer = window.setTimeout(() => {
      // 仅当没有方向键按住时才隐藏
      if (!(this.held['ArrowLeft'] || this.held['ArrowRight'] || this.held['ArrowUp'] || this.held['ArrowDown'])) {
        this.hideCursor(1000, 'inactivity')
      }
    }, this.inactivityMs)

    // 首次使用事件（供外部弹窗提示）
    if (!(window as any).__vmFirstUseNotified) {
      (window as any).__vmFirstUseNotified = true
      try { window.dispatchEvent(new CustomEvent('virtual-mouse-first-use')) } catch {}
    }
  }

  private findScrollableContainer(start: Element | null, axis: 'y' | 'x'): Element | null {
    let el: Element | null = start
    const isScrollable = (node: Element) => {
      const style = window.getComputedStyle(node as Element)
      if (axis === 'y') {
        const overflowY = style.overflowY
        if (!/(auto|scroll)/.test(overflowY)) return false
        return (node as HTMLElement).scrollHeight > (node as HTMLElement).clientHeight
      } else {
        const overflowX = style.overflowX
        if (!/(auto|scroll)/.test(overflowX)) return false
        return (node as HTMLElement).scrollWidth > (node as HTMLElement).clientWidth
      }
    }
    while (el && el !== document.body && el !== document.documentElement) {
      if (isScrollable(el)) return el
      el = el.parentElement
    }
    // 回退：页面主体
    return (document.scrollingElement as Element) || document.documentElement
  }

  private scrollByDelta(delta: number) {
    const elAtPointer = document.elementFromPoint(this.pos.x, this.pos.y)
    const container = this.findScrollableContainer(elAtPointer, 'y')
    // 优先使用元素滚动，其次使用 window
    if (container && container !== document.documentElement && container !== document.body) {
      try {
        ;(container as any).scrollBy?.({ top: delta, behavior: 'smooth' })
      } catch {
        ;(container as HTMLElement).scrollTop += delta
      }
    } else {
      try { window.scrollBy({ top: delta, behavior: 'smooth' }) } catch { window.scrollBy(0, delta) }
    }
  }

  private scrollByDeltaX(delta: number) {
    const elAtPointer = document.elementFromPoint(this.pos.x, this.pos.y)
    const container = this.findScrollableContainer(elAtPointer, 'x')
    if (container && container !== document.documentElement && container !== document.body) {
      try {
        ;(container as any).scrollBy?.({ left: delta, behavior: 'smooth' })
      } catch {
        ;(container as HTMLElement).scrollLeft += delta
      }
    } else {
      try { window.scrollBy({ left: delta, behavior: 'smooth' }) } catch { (document.scrollingElement || document.documentElement).scrollLeft += delta }
    }
  }

  private onKeyDown = (e: KeyboardEvent) => {
    if (!this.enabled) return
    const active = document.activeElement as HTMLElement | null
    // 规范化方向键（兼容部分 AndroidTV/老设备）
    const normKey = (() => {
      const k = e.key
      if (k === 'ArrowLeft' || k === 'ArrowRight' || k === 'ArrowUp' || k === 'ArrowDown') return k
      // 兼容 DPAD
      if ((k as any) === 'Left') return 'ArrowLeft'
      if ((k as any) === 'Right') return 'ArrowRight'
      if ((k as any) === 'Up') return 'ArrowUp'
      if ((k as any) === 'Down') return 'ArrowDown'
      // 兼容 keyCode
      const code: any = (e as any).keyCode
      if (code === 37) return 'ArrowLeft'
      if (code === 38) return 'ArrowUp'
      if (code === 39) return 'ArrowRight'
      if (code === 40) return 'ArrowDown'
      return k
    })()
    const isEditable = !!active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA' || active.tagName === 'SELECT' || active.getAttribute('contenteditable') === 'true')

    // 调试日志：按键与状态
    try {
      console.log('[VM][keydown]', { key: e.key, normKey, repeat: e.repeat, ts: Date.now() })
    } catch {}

    // Enter：在非输入模式下触发鼠标点击
    if (normKey === 'Enter' && !isEditable) {
      e.preventDefault(); e.stopPropagation()
      const el = document.elementFromPoint(this.pos.x, this.pos.y) as HTMLElement | null
      if (el) {
        const ev: MouseEventInit = { bubbles: true, cancelable: true, clientX: this.pos.x, clientY: this.pos.y }
        el.dispatchEvent(new MouseEvent('mousedown', ev))
        el.dispatchEvent(new MouseEvent('mouseup', ev))
        el.dispatchEvent(new MouseEvent('click', ev))
      }
      this.registerActivity()
      return
    }

    // 输入聚焦时：方向键不接管，仅处理 Esc/Back 退出并显示光标
    if (isEditable) {
      if (normKey === 'Escape' || normKey === 'Backspace' || normKey === 'BrowserBack') {
        e.preventDefault(); e.stopPropagation()
        try { active?.blur() } catch {}
        this.showCursor(250)
      }
      return
    }
    if (normKey === 'ArrowLeft' || normKey === 'ArrowRight' || normKey === 'ArrowUp' || normKey === 'ArrowDown') {
      // 阻止默认与冒泡，避免焦点移动/页面滚动
      e.preventDefault(); e.stopPropagation()
      // keydown 仅用于移动，双击滚动改由 keyup 识别，更稳定
      this.held[normKey] = true
      this.updateVelocity()
      try { console.log('[VM] move update', { held: { ...this.held }, vel: { ...this.vel }, speed: this.speed }) } catch {}
      this.registerActivity()
    }
    // AndroidTV 返回键兼容（无输入态时，模拟 Esc 行为）
    if (normKey === 'Backspace' || normKey === 'BrowserBack') {
      e.preventDefault(); e.stopPropagation()
      const activeEl = document.activeElement as HTMLElement | null
      if (activeEl && (activeEl.tagName === 'INPUT' || activeEl.tagName === 'TEXTAREA' || activeEl.tagName === 'SELECT' || activeEl.getAttribute('contenteditable') === 'true')) {
        try { activeEl.blur() } catch {}
      } else {
        // 非输入态：尽量不影响页面路由，交由业务自行处理返回
        // 这里仅发出一个事件供业务层监听
        try { window.dispatchEvent(new CustomEvent('virtual-mouse-back')) } catch {}
      }
      this.showCursor(250)
      return
    }
  }

  private onKeyUp = (e: KeyboardEvent) => {
    if (!this.enabled) return
    // 规范化
    const normKey = (() => {
      const k = e.key
      if (k === 'ArrowLeft' || k === 'ArrowRight' || k === 'ArrowUp' || k === 'ArrowDown') return k
      if ((k as any) === 'Left') return 'ArrowLeft'
      if ((k as any) === 'Right') return 'ArrowRight'
      if ((k as any) === 'Up') return 'ArrowUp'
      if ((k as any) === 'Down') return 'ArrowDown'
      const code: any = (e as any).keyCode
      if (code === 37) return 'ArrowLeft'
      if (code === 38) return 'ArrowUp'
      if (code === 39) return 'ArrowRight'
      if (code === 40) return 'ArrowDown'
      return k
    })()
    if (normKey === 'ArrowLeft' || normKey === 'ArrowRight' || normKey === 'ArrowUp' || normKey === 'ArrowDown') {
      e.preventDefault(); e.stopPropagation()
      delete this.held[normKey]
      this.updateVelocity()
      if (!(this.held['ArrowLeft'] || this.held['ArrowRight'] || this.held['ArrowUp'] || this.held['ArrowDown'])) {
        this.speed = this.opts.baseSpeed
      }
      // 在 keyup 上做双击滚动识别（repeat 不干扰）
      const now = Date.now()
      const lastUp = this.lastKeyUpTs[normKey] || 0
      const interval = now - lastUp
      this.lastKeyUpTs[normKey] = now
      if (interval <= this.dblPressThreshold) {
        this.keyUpCount[normKey] = (this.keyUpCount[normKey] || 1) + 1
      } else {
        this.keyUpCount[normKey] = 1
      }
      try { console.log('[VM][keyup]', { normKey, interval, count: this.keyUpCount[normKey] }) } catch {}
      if (this.keyUpCount[normKey] <= 2 && interval <= this.dblPressThreshold) {
        if (normKey === 'ArrowUp' || normKey === 'ArrowDown') {
          const delta = Math.round(window.innerHeight * 0.6) * (normKey === 'ArrowDown' ? 1 : -1)
          try { console.log('[VM] vertical double via keyup', { normKey, delta, interval }) } catch {}
          this.scrollByDelta(delta)
        } else if (normKey === 'ArrowLeft' || normKey === 'ArrowRight') {
          const deltaX = Math.round(window.innerWidth * 0.6) * (normKey === 'ArrowRight' ? 1 : -1)
          try { console.log('[VM] horizontal double via keyup', { normKey, deltaX, interval }) } catch {}
          this.scrollByDeltaX(deltaX)
        }
        // 重置，避免第三击误抑制
        this.lastKeyUpTs[normKey] = 0
        this.keyUpCount[normKey] = 0
      }
      this.registerActivity()
    }
  }

  enable() {
    if (this.enabled) return
    this.enabled = true
    // 初始放到视口中心
    this.pos = { x: window.innerWidth / 2, y: window.innerHeight / 2 }
    this.setCursorPos(this.pos)
    // capture: true 进一步拦截，优先于页面其它监听器
    window.addEventListener('keydown', this.onKeyDown, { passive: false, capture: true })
    window.addEventListener('keyup', this.onKeyUp, { passive: false, capture: true })
    // 避免 AndroidTV 焦点波动导致 held 被清空，仅重置速度
    window.addEventListener('blur', () => { this.speed = this.opts.baseSpeed; this.updateVelocity() })
    window.addEventListener('focusin', this.onFocusIn as any)
    window.addEventListener('focusout', this.onFocusOut as any)
    this.startAcceleration()
    this.rafId = window.requestAnimationFrame(this.loop)
  }

  disable() {
    if (!this.enabled) return
    this.enabled = false
    this.stopAcceleration()
    if (this.rafId !== null) { cancelAnimationFrame(this.rafId); this.rafId = null }
    window.removeEventListener('keydown', this.onKeyDown)
    window.removeEventListener('keyup', this.onKeyUp)
    window.removeEventListener('focusin', this.onFocusIn as any)
    window.removeEventListener('focusout', this.onFocusOut as any)
    try { this.cursor.remove() } catch {}
  }

  private hideCursor(durationMs: number, reason: 'input' | 'inactivity') {
    if (reason === 'input') this.hiddenByInputFocus = true
    if (reason === 'inactivity') this.hiddenByInactivity = true
    this.cursor.style.transition = `background-color 120ms ease, opacity ${durationMs}ms ease`
    // 强制 reflow，确保过渡时间更改被浏览器应用
    void this.cursor.offsetWidth
    this.cursor.style.opacity = '0'
  }
  private showCursor(durationMs: number) {
    this.hiddenByInactivity = false
    if (this.hiddenByInputFocus) return // 输入态仍保持隐藏
    this.cursor.style.transition = `background-color 120ms ease, opacity ${durationMs}ms ease`
    void this.cursor.offsetWidth
    this.cursor.style.opacity = '1'
  }
  private onFocusIn = (e: FocusEvent) => {
    const t = e.target as HTMLElement | null
    if (!t) return
    const isEditable = t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.getAttribute('contenteditable') === 'true'
    if (isEditable) this.hideCursor(250, 'input')
  }
  private onFocusOut = (_e: FocusEvent) => { this.hiddenByInputFocus = false; this.showCursor(250) }
}

// 工具：便捷启用（按需加载）
export const initVirtualMouse = (opts?: VirtualMouseOptions) => {
  const vm = new VirtualMouse(opts)
  vm.enable()
  ;(window as any).__virtualMouse = vm
  return vm
}


