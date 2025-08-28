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
    // 获取调用堆栈信息
    const stack = new Error().stack
    const callerInfo = stack?.split('\n')[2]?.trim() || 'unknown'
    
    console.log('[VM][registerActivity] 触发显示原因:', {
      hiddenByInputFocus: this.hiddenByInputFocus,
      hiddenByInactivity: this.hiddenByInactivity,
      currentOpacity: this.cursor.style.opacity,
      callerInfo,
      timestamp: Date.now()
    })
    
    // 若由输入框隐藏则不显示
    if (!this.hiddenByInputFocus) {
      console.log('[VM][registerActivity] 显示光标 (非输入框隐藏状态)')
      this.showCursor(250)
    } else {
      console.log('[VM][registerActivity] 跳过显示 (输入框隐藏状态)')
    }
    
    // 重置 10s 无操作隐藏
    if (this.inactivityTimer) { window.clearTimeout(this.inactivityTimer); this.inactivityTimer = null }
    this.inactivityTimer = window.setTimeout(() => {
      // 仅当没有方向键按住时才隐藏
      if (!(this.held['ArrowLeft'] || this.held['ArrowRight'] || this.held['ArrowUp'] || this.held['ArrowDown'])) {
        console.log('[VM][registerActivity] 设置无操作隐藏定时器 (10s)')
        this.hideCursor(1000, 'inactivity')
      } else {
        console.log('[VM][registerActivity] 跳过无操作隐藏 (方向键按住中)')
      }
    }, this.inactivityMs)

    // 首次使用事件（供外部弹窗提示）
    if (!(window as any).__vmFirstUseNotified) {
      (window as any).__vmFirstUseNotified = true
      console.log('[VM][registerActivity] 触发首次使用事件')
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
      console.log('[VM][keydown] 按键按下:', { 
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
      console.log('[VM][keydown] Enter 键触发点击:', {
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
      console.log('[VM][keydown] Enter 键点击完成，不触发 registerActivity')
      return
    }

    // 输入聚焦时：方向键不接管，仅处理 Esc 退出并显示光标
    if (isEditable) {
      if (normKey === 'Escape') {
        console.log('[VM][keydown] 输入框退出键 (Esc):', {
          normKey,
          activeElement: active?.tagName,
          timestamp: Date.now()
        })
        e.preventDefault(); e.stopPropagation()
        try { active?.blur() } catch {}
        this.showCursor(250)
      } else {
        console.log('[VM][keydown] 输入框内按键，跳过处理:', {
          normKey,
          activeElement: active?.tagName,
          timestamp: Date.now()
        })
      }
      return
    }
    if (normKey === 'ArrowLeft' || normKey === 'ArrowRight' || normKey === 'ArrowUp' || normKey === 'ArrowDown') {
      console.log('[VM][keydown] 方向键移动:', {
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
      try { console.log('[VM] move update', { held: { ...this.held }, vel: { ...this.vel }, speed: this.speed }) } catch {}
      
      // 即使光标隐藏也要显示光标并更新位置
      if (this.hiddenByMouse) {
        this.hiddenByMouse = false
        this.showCursor(250)
        console.log('[VM][keydown] 方向键触发，重新显示虚拟光标')
      } else {
        this.showCursor(250)
      }
      
      console.log('[VM][keydown] 方向键触发 registerActivity')
      this.registerActivity()
    }
    // AndroidTV 返回键兼容（仅在非输入框状态下处理）
    if (normKey === 'Backspace' || normKey === 'BrowserBack') {
      console.log('[VM][keydown] AndroidTV 返回键:', {
        normKey,
        activeElement: document.activeElement?.tagName,
        isEditable,
        timestamp: Date.now()
      })
      
      // 如果在输入框中，让 Backspace 正常处理（删除文本）
      if (isEditable) {
        console.log('[VM][keydown] 输入框中的 Backspace，允许正常删除文本')
        return // 不阻止默认行为，让 Backspace 正常删除文本
      }
      
      // 非输入框状态：处理返回逻辑
      e.preventDefault(); e.stopPropagation()
      console.log('[VM][keydown] 非输入框返回键，触发返回事件')
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
        console.log('[VM][keyup] 方向键释放:', { 
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
            console.log('[VM] 垂直双击滚动触发:', { 
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
            console.log('[VM] 水平双击滚动触发:', { 
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
          console.log('[VM][keyup] 双击滚动未触发:', {
            normKey,
            interval,
            count: this.keyUpCount[normKey],
            threshold: this.dblPressThreshold,
            reason: this.keyUpCount[normKey] > 2 ? '超过2次点击' : '间隔过长',
            timestamp: Date.now()
          })
        } catch {}
      }
      console.log('[VM][keyup] 方向键释放触发 registerActivity')
      this.registerActivity()
    }
  }

  enable() {
    if (this.enabled) {
      console.log('[VM][enable] 虚拟光标已启用，跳过')
      return
    }
    console.log('[VM][enable] 启用虚拟光标:', {
      opts: this.opts,
      initialPos: { x: window.innerWidth / 2, y: window.innerHeight / 2 },
      timestamp: Date.now()
    })
    this.enabled = true
    // 初始放到视口中心
    this.pos = { x: window.innerWidth / 2, y: window.innerHeight / 2 }
    this.setCursorPos(this.pos)
    

    
    // 添加全局按键监听器，打印所有物理按键
    this.addGlobalKeyLogger()
    
    // 添加鼠标事件监听器
    this.addMouseEventListeners()
    
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
      console.log('[VM][disable] 虚拟光标未启用，跳过')
      return
    }
    console.log('[VM][disable] 禁用虚拟光标:', {
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
    
    // 清理鼠标活动定时器
    if (this.mouseActivityTimer) {
      clearTimeout(this.mouseActivityTimer)
      this.mouseActivityTimer = null
    }
    

    
    try { this.cursor.remove() } catch {}
  }

  private hideCursor(durationMs: number, reason: 'input' | 'inactivity' | 'mouse') {
    console.log('[VM][hideCursor] 隐藏光标:', {
      reason,
      durationMs,
      hiddenByInputFocus: this.hiddenByInputFocus,
      hiddenByInactivity: this.hiddenByInactivity,
      hiddenByMouse: this.hiddenByMouse,
      timestamp: Date.now()
    })
    
    if (reason === 'input') this.hiddenByInputFocus = true
    if (reason === 'inactivity') this.hiddenByInactivity = true
    if (reason === 'mouse') this.hiddenByMouse = true
    this.cursor.style.transition = `background-color 120ms ease, opacity ${durationMs}ms ease`
    // 强制 reflow，确保过渡时间更改被浏览器应用
    void this.cursor.offsetWidth
    this.cursor.style.opacity = '0'
  }
  private showCursor(durationMs: number) {
    console.log('[VM][showCursor] 显示光标:', {
      durationMs,
      hiddenByInputFocus: this.hiddenByInputFocus,
      hiddenByInactivity: this.hiddenByInactivity,
      hiddenByMouse: this.hiddenByMouse,
      timestamp: Date.now()
    })
    
    this.hiddenByInactivity = false
    if (this.hiddenByInputFocus) {
      console.log('[VM][showCursor] 跳过显示 (输入框隐藏状态)')
      return // 输入态仍保持隐藏
    }
    if (this.hiddenByMouse) {
      console.log('[VM][showCursor] 跳过显示 (鼠标隐藏状态)')
      return // 鼠标活动时保持隐藏
    }
    this.cursor.style.transition = `background-color 120ms ease, opacity ${durationMs}ms ease`
    void this.cursor.offsetWidth
    this.cursor.style.opacity = '1'
  }
  private onFocusIn = (e: FocusEvent) => {
    const t = e.target as HTMLElement | null
    if (!t) return
    const isEditable = t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.getAttribute('contenteditable') === 'true'
    console.log('[VM][onFocusIn] 焦点进入:', {
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
    
    console.log('[VM][onFocusOut] 焦点离开:', {
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
    console.log('[VM][GLOBAL_KEY] 物理按键检测:', {
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
        console.log('[VM][GLOBAL_KEY] Android 按键事件发送失败:', error)
      }
    }
  }

  private addGlobalKeyLogger() {
    console.log('[VM] 添加全局按键监听器')
    
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
      console.log('[VM] Android 按键事件接口已可用')
    }
    
    // 测试监听器是否正常工作
    console.log('[VM] 全局按键监听器已添加，监听事件类型:', events.concat(androidEvents))
    console.log('[VM] 请按任意键测试，如果看不到日志，可能是 Android 系统拦截了按键事件')
  }

  // 新增：添加鼠标事件监听器
  private addMouseEventListeners() {
    console.log('[VM] 添加鼠标事件监听器')
    
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
    
    // 监听触摸事件（移动设备）
    const touchEvents = ['touchstart', 'touchmove', 'touchend', 'touchcancel']
    touchEvents.forEach(eventType => {
      document.addEventListener(eventType, this.onMouseActivity, { passive: true })
      window.addEventListener(eventType, this.onMouseActivity, { passive: true })
    })
    
    console.log('[VM] 鼠标事件监听器已添加，监听事件类型:', mouseEvents.concat(touchEvents))
  }

  // 新增：鼠标活动处理
  private onMouseActivity = (e: Event) => {
    // 检查是否是由虚拟鼠标系统触发的事件
    if (e instanceof MouseEvent && e.detail === 0 && e.isTrusted === false) {
      console.log('[VM] 检测到虚拟鼠标事件，跳过隐藏逻辑')
      return
    }
    
    // 清除鼠标无活动定时器
    if (this.mouseActivityTimer) {
      clearTimeout(this.mouseActivityTimer)
      this.mouseActivityTimer = null
    }
    
    // 如果虚拟光标未被鼠标隐藏，立即隐藏
    if (!this.hiddenByMouse) {
      this.hideCursor(250, 'mouse')
      console.log('[VM] 检测到鼠标活动，隐藏虚拟光标')
    }
    
    // 设置鼠标无活动定时器
    this.mouseActivityTimer = window.setTimeout(() => {
      this.hiddenByMouse = false
      console.log('[VM] 鼠标无活动超时，允许显示虚拟光标')
    }, this.mouseInactivityMs)
  }

  // 同步虚拟光标位置到鼠标位置
  private syncCursorToMouse(mouseX: number, mouseY: number) {
    // 更新虚拟光标位置到鼠标位置
    this.pos.x = mouseX
    this.pos.y = mouseY
    this.setCursorPos(this.pos)
    
    console.log('[VM] 同步虚拟光标位置到鼠标:', {
      mouseX,
      mouseY,
      virtualPos: { ...this.pos },
      timestamp: Date.now()
    })
  }



  // 新增：移除鼠标事件监听器
  private removeMouseEventListeners() {
    console.log('[VM] 移除鼠标事件监听器')
    
    const mouseEvents = [
      'mousedown', 'mouseup', 'mousemove', 'mouseenter', 'mouseleave',
      'mouseover', 'mouseout', 'click', 'dblclick', 'contextmenu',
      'wheel', 'scroll'
    ]
    
    const touchEvents = ['touchstart', 'touchmove', 'touchend', 'touchcancel']
    
    mouseEvents.forEach(eventType => {
      document.removeEventListener(eventType, this.onMouseActivity)
      window.removeEventListener(eventType, this.onMouseActivity)
    })
    
    touchEvents.forEach(eventType => {
      document.removeEventListener(eventType, this.onMouseActivity)
      window.removeEventListener(eventType, this.onMouseActivity)
    })
  }

  private removeGlobalKeyLogger() {
    console.log('[VM] 移除全局按键监听器')
    
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


