// Android WebView JS Bridge interfaces
interface AndroidKV {
  setItem(key: string, value: string): void
  getItem(key: string): string | null
  removeItem(key: string): void
  hasKey(key: string): boolean
}

interface AndroidSession {
  save(origin: string, json: string): void
  load(origin: string): string
}

interface AndroidDownload {
  downloadFile(url: string, filename: string): void
  downloadBlobFile(base64Data: string, filename: string): void
  saveBlobData(jsonData: string, filename: string): void
  downloadBlobFileWithPicker(base64Data: string, filename: string): void
  downloadFileWithPicker(url: string, filename: string): void
}

declare global {
  interface Window {
    AndroidKV?: AndroidKV
    AndroidSession?: AndroidSession
    AndroidDownload?: AndroidDownload
  }
}

export {}
