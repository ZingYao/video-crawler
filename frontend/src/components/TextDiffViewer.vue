<template>
  <div class="text-diff-viewer">
    <div class="diff-header">
      <div class="diff-title-left">{{ leftTitle }}</div>
      <div class="diff-title-right">{{ rightTitle }}</div>
      <div class="diff-controls">
        <a-button size="small" @click="toggleUnchanged">
          {{ showUnchanged ? '隐藏相同' : '显示相同' }}
        </a-button>
      </div>
    </div>
    <div class="diff-content">
      <div class="diff-pane left-pane">
        <div class="diff-lines">
          <div
            v-for="(line, index) in visibleLeftLines"
            :key="`left-${index}`"
            class="diff-line"
            :class="getLineClass(line)"
          >
            <span class="line-number">{{ line.originalIndex + 1 }}</span>
            <span class="line-content" v-html="escapeHtml(line.content)"></span>
          </div>
        </div>
      </div>
      <div class="diff-pane right-pane">
        <div class="diff-lines">
          <div
            v-for="(line, index) in visibleRightLines"
            :key="`right-${index}`"
            class="diff-line"
            :class="getLineClass(line)"
          >
            <span class="line-number">{{ line.originalIndex + 1 }}</span>
            <span class="line-content" v-html="escapeHtml(line.content)"></span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { diffWords } from 'diff'

interface DiffLine {
  content: string
  type: 'added' | 'removed' | 'unchanged'
  originalIndex: number
}

interface Props {
  oldText: string
  newText: string
  leftTitle?: string
  rightTitle?: string
  showAllByDefault?: boolean
  contextLines?: number
}

const props = withDefaults(defineProps<Props>(), {
  leftTitle: '原始结果',
  rightTitle: '转换结果',
  showAllByDefault: false,
  contextLines: 10
})

// 控制是否显示相同内容
const showUnchanged = ref(props.showAllByDefault)

const toggleUnchanged = () => {
  showUnchanged.value = !showUnchanged.value
}

// 计算差异
const diffResult = computed(() => {
  return diffWords(props.oldText, props.newText)
})

// 生成左侧行（原始文本）
const leftLines = computed(() => {
  const lines: DiffLine[] = []
  let lineNumber = 1
  
  diffResult.value.forEach(part => {
    if (part.added) {
      // 新增的内容在右侧显示，左侧显示空行
      lines.push({ content: '', type: 'unchanged', originalIndex: lineNumber - 1 })
    } else if (part.removed) {
      // 删除的内容在左侧显示为红色
      const partLines = part.value.split('\n')
      partLines.forEach(line => {
        if (line.trim() !== '') {
          lines.push({ content: line, type: 'removed', originalIndex: lineNumber - 1 })
          lineNumber++
        }
      })
    } else {
      // 未改变的内容
      const partLines = part.value.split('\n')
      partLines.forEach(line => {
        if (line.trim() !== '') {
          lines.push({ content: line, type: 'unchanged', originalIndex: lineNumber - 1 })
          lineNumber++
        }
      })
    }
  })
  
  return lines
})

// 生成右侧行（新文本）
const rightLines = computed(() => {
  const lines: DiffLine[] = []
  let lineNumber = 1
  
  diffResult.value.forEach(part => {
    if (part.removed) {
      // 删除的内容在右侧显示空行
      lines.push({ content: '', type: 'unchanged', originalIndex: lineNumber - 1 })
    } else if (part.added) {
      // 新增的内容在右侧显示为绿色
      const partLines = part.value.split('\n')
      partLines.forEach(line => {
        if (line.trim() !== '') {
          lines.push({ content: line, type: 'added', originalIndex: lineNumber - 1 })
          lineNumber++
        }
      })
    } else {
      // 未改变的内容
      const partLines = part.value.split('\n')
      partLines.forEach(line => {
        if (line.trim() !== '') {
          lines.push({ content: line, type: 'unchanged', originalIndex: lineNumber - 1 })
          lineNumber++
        }
      })
    }
  })
  
  return lines
})

// 获取可见的左侧行（根据折叠设置过滤，包含上下文）
const visibleLeftLines = computed(() => {
  if (showUnchanged.value) {
    return leftLines.value
  }
  
  const lines = leftLines.value
  const contextLines = props.contextLines
  const visibleIndices = new Set<number>()
  
  // 找到所有差异行的索引
  lines.forEach((line, index) => {
    if (line.type !== 'unchanged' || line.content.trim() === '') {
      // 添加差异行本身
      visibleIndices.add(index)
      
      // 添加上下文行
      for (let i = Math.max(0, index - contextLines); i <= Math.min(lines.length - 1, index + contextLines); i++) {
        visibleIndices.add(i)
      }
    }
  })
  
  // 返回可见的行，保持原始顺序
  return lines.filter((_, index) => visibleIndices.has(index))
})

// 获取可见的右侧行（根据折叠设置过滤，包含上下文）
const visibleRightLines = computed(() => {
  if (showUnchanged.value) {
    return rightLines.value
  }
  
  const lines = rightLines.value
  const contextLines = props.contextLines
  const visibleIndices = new Set<number>()
  
  // 找到所有差异行的索引
  lines.forEach((line, index) => {
    if (line.type !== 'unchanged' || line.content.trim() === '') {
      // 添加差异行本身
      visibleIndices.add(index)
      
      // 添加上下文行
      for (let i = Math.max(0, index - contextLines); i <= Math.min(lines.length - 1, index + contextLines); i++) {
        visibleIndices.add(i)
      }
    }
  })
  
  // 返回可见的行，保持原始顺序
  return lines.filter((_, index) => visibleIndices.has(index))
})

// 获取行的CSS类
const getLineClass = (line: DiffLine) => {
  return {
    'line-added': line.type === 'added',
    'line-removed': line.type === 'removed',
    'line-unchanged': line.type === 'unchanged'
  }
}

// HTML转义
const escapeHtml = (text: string): string => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}
</script>

<style scoped>
.text-diff-viewer {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  background: #00000000;
  font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
}

.diff-header {
  display: flex;
  background: #f5f5f5;
  border-bottom: 1px solid #d9d9d9;
  border-radius: 6px 6px 0 0;
}

.diff-title-left,
.diff-title-right {
  flex: 1;
  padding: 8px 12px;
  font-weight: 600;
  color: #333;
  text-align: center;
  border-right: 1px solid #d9d9d9;
}

.diff-title-right {
  border-right: none;
}

.diff-controls {
  padding: 8px 12px;
  border-left: 1px solid #d9d9d9;
  display: flex;
  align-items: center;
}

.diff-content {
  display: flex;
  max-height: 600px;
  overflow: auto;
  background: #00000000;
}

.diff-pane {
  flex: 1;
  border-right: 1px solid #d9d9d9;
}

.diff-pane:last-child {
  border-right: none;
}

.diff-lines {
  padding: 8px 0;
}

.diff-line {
  display: flex;
  padding: 2px 12px;
  white-space: pre-wrap;
  word-break: break-word;
}

.diff-line:hover {
  background: #f7f7f7;
}

.line-number {
  min-width: 40px;
  color: #999;
  font-size: 11px;
  text-align: right;
  margin-right: 12px;
  user-select: none;
}

.line-content {
  flex: 1;
  color: #333;
}

.line-added {
  background: #e6ffed;
}

.line-added .line-content {
  color: #22863a;
}

.line-removed {
  background: #ffeef0;
}

.line-removed .line-content {
  color: #cb2431;
  text-decoration: line-through;
}

.line-unchanged {
  background: transparent;
}

.line-unchanged .line-content {
  color: #000000;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .diff-content {
    flex-direction: column;
  }
  
  .diff-pane {
    border-right: none;
    border-bottom: 1px solid #d9d9d9;
  }
  
  .diff-pane:last-child {
    border-bottom: none;
  }
  
  .diff-title-left,
  .diff-title-right {
    border-right: none;
    border-bottom: 1px solid #d9d9d9;
  }
  
  .diff-title-right {
    border-bottom: none;
  }
}
</style>
