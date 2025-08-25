// 本地历史记录管理工具
// 用于在无需登录模式下缓存观看记录

export interface LocalVideoHistory {
  id: string
  video_id: string
  video_title: string
  video_url: string
  source_id: string
  source_name: string
  watch_time: number // 观看时长（秒）
  progress: number // 观看进度（0-1）
  created_at: string
  updated_at: string
}

export interface LocalSearchHistory {
  id: string
  keyword: string
  source_id: string
  created_at: string
}

class LocalHistoryManager {
  private readonly VIDEO_HISTORY_KEY = 'video_crawler_video_history'
  private readonly SEARCH_HISTORY_KEY = 'video_crawler_search_history'
  private readonly MAX_VIDEO_HISTORY = 100
  private readonly MAX_SEARCH_HISTORY = 50

  // 生成唯一ID
  private generateId(): string {
    return Date.now().toString(36) + Math.random().toString(36).substr(2)
  }

  // 获取当前时间字符串
  private getCurrentTime(): string {
    return new Date().toISOString()
  }

  // 视频观看历史相关方法
  addVideoHistory(
    videoId: string,
    videoTitle: string,
    videoUrl: string,
    sourceId: string,
    sourceName: string,
    watchTime: number = 0,
    progress: number = 0
  ): void {
    try {
      const histories = this.getVideoHistories()
      
      // 检查是否已存在相同视频的记录
      const existingIndex = histories.findIndex(h => h.video_id === videoId)
      if (existingIndex !== -1) {
        // 更新现有记录
        histories[existingIndex].watch_time = watchTime
        histories[existingIndex].progress = progress
        histories[existingIndex].updated_at = this.getCurrentTime()
      } else {
        // 添加新记录
        const newHistory: LocalVideoHistory = {
          id: this.generateId(),
          video_id: videoId,
          video_title: videoTitle,
          video_url: videoUrl,
          source_id: sourceId,
          source_name: sourceName,
          watch_time: watchTime,
          progress: progress,
          created_at: this.getCurrentTime(),
          updated_at: this.getCurrentTime()
        }
        
        histories.push(newHistory)
        
        // 限制记录数量
        if (histories.length > this.MAX_VIDEO_HISTORY) {
          histories.shift() // 删除最旧的记录
        }
      }
      
      // 保存到localStorage
      localStorage.setItem(this.VIDEO_HISTORY_KEY, JSON.stringify(histories))
      console.log('本地视频历史记录已保存:', histories.length, '条记录')
    } catch (error) {
      console.error('保存本地视频历史记录失败:', error)
    }
  }

  getVideoHistories(): LocalVideoHistory[] {
    try {
      const data = localStorage.getItem(this.VIDEO_HISTORY_KEY)
      if (!data) return []
      
      const histories = JSON.parse(data) as LocalVideoHistory[]
      // 按更新时间倒序排序
      return histories.sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
    } catch (error) {
      console.error('读取本地视频历史记录失败:', error)
      return []
    }
  }

  updateVideoProgress(videoId: string, watchTime: number, progress: number): void {
    try {
      const histories = this.getVideoHistories()
      const index = histories.findIndex(h => h.video_id === videoId)
      if (index !== -1) {
        histories[index].watch_time = watchTime
        histories[index].progress = progress
        histories[index].updated_at = this.getCurrentTime()
        localStorage.setItem(this.VIDEO_HISTORY_KEY, JSON.stringify(histories))
      }
    } catch (error) {
      console.error('更新视频进度失败:', error)
    }
  }

  deleteVideoHistory(videoId: string): void {
    try {
      const histories = this.getVideoHistories()
      const filtered = histories.filter(h => h.video_id !== videoId)
      localStorage.setItem(this.VIDEO_HISTORY_KEY, JSON.stringify(filtered))
      console.log('删除视频历史记录:', videoId)
    } catch (error) {
      console.error('删除视频历史记录失败:', error)
    }
  }

  clearVideoHistories(): void {
    try {
      localStorage.removeItem(this.VIDEO_HISTORY_KEY)
      console.log('清空所有视频历史记录')
    } catch (error) {
      console.error('清空视频历史记录失败:', error)
    }
  }

  // 搜索历史相关方法
  addSearchHistory(keyword: string, sourceId: string): void {
    try {
      const histories = this.getSearchHistories()
      
      // 检查是否已存在相同关键词的记录
      const existingIndex = histories.findIndex(h => h.keyword === keyword && h.source_id === sourceId)
      if (existingIndex !== -1) {
        // 更新现有记录的时间
        histories[existingIndex].created_at = this.getCurrentTime()
      } else {
        // 添加新记录
        const newHistory: LocalSearchHistory = {
          id: this.generateId(),
          keyword: keyword,
          source_id: sourceId,
          created_at: this.getCurrentTime()
        }
        
        histories.push(newHistory)
        
        // 限制记录数量
        if (histories.length > this.MAX_SEARCH_HISTORY) {
          histories.shift() // 删除最旧的记录
        }
      }
      
      // 保存到localStorage
      localStorage.setItem(this.SEARCH_HISTORY_KEY, JSON.stringify(histories))
      console.log('本地搜索历史记录已保存:', histories.length, '条记录')
    } catch (error) {
      console.error('保存本地搜索历史记录失败:', error)
    }
  }

  getSearchHistories(): LocalSearchHistory[] {
    try {
      const data = localStorage.getItem(this.SEARCH_HISTORY_KEY)
      if (!data) return []
      
      const histories = JSON.parse(data) as LocalSearchHistory[]
      // 按创建时间倒序排序
      return histories.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    } catch (error) {
      console.error('读取本地搜索历史记录失败:', error)
      return []
    }
  }

  deleteSearchHistory(keyword: string, sourceId: string): void {
    try {
      const histories = this.getSearchHistories()
      const filtered = histories.filter(h => !(h.keyword === keyword && h.source_id === sourceId))
      localStorage.setItem(this.SEARCH_HISTORY_KEY, JSON.stringify(filtered))
      console.log('删除搜索历史记录:', keyword, sourceId)
    } catch (error) {
      console.error('删除搜索历史记录失败:', error)
    }
  }

  clearSearchHistories(): void {
    try {
      localStorage.removeItem(this.SEARCH_HISTORY_KEY)
      console.log('清空所有搜索历史记录')
    } catch (error) {
      console.error('清空搜索历史记录失败:', error)
    }
  }

  // 清空所有历史记录
  clearAllHistories(): void {
    this.clearVideoHistories()
    this.clearSearchHistories()
  }
}

// 导出单例实例
export const localHistoryManager = new LocalHistoryManager()
