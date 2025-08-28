// 虚拟鼠标：方向键控制移动，持续按压指数级加速，限制在屏幕边缘

type Vec2 = { x: number; y: number }

const clamp = (v: number, min: number, max: number) => Math.max(min, Math.min(max, v))
// 静默日志：将本文件中的所有 [VM] 调试输出置空
// 统一使用 vmLog，便于后续按需开启
const vmLog = (..._args: any[]) => {}

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

// 全局虚拟光标实例管理
let globalVirtualMouse: VirtualMouse | null = null

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
  private hiddenByMouse = false // 新增：被鼠标隐藏
  private mouseActivityTimer: number | null = null
  private mouseInactivityMs = 3000 // 鼠标无活动3秒后重新显示虚拟光标
  private lastMousePos = { x: 0, y: 0 } // 记录鼠标最后位置
  private isKeyboardMode = false // 是否处于键盘控制模式
  private lastHoveredElement: Element | null = null // 记录最后hover的元素
  private hiddenByTouch = false
  private touchHideTimer: number | null = null

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
    // 注入光标呼吸动画样式
    this.injectCursorAnimationStyle()
    el.style.cssText = [
      'position:fixed',
      `width:${size}px`,
      `height:${size}px`,
      'border-radius:50%\n',
      'pointer-events:none',
      'z-index:2147483647',
      'background:#10b981',
      'border:2px solid #ffffff',
      // 初始阴影，动画会在此基础上呼吸
      'box-shadow:0 0 10px rgba(16,185,129,0.55), 0 0 0 0 rgba(16,185,129,0.25)',
      'transform:translate(-50%, -50%)',
      'opacity:0',
      'transition: background-color 120ms ease, opacity 250ms ease',
      // 呼吸动画
      'animation: vm-breathe 2.2s ease-in-out infinite',
      'mix-blend-mode: normal',
    ].join(';')
    document.body.appendChild(el)
    return el
  }

  // 注入虚拟光标呼吸动画 keyframes（仅一次）
  private injectCursorAnimationStyle() {
    if (document.getElementById('vm-cursor-anim-style')) return
    const s = document.createElement('style')
    s.id = 'vm-cursor-anim-style'
    s.textContent = `
      @keyframes vm-breathe {
        0% { box-shadow: 0 0 10px rgba(16,185,129,0.55), 0 0 0 0 rgba(16,185,129,0.25); }
        50% { box-shadow: 0 0 14px rgba(16,185,129,0.85), 0 0 0 8px rgba(16,185,129,0.12); }
        100% { box-shadow: 0 0 10px rgba(16,185,129,0.55), 0 0 0 0 rgba(16,185,129,0.25); }
      }
    `
    try { document.head.appendChild(s) } catch {}
  }

  private setCursorPos(p: Vec2) {
    this.cursor.style.left = `${p.x}px`
    this.cursor.style.top = `${p.y}px`
    
    // 触发虚拟光标的鼠标事件
    this.triggerVirtualMouseEvents(p)
  }

  // 触发虚拟光标的鼠标事件
  private triggerVirtualMouseEvents(pos: Vec2) {
    // 获取虚拟光标位置下的元素
    const element = document.elementFromPoint(pos.x, pos.y)
    if (!element) return
    
    // 检查是否需要触发hover事件
    const lastElement = this.lastHoveredElement
    if (lastElement !== element) {
      // 记录并清理上一个元素（由事件自身负责样式）
      // 触发mouseleave事件
      if (lastElement) {
        const leaveEvent = new MouseEvent('mouseleave', {
          bubbles: true,
          cancelable: true,
          clientX: pos.x,
          clientY: pos.y,
          relatedTarget: element
        })
        lastElement.dispatchEvent(leaveEvent)
        
        // 也触发mouseout事件
        const outEvent = new MouseEvent('mouseout', {
          bubbles: true,
          cancelable: true,
          clientX: pos.x,
          clientY: pos.y,
          relatedTarget: element
        })
        lastElement.dispatchEvent(outEvent)
      }
      
      // 触发mouseenter事件
      const enterEvent = new MouseEvent('mouseenter', {
        bubbles: true,
        cancelable: true,
        clientX: pos.x,
        clientY: pos.y,
        relatedTarget: lastElement
      })
      element.dispatchEvent(enterEvent)
      
      // 触发mouseover事件
      const overEvent = new MouseEvent('mouseover', {
        bubbles: true,
        cancelable: true,
        clientX: pos.x,
        clientY: pos.y,
        relatedTarget: lastElement
      })
      element.dispatchEvent(overEvent)
      
      // 触发mousemove事件（某些组件可能需要这个来更新hover状态）
      const moveEvent = new MouseEvent('mousemove', {
        bubbles: true,
        cancelable: true,
        clientX: pos.x,
        clientY: pos.y,
        relatedTarget: lastElement
      })
      element.dispatchEvent(moveEvent)
      
      this.lastHoveredElement = element
    } else if (element) {
      // 即使元素没有变化，也触发mousemove事件来保持hover状态
      const moveEvent = new MouseEvent('mousemove', {
        bubbles: true,
        cancelable: true,
        clientX: pos.x,
        clientY: pos.y
      })
      element.dispatchEvent(moveEvent)
    }
  }

  // 已移除伪 hover 样式与类名映射，仅保留事件分发以触发组件自身的 hover 逻辑



  // 设置鼠标样式为虚拟光标样式
  private setVirtualCursorStyle() {
    const cursorSize = this.opts.cursorSize
    const cursorColor = '#10b981'
    const borderColor = '#ffffff'
    
    // 创建 SVG 数据 URL
    const svg = `
      <svg width="${cursorSize}" height="${cursorSize}" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="1.8" result="coloredBlur"/>
            <feMerge>
              <feMergeNode in="coloredBlur"/>
              <feMergeNode in="SourceGraphic"/>
            </feMerge>
          </filter>
        </defs>
        <circle cx="${cursorSize/2}" cy="${cursorSize/2}" r="${cursorSize/2-2}"
                fill="${cursorColor}" stroke="${borderColor}" stroke-width="2" filter="url(#glow)"/>
        <circle cx="${cursorSize/2}" cy="${cursorSize/2}" r="${cursorSize/2-4}" fill="${cursorColor}"/>
      </svg>
    `
    
    const dataUrl = `data:image/svg+xml;base64,${btoa(svg)}`
    
    // 设置全局鼠标样式，包括所有状态
    const cursorStyle = `url('${dataUrl}') ${cursorSize/2} ${cursorSize/2}, auto`
    document.body.style.cursor = cursorStyle
    
    // 为所有可交互元素设置相同的鼠标样式
    const style = document.createElement('style')
    style.id = 'virtual-mouse-cursor-style'
    style.textContent = `
      * {
        cursor: ${cursorStyle} !important;
      }
    `
    
    // 移除旧的样式（如果存在）
    const oldStyle = document.getElementById('virtual-mouse-cursor-style')
    if (oldStyle) {
      oldStyle.remove()
    }
    
    document.head.appendChild(style)
    
    vmLog('[VM] 设置全局鼠标样式为虚拟光标样式')
  }

  // 设置鼠标样式为透明（隐藏鼠标）
  private setTransparentCursorStyle() {
    // 直接设置为none，完全隐藏鼠标光标
    document.body.style.cursor = 'none'
    
    // 注入全局样式，隐藏所有元素上的系统光标（包含手指/禁用等所有变体）
    this.ensureHideSystemCursorStyle()

    // 移除虚拟光标样式表
    const style = document.getElementById('virtual-mouse-cursor-style')
    if (style) {
      style.remove()
    }
    
    vmLog('[VM] 设置鼠标样式为透明（none）')
  }

  // 恢复默认鼠标样式
  private restoreDefaultCursorStyle() {
    document.body.style.cursor = ''
    
    // 移除自定义鼠标样式
    const style = document.getElementById('virtual-mouse-cursor-style')
    if (style) {
      style.remove()
    }
    const hideStyle = document.getElementById('vm-hide-system-cursor')
    if (hideStyle) {
      hideStyle.remove()
    }
    
    vmLog('[VM] 恢复默认鼠标样式')
  }

  // 确保隐藏系统光标样式存在；若被路由切换/框架重渲染移除，自动补回
  private ensureHideSystemCursorStyle() {
    if (!document.getElementById('vm-hide-system-cursor')) {
      const style = document.createElement('style')
      style.id = 'vm-hide-system-cursor'
      style.textContent = `* { cursor: none !important; }`
      try { document.head.appendChild(style) } catch {}
    }
    // 处理同源 iframe
    const iframes = Array.from(document.querySelectorAll('iframe')) as HTMLIFrameElement[]
    for (const f of iframes) {
      try {
        const doc = f.contentDocument
        if (!doc) continue
        if (!doc.getElementById('vm-hide-system-cursor')) {
          const s = doc.createElement('style')
          s.id = 'vm-hide-system-cursor'
          s.textContent = `* { cursor: none !important; }`
          doc.head?.appendChild(s)
        }
      } catch {}
    }
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
    // 获取调用堆栈信息
    const stack = new Error().stack
    const callerInfo = stack?.split('\n')[2]?.trim() || 'unknown'
    
    vmLog('[VM][registerActivity] 触发显示原因:', {
      hiddenByInputFocus: this.hiddenByInputFocus,
      hiddenByInactivity: this.hiddenByInactivity,
      currentOpacity: this.cursor.style.opacity,
      callerInfo,
      timestamp: Date.now()
    })
    
    // 若由输入框隐藏则不显示
    if (!this.hiddenByInputFocus) {
      vmLog('[VM][registerActivity] 显示光标 (非输入框隐藏状态)')
      this.showCursor(250)
    } else {
      vmLog('[VM][registerActivity] 跳过显示 (输入框隐藏状态)')
    }
    
    // 重置 10s 无操作隐藏
    if (this.inactivityTimer) { window.clearTimeout(this.inactivityTimer); this.inactivityTimer = null }
    this.inactivityTimer = window.setTimeout(() => {
      // 仅当没有方向键按住时才隐藏
      if (!(this.held['ArrowLeft'] || this.held['ArrowRight'] || this.held['ArrowUp'] || this.held['ArrowDown'])) {
        vmLog('[VM][registerActivity] 设置无操作隐藏定时器 (10s)')
        this.hideCursor(1000, 'inactivity')
      } else {
        vmLog('[VM][registerActivity] 跳过无操作隐藏 (方向键按住中)')
      }
    }, this.inactivityMs)

    // 首次使用事件（供外部弹窗提示）
    if (!(window as any).__vmFirstUseNotified) {
      (window as any).__vmFirstUseNotified = true
      vmLog('[VM][registerActivity] 触发首次使用事件')
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
      // Android 返回键映射到 Escape
      if (code === 4) return 'Escape'
      return k
    })()
    const isEditable = !!active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA' || active.tagName === 'SELECT' || active.getAttribute('contenteditable') === 'true')

    // 调试日志：按键与状态
    try {
      vmLog('[VM][keydown] 按键按下:', { 
        key: e.key, 
        normKey, 
        repeat: e.repeat, 
        isEditable,
        activeElement: active?.tagName,
        held: { ...this.held },
        timestamp: Date.now() 
      })
    } catch {}

    // Enter：在非输入模式下触发鼠标点击
    if (normKey === 'Enter' && !isEditable) {
      vmLog('[VM][keydown] Enter 键触发点击:', {
        pos: { ...this.pos },
        targetElement: document.elementFromPoint(this.pos.x, this.pos.y)?.tagName,
        timestamp: Date.now()
      })
      e.preventDefault(); e.stopPropagation()
      const el = document.elementFromPoint(this.pos.x, this.pos.y) as HTMLElement | null
      if (el) {
        const ev: MouseEventInit = { bubbles: true, cancelable: true, clientX: this.pos.x, clientY: this.pos.y }
        el.dispatchEvent(new MouseEvent('mousedown', ev))
        el.dispatchEvent(new MouseEvent('mouseup', ev))
        el.dispatchEvent(new MouseEvent('click', ev))
      }
      // 不调用 registerActivity，避免虚拟光标被隐藏
      vmLog('[VM][keydown] Enter 键点击完成，不触发 registerActivity')
      return
    }

    // 输入聚焦时：方向键不接管，仅处理 Esc 退出并显示光标
    if (isEditable) {
      if (normKey === 'Escape') {
        vmLog('[VM][keydown] 输入框退出键 (Esc):', {
          normKey,
          activeElement: active?.tagName,
          timestamp: Date.now()
        })
        e.preventDefault(); e.stopPropagation()
        try { active?.blur() } catch {}
        this.showCursor(250)
      } else {
        vmLog('[VM][keydown] 输入框内按键，跳过处理:', {
          normKey,
          activeElement: active?.tagName,
          timestamp: Date.now()
        })
      }
      return
    }
    if (normKey === 'ArrowLeft' || normKey === 'ArrowRight' || normKey === 'ArrowUp' || normKey === 'ArrowDown') {
      vmLog('[VM][keydown] 方向键移动:', {
        normKey,
        repeat: e.repeat,
        held: { ...this.held },
        vel: { ...this.vel },
        speed: this.speed,
        pos: { ...this.pos },
        timestamp: Date.now()
      })
      // 阻止默认与冒泡，避免焦点移动/页面滚动
      e.preventDefault(); e.stopPropagation()
      // keydown 仅用于移动，双击滚动改由 keyup 识别，更稳定
      this.held[normKey] = true
      this.updateVelocity()
      try { vmLog('[VM] move update', { held: { ...this.held }, vel: { ...this.vel }, speed: this.speed }) } catch {}
      
      // 进入键盘控制模式
      if (!this.isKeyboardMode) {
        this.isKeyboardMode = true
        this.setTransparentCursorStyle()
        vmLog('[VM][keydown] 进入键盘控制模式，隐藏鼠标光标')
      }
      
      // 如果虚拟光标被鼠标隐藏，在鼠标最后位置显示
      if (this.hiddenByMouse) {
        this.hiddenByMouse = false
        this.pos.x = this.lastMousePos.x
        this.pos.y = this.lastMousePos.y
        this.setCursorPos(this.pos)
        this.showCursor(250)
        vmLog('[VM][keydown] 方向键触发，在鼠标最后位置显示虚拟光标:', this.lastMousePos)
      } else {
        this.showCursor(250)
      }
      
      // 确保触发hover事件
      this.triggerVirtualMouseEvents(this.pos)
      
      vmLog('[VM][keydown] 方向键触发 registerActivity')
      this.registerActivity()
    }
    // AndroidTV 返回键兼容（仅在非输入框状态下处理）
    if (normKey === 'Backspace' || normKey === 'BrowserBack') {
      vmLog('[VM][keydown] AndroidTV 返回键:', {
        normKey,
        activeElement: document.activeElement?.tagName,
        isEditable,
        timestamp: Date.now()
      })
      
      // 如果在输入框中，让 Backspace 正常处理（删除文本）
      if (isEditable) {
        vmLog('[VM][keydown] 输入框中的 Backspace，允许正常删除文本')
        return // 不阻止默认行为，让 Backspace 正常删除文本
      }
      
      // 非输入框状态：处理返回逻辑
      e.preventDefault(); e.stopPropagation()
      vmLog('[VM][keydown] 非输入框返回键，触发返回事件')
      // 非输入态：尽量不影响页面路由，交由业务自行处理返回
      // 这里仅发出一个事件供业务层监听
      try { window.dispatchEvent(new CustomEvent('virtual-mouse-back')) } catch {}
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
      // Android 返回键映射到 Escape
      if (code === 4) return 'Escape'
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
      try { 
        vmLog('[VM][keyup] 方向键释放:', { 
          normKey, 
          interval, 
          count: this.keyUpCount[normKey],
          threshold: this.dblPressThreshold,
          held: { ...this.held },
          timestamp: Date.now()
        }) 
      } catch {}
      if (this.keyUpCount[normKey] <= 2 && interval <= this.dblPressThreshold) {
        if (normKey === 'ArrowUp' || normKey === 'ArrowDown') {
          const delta = Math.round(window.innerHeight * 0.6) * (normKey === 'ArrowDown' ? 1 : -1)
          try { 
            vmLog('[VM] 垂直双击滚动触发:', { 
              normKey, 
              delta, 
              interval,
              windowHeight: window.innerHeight,
              timestamp: Date.now()
            }) 
          } catch {}
          this.scrollByDelta(delta)
        } else if (normKey === 'ArrowLeft' || normKey === 'ArrowRight') {
          const deltaX = Math.round(window.innerWidth * 0.6) * (normKey === 'ArrowRight' ? 1 : -1)
          try { 
            vmLog('[VM] 水平双击滚动触发:', { 
              normKey, 
              deltaX, 
              interval,
              windowWidth: window.innerWidth,
              timestamp: Date.now()
            }) 
          } catch {}
          this.scrollByDeltaX(deltaX)
        }
        // 重置，避免第三击误抑制
        this.lastKeyUpTs[normKey] = 0
        this.keyUpCount[normKey] = 0
      } else {
        try {
          vmLog('[VM][keyup] 双击滚动未触发:', {
            normKey,
            interval,
            count: this.keyUpCount[normKey],
            threshold: this.dblPressThreshold,
            reason: this.keyUpCount[normKey] > 2 ? '超过2次点击' : '间隔过长',
            timestamp: Date.now()
          })
        } catch {}
      }
      vmLog('[VM][keyup] 方向键释放触发 registerActivity')
      this.registerActivity()
    }
  }

  enable() {
    if (this.enabled) {
      vmLog('[VM][enable] 虚拟光标已启用，跳过')
      return
    }
    vmLog('[VM][enable] 启用虚拟光标:', {
      opts: this.opts,
      initialPos: { x: window.innerWidth / 2, y: window.innerHeight / 2 },
      timestamp: Date.now()
    })
    this.enabled = true
    // 初始放到视口中心
    this.pos = { x: window.innerWidth / 2, y: window.innerHeight / 2 }
    this.setCursorPos(this.pos)
    
    // 使用自定义虚拟光标作为鼠标指针：隐藏系统鼠标
    this.setTransparentCursorStyle()
    // 观察 <head> 变化，若样式被移除（热更新/路由切换），自动补回
    try {
      const head = document.head
      const mo = new MutationObserver(() => this.ensureHideSystemCursorStyle())
      mo.observe(head, { childList: true, subtree: false })
      ;(this as any)._vmHeadObserver = mo
    } catch {}
    

    
    // 添加全局按键监听器，打印所有物理按键
    this.addGlobalKeyLogger()
    
    // 添加鼠标事件监听器
    this.addMouseEventListeners()
    this.addTouchEventListeners()
    
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
    if (!this.enabled) {
      vmLog('[VM][disable] 虚拟光标未启用，跳过')
      return
    }
    vmLog('[VM][disable] 禁用虚拟光标:', {
      timestamp: Date.now()
    })
    this.enabled = false
    this.stopAcceleration()
    if (this.rafId !== null) { cancelAnimationFrame(this.rafId); this.rafId = null }
    window.removeEventListener('keydown', this.onKeyDown)
    window.removeEventListener('keyup', this.onKeyUp)
    window.removeEventListener('focusin', this.onFocusIn as any)
    window.removeEventListener('focusout', this.onFocusOut as any)
    
    // 移除全局按键监听器
    this.removeGlobalKeyLogger()
    
    // 移除鼠标事件监听器
    this.removeMouseEventListeners()
    this.removeTouchEventListeners()
    
    // 清理鼠标活动定时器
    if (this.mouseActivityTimer) {
      clearTimeout(this.mouseActivityTimer)
      this.mouseActivityTimer = null
    }
    if (this.touchHideTimer) { window.clearTimeout(this.touchHideTimer); this.touchHideTimer = null }
    
    // 恢复默认鼠标样式
    this.restoreDefaultCursorStyle()
    // 断开观察器
    try { (this as any)._vmHeadObserver?.disconnect?.() } catch {}
    
    // 清理hover状态
    this.lastHoveredElement = null
    

    
    try { this.cursor.remove() } catch {}
  }

  private hideCursor(durationMs: number, reason: 'input' | 'inactivity' | 'mouse') {
    if (reason === 'input') this.hiddenByInputFocus = true
    if (reason === 'inactivity') this.hiddenByInactivity = true
    if (reason === 'mouse') this.hiddenByMouse = true
    this.cursor.style.transition = `background-color 120ms ease, opacity ${durationMs}ms ease`
    // 强制 reflow，确保过渡时间更改被浏览器应用
    void this.cursor.offsetWidth
    this.cursor.style.opacity = '0'
  }
  private showCursor(durationMs: number) {
    vmLog('[VM][showCursor] 显示光标:', {
      durationMs,
      hiddenByInputFocus: this.hiddenByInputFocus,
      hiddenByInactivity: this.hiddenByInactivity,
      hiddenByMouse: this.hiddenByMouse,
      hiddenByTouch: this.hiddenByTouch,
      timestamp: Date.now()
    })
    
    this.hiddenByInactivity = false
    if (this.hiddenByInputFocus) {
      vmLog('[VM][showCursor] 跳过显示 (输入框隐藏状态)')
      return // 输入态仍保持隐藏
    }
    if (this.hiddenByMouse) {
      vmLog('[VM][showCursor] 跳过显示 (鼠标隐藏状态)')
      return // 鼠标活动时保持隐藏
    }
    if (this.hiddenByTouch) {
      return // 触摸时保持隐藏
    }
    this.cursor.style.transition = `background-color 120ms ease, opacity ${durationMs}ms ease`
    void this.cursor.offsetWidth
    this.cursor.style.opacity = '1'
  }
  private onFocusIn = (e: FocusEvent) => {
    const t = e.target as HTMLElement | null
    if (!t) return
    const isEditable = t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.getAttribute('contenteditable') === 'true'
    vmLog('[VM][onFocusIn] 焦点进入:', {
      tagName: t.tagName,
      isEditable,
      timestamp: Date.now()
    })
    if (isEditable) {
      this.hideCursor(250, 'input')
    }
  }
  private onFocusOut = (e: FocusEvent) => { 
    const t = e.target as HTMLElement | null
    if (!t) return
    const isEditable = t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.getAttribute('contenteditable') === 'true'
    
    vmLog('[VM][onFocusOut] 焦点离开:', {
      tagName: t.tagName,
      isEditable,
      timestamp: Date.now()
    })
    
    // 删除焦点离开时显示光标的逻辑，只保留方向键触发
    if (isEditable) {
      this.hiddenByInputFocus = false
      // 不再自动显示光标，等待方向键触发
    }
  }

  // 全局按键监听器，打印所有物理按键
  private globalKeyLogger = (e: Event) => {
    // 只处理键盘事件
    if (!(e instanceof KeyboardEvent)) return
    
    const ke = e as KeyboardEvent
    vmLog('[VM][GLOBAL_KEY] 物理按键检测:', {
      key: ke.key,
      code: ke.code,
      keyCode: ke.keyCode,
      which: ke.which,
      charCode: ke.charCode,
      type: ke.type,
      repeat: ke.repeat,
      ctrlKey: ke.ctrlKey,
      shiftKey: ke.shiftKey,
      altKey: ke.altKey,
      metaKey: ke.metaKey,
      location: ke.location,
      isComposing: ke.isComposing,
      timeStamp: ke.timeStamp,
      target: (ke.target as HTMLElement)?.tagName,
      activeElement: document.activeElement?.tagName,
      timestamp: Date.now()
    })
    
    // 如果检测到 Android 环境，将按键事件发送到 Android 端
    if ((window as any).AndroidKeyEvent) {
      try {
        (window as any).AndroidKeyEvent.onKeyEvent(
          ke.type,
          ke.keyCode,
          ke.key,
          ke.ctrlKey,
          ke.shiftKey,
          ke.altKey,
          ke.metaKey
        )
      } catch (error) {
        vmLog('[VM][GLOBAL_KEY] Android 按键事件发送失败:', error)
      }
    }
  }

  private addGlobalKeyLogger() {
    vmLog('[VM] 添加全局按键监听器')
    
    // 监听多种事件类型，包括可能的 Android 事件
    const events = ['keydown', 'keyup', 'keypress', 'input', 'beforeinput']
    
    events.forEach(eventType => {
      window.addEventListener(eventType, this.globalKeyLogger, { capture: true, passive: true })
      document.addEventListener(eventType, this.globalKeyLogger, { capture: true, passive: true })
      document.body.addEventListener(eventType, this.globalKeyLogger, { capture: true, passive: true })
    })
    
    // 监听可能的 Android 特定事件
    const androidEvents = ['backbutton', 'menubutton', 'searchbutton', 'volumeup', 'volumedown']
    androidEvents.forEach(eventType => {
      document.addEventListener(eventType, this.globalKeyLogger, { capture: true, passive: true })
    })
    
    // Android 按键事件注入接口（由 Android 端直接调用）
    if ((window as any).AndroidKeyEvent) {
      vmLog('[VM] Android 按键事件接口已可用')
    }
    
    // 测试监听器是否正常工作
    vmLog('[VM] 全局按键监听器已添加，监听事件类型:', events.concat(androidEvents))
    vmLog('[VM] 请按任意键测试，如果看不到日志，可能是 Android 系统拦截了按键事件')
  }

  // 新增：添加鼠标事件监听器
  private addMouseEventListeners() {
    vmLog('[VM] 添加鼠标事件监听器')
    
    // 监听所有鼠标事件
    const mouseEvents = [
      'mousedown', 'mouseup', 'mousemove', 'mouseenter', 'mouseleave',
      'mouseover', 'mouseout', 'click', 'dblclick', 'contextmenu',
      'wheel', 'scroll'
    ]
    
    mouseEvents.forEach(eventType => {
      document.addEventListener(eventType, this.onMouseActivity, { passive: true })
      window.addEventListener(eventType, this.onMouseActivity, { passive: true })
    })
    
    vmLog('[VM] 鼠标事件监听器已添加，监听事件类型:', mouseEvents)
  }

  // 新增：鼠标活动处理（用虚拟光标跟随鼠标，实现与键盘一致的视觉）
  private onMouseActivity = (e: Event) => {
    // 检查是否是由虚拟鼠标系统触发的事件
    if (e instanceof MouseEvent && e.detail === 0 && e.isTrusted === false) {
      return
    }
    
    // 跟随鼠标移动虚拟光标，并显示
    if (e instanceof MouseEvent) {
      // 若鼠标移出可视区域，立即隐藏虚拟光标
      const outOfViewport = (
        e.clientX < 0 || e.clientY < 0 ||
        e.clientX > window.innerWidth || e.clientY > window.innerHeight
      )
      if (outOfViewport) {
        this.hideCursor(120, 'mouse')
        return
      }
      this.lastMousePos.x = e.clientX
      this.lastMousePos.y = e.clientY
      this.pos.x = e.clientX
      this.pos.y = e.clientY
      this.setCursorPos(this.pos)
      this.showCursor(180)
      this.isKeyboardMode = false
      this.hiddenByMouse = false
    }
    // 不再基于鼠标活动隐藏虚拟光标，移除定时器
    if (this.mouseActivityTimer) { clearTimeout(this.mouseActivityTimer); this.mouseActivityTimer = null }
  }

  // 同步虚拟光标位置到鼠标位置
  private syncCursorToMouse(mouseX: number, mouseY: number) {
    // 更新虚拟光标位置到鼠标位置
    this.pos.x = mouseX
    this.pos.y = mouseY
    this.setCursorPos(this.pos)
    
    vmLog('[VM] 同步虚拟光标位置到鼠标:', {
      mouseX,
      mouseY,
      virtualPos: { ...this.pos },
      timestamp: Date.now()
    })
  }



  // 新增：移除鼠标事件监听器
  private removeMouseEventListeners() {
    vmLog('[VM] 移除鼠标事件监听器')
    
    const mouseEvents = [
      'mousedown', 'mouseup', 'mousemove', 'mouseenter', 'mouseleave',
      'mouseover', 'mouseout', 'click', 'dblclick', 'contextmenu',
      'wheel', 'scroll'
    ]
    
    mouseEvents.forEach(eventType => {
      document.removeEventListener(eventType, this.onMouseActivity)
      window.removeEventListener(eventType, this.onMouseActivity)
    })
  }

  // 触摸事件：按下/移动时隐藏虚拟光标，结束后短时间内不显示
  private onTouchStart = (_e: TouchEvent) => {
    this.hiddenByTouch = true
    this.hideCursor(120, 'mouse')
    if (this.touchHideTimer) { window.clearTimeout(this.touchHideTimer); this.touchHideTimer = null }
  }

  private onTouchMove = (_e: TouchEvent) => {
    this.hiddenByTouch = true
    this.hideCursor(120, 'mouse')
  }

  private onTouchEnd = (_e: TouchEvent) => {
    // 延迟恢复允许显示，防止手指抬起瞬间出现
    if (this.touchHideTimer) { window.clearTimeout(this.touchHideTimer) }
    this.touchHideTimer = window.setTimeout(() => {
      this.hiddenByTouch = false
      this.touchHideTimer = null
    }, 800)
  }

  private addTouchEventListeners() {
    const opts: AddEventListenerOptions = { passive: true }
    window.addEventListener('touchstart', this.onTouchStart, opts)
    window.addEventListener('touchmove', this.onTouchMove, opts)
    window.addEventListener('touchend', this.onTouchEnd, opts)
    window.addEventListener('touchcancel', this.onTouchEnd, opts)
  }

  private removeTouchEventListeners() {
    window.removeEventListener('touchstart', this.onTouchStart)
    window.removeEventListener('touchmove', this.onTouchMove)
    window.removeEventListener('touchend', this.onTouchEnd)
    window.removeEventListener('touchcancel', this.onTouchEnd)
  }

  private removeGlobalKeyLogger() {
    vmLog('[VM] 移除全局按键监听器')
    
    // 移除所有事件监听器
    const events = ['keydown', 'keyup', 'keypress', 'input', 'beforeinput']
    const androidEvents = ['backbutton', 'menubutton', 'searchbutton', 'volumeup', 'volumedown']
    
    events.forEach(eventType => {
      window.removeEventListener(eventType, this.globalKeyLogger, { capture: true })
      document.removeEventListener(eventType, this.globalKeyLogger, { capture: true })
      document.body.removeEventListener(eventType, this.globalKeyLogger, { capture: true })
    })
    
    androidEvents.forEach(eventType => {
      document.removeEventListener(eventType, this.globalKeyLogger, { capture: true })
    })
  }

  // 公共方法：检查是否启用
  public isEnabled(): boolean {
    return this.enabled
  }
}

// 工具：便捷启用（按需加载）
export const initVirtualMouse = (opts?: VirtualMouseOptions) => {
  const vm = new VirtualMouse(opts)
  vm.enable()
  ;(window as any).__virtualMouse = vm
  return vm
}

// 全局虚拟光标管理函数
export const enableVirtualMouse = (opts?: VirtualMouseOptions) => {
  // 如果已经存在实例，先禁用
  if (globalVirtualMouse) {
    globalVirtualMouse.disable()
  }
  
  // 创建新实例并启用
  globalVirtualMouse = new VirtualMouse(opts)
  globalVirtualMouse.enable()
  ;(window as any).__virtualMouse = globalVirtualMouse
  
  return globalVirtualMouse
}

export const disableVirtualMouse = () => {
  if (globalVirtualMouse) {
    globalVirtualMouse.disable()
    globalVirtualMouse = null
    ;(window as any).__virtualMouse = null
  }
}

export const isVirtualMouseEnabled = () => {
  return globalVirtualMouse !== null && globalVirtualMouse.isEnabled()
}

export const getVirtualMouseInstance = () => {
  return globalVirtualMouse
}


