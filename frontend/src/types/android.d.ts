// Android WebView JavaScript接口类型定义

declare global {
  interface Window {
    // Android WebView接口
    AndroidDownload?: {
      downloadBlobFile(base64Data: string, filename: string): void
      downloadBlobFileWithPicker(base64Data: string, filename: string): void
      downloadFile(url: string, filename: string): void
      downloadFileWithPicker(url: string, filename: string): void
      saveBlobData(jsonData: string, filename: string): void
    }
    
    // Android KV存储接口
    AndroidKV?: {
      setItem(key: string, value: string): void
      getItem(key: string): string | null
      removeItem(key: string): void
      hasKey(key: string): boolean
    }
    
    // Android Session接口
    AndroidSession?: {
      save(origin: string, json: string): void
      load(origin: string): string | null
    }
  }
}

export {}
